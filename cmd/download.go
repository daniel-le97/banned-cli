/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [channel-id]",
	Short: "Download all videos for a channel",
	Long: `Downloads all videos for a specified channel ID to the configured download folder.
Shows total file size and prompts for confirmation before downloading.

Example:
  banned download 5b885d33e6646a0015a6fa2d
  banned download 5b885d33e6646a0015a6fa2d --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channelID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if err := downloadChannelVideos(channelID, force); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

// downloadChannelVideos downloads all videos for a channel after showing total size and prompting user
func downloadChannelVideos(channelID string, force bool) error {
	fmt.Printf("📺 Preparing to download videos for channel: %s\n\n", channelID)

	// Get all videos from database
	videos, err := GetChannelVideos(channelID, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to get videos from database: %w", err)
	}

	if len(videos) == 0 {
		fmt.Println("No videos found for this channel ID")
		return nil
	}

	// Filter videos that have direct URLs and file sizes
	var downloadableVideos []Video
	var totalSize int64
	var videosWithoutURL, videosWithoutSize int

	for _, video := range videos {
		if video.DirectURL == "" {
			videosWithoutURL++
			continue
		}
		if video.FileSize <= 0 {
			videosWithoutSize++
			continue
		}
		downloadableVideos = append(downloadableVideos, video)
		totalSize += video.FileSize
	}

	fmt.Printf("📊 Download Summary:\n")
	fmt.Printf("   Total videos in channel: %d\n", len(videos))
	fmt.Printf("   Videos ready for download: %d\n", len(downloadableVideos))
	if videosWithoutURL > 0 {
		fmt.Printf("   Videos without direct URLs: %d (run 'banned fetch' first)\n", videosWithoutURL)
	}
	if videosWithoutSize > 0 {
		fmt.Printf("   Videos without file sizes: %d (run 'banned fetch file-sizes' first)\n", videosWithoutSize)
	}
	fmt.Printf("   Total download size: %s\n\n", formatFileSize(totalSize))

	if len(downloadableVideos) == 0 {
		fmt.Println("No videos ready for download. Make sure to run fetch commands first.")
		return nil
	}

	// Get download directory from config
	downloadDir, err := getDownloadDirectory()
	if err != nil {
		return fmt.Errorf("failed to get download directory: %w", err)
	}

	fmt.Printf("📁 Download directory: %s\n\n", downloadDir)

	// Prompt user for confirmation unless force flag is used
	if !force {
		fmt.Printf("Do you want to proceed with downloading %d videos (%s)? [y/N]: ", len(downloadableVideos), formatFileSize(totalSize))
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Download cancelled.")
			return nil
		}
	}

	// Start downloading
	fmt.Printf("\n🚀 Starting download of %d videos...\n\n", len(downloadableVideos))
	return downloadVideos(downloadableVideos, downloadDir)
}

// downloadVideos downloads a list of videos concurrently
func downloadVideos(videos []Video, downloadDir string) error {
	// Create download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Use semaphore to limit concurrent downloads
	semaphore := make(chan struct{}, 3) // Limit to 3 concurrent downloads
	var wg sync.WaitGroup
	var successCount, errorCount int64

	for i, video := range videos {
		wg.Add(1)
		go func(index int, v Video) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create safe filename
			filename := sanitizeFilename(v.Title) + ".mp4"
			filePath := filepath.Join(downloadDir, filename)

			fmt.Printf("[%d/%d] Downloading: %s\n", index+1, len(videos), v.Title)

			if err := downloadVideoFile(v.DirectURL, filePath); err != nil {
				fmt.Printf("❌ Failed to download %s: %v\n", v.Title, err)
				errorCount++
			} else {
				fmt.Printf("✅ Downloaded: %s\n", filename)
				successCount++
			}
		}(i, video)
	}

	wg.Wait()

	fmt.Printf("\n🎉 Download complete!\n")
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", errorCount)
	fmt.Printf("   Total: %d\n", len(videos))

	return nil
}

// downloadVideoFile downloads a video file from URL to the specified path
func downloadVideoFile(url, filepath string) error {
	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return nil // File already exists, skip
	}

	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for video downloads
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to file
	_, err = io.Copy(file, resp.Body)
	return err
}

// sanitizeFilename removes or replaces invalid characters for filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	// Limit length to avoid filesystem issues
	if len(filename) > 200 {
		filename = filename[:200]
	}
	return filename
}

// getDownloadDirectory gets the download directory from config or uses default
func getDownloadDirectory() (string, error) {
	var downloadDir string

	// Try to get from config first
	if configDir, err := GetSetting("download_dir"); err == nil && configDir != "" {
		downloadDir = configDir
	} else {
		// Default to ./downloads in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		downloadDir = filepath.Join(cwd, "downloads")
	}

	// Expand tilde (~) to home directory if present
	if strings.HasPrefix(downloadDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		downloadDir = filepath.Join(homeDir, downloadDir[2:])
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory %s: %w", downloadDir, err)
	}

	return downloadDir, nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Add flags
	downloadCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt and start downloading immediately")
}
