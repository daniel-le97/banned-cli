/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

// VideoInfo holds video information with file size
type VideoInfo struct {
	ID        string
	Title     string
	DirectURL string
	FileSize  int64
	Error     error
}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [channel-id]",
	Short: "Get all direct URLs and file sizes for videos from a given channel",
	Long: `Retrieves all videos from the database for a specified channel ID 
and fetches the file size for each video's direct URL.

Example:
  banned download 5b885d33e6646a0015a6fa2d
  banned download 5b885d33e6646a0015a6fa2d --stats`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channelID := args[0]
		showStats, _ := cmd.Flags().GetBool("stats")

		if err := getChannelVideoSizes(channelID, showStats); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

// getChannelVideoSizes retrieves all videos for a channel and gets their file sizes
func getChannelVideoSizes(channelID string, showStats bool) error {
	fmt.Printf("📺 Fetching videos for channel: %s\n\n", channelID)

	// Get all videos from database (using a large limit to get all)
	videos, err := GetChannelVideos(channelID, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to get videos from database: %w", err)
	}

	if len(videos) == 0 {
		fmt.Println("No videos found for this channel ID")
		return nil
	}

	// Count how many videos already have file sizes cached
	var cachedSizes, needsFetching int
	for _, video := range videos {
		if video.FileSize > 0 {
			cachedSizes++
		} else if video.DirectURL != "" {
			needsFetching++
		}
	}

	fmt.Printf("Found %d videos (%d cached, %d need fetching)...\n\n", len(videos), cachedSizes, needsFetching)

	// Stats tracking
	var cachedCount, fetchedCount int64

	// Create channels for concurrent processing
	videoInfoChan := make(chan VideoInfo, len(videos))
	var wg sync.WaitGroup

	// Limit concurrent requests to avoid overwhelming servers
	semaphore := make(chan struct{}, 10)

	// Process each video concurrently
	for _, video := range videos {
		if video.DirectURL == "" {
			continue // Skip videos without direct URLs
		}

		wg.Add(1)
		go func(v Video) {
			defer wg.Done()

			info := VideoInfo{
				ID:        v.ID,
				Title:     v.Title,
				DirectURL: v.DirectURL,
			}

			// Check if we already have file size from database
			if v.FileSize > 0 {
				info.FileSize = v.FileSize
				info.Error = nil
				if showStats {
					cachedCount++
				}
				videoInfoChan <- info
				return
			}

			// Acquire semaphore for HTTP request
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Get file size from HTTP request
			size, err := getFileSize(v.DirectURL)
			info.FileSize = size
			info.Error = err

			// Update database with file size if successful
			if err == nil && size > 0 {
				if showStats {
					fetchedCount++
				}
				if updateErr := UpdateVideoFileSize(v.ID, size); updateErr != nil {
					// Don't fail the whole operation, just log the error
					fmt.Printf("Warning: Failed to update file size in database for video %s: %v\n", v.ID, updateErr)
				}
			}

			videoInfoChan <- info
		}(video)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(videoInfoChan)
	}()

	// Collect and display results
	var totalSize int64
	successCount := 0

	fmt.Printf("%-50s %-15s %s\n", "Title", "Size", "URL")
	fmt.Printf("%-50s %-15s %s\n", "-----", "----", "---")

	for info := range videoInfoChan {
		titleTrunc := info.Title
		if len(titleTrunc) > 47 {
			titleTrunc = titleTrunc[:47] + "..."
		}

		if info.Error != nil {
			fmt.Printf("%-50s %-15s %s (Error: %v)\n", titleTrunc, "Unknown", info.DirectURL, info.Error)
		} else {
			sizeStr := formatFileSize(info.FileSize)
			fmt.Printf("%-50s %-15s %s\n", titleTrunc, sizeStr, info.DirectURL)
			totalSize += info.FileSize
			successCount++
		}
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Videos processed: %d\n", len(videos))
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Total size: %s\n", formatFileSize(totalSize))

	if showStats {
		cached := atomic.LoadInt64(&cachedCount)
		fetched := atomic.LoadInt64(&fetchedCount)
		fmt.Printf("\n📈 Performance Stats:\n")
		fmt.Printf("   Cached from DB: %d\n", cached)
		fmt.Printf("   Fetched via HTTP: %d\n", fetched)
		if cached+fetched > 0 {
			cacheRatio := float64(cached) / float64(cached+fetched) * 100
			fmt.Printf("   Cache hit ratio: %.1f%%\n", cacheRatio)
		}
	}

	return nil
}

// getFileSize gets the content length of a URL using HEAD request
func getFileSize(url string) (int64, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("no content-length header")
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid content-length: %s", contentLength)
	}

	return size, nil
}

// formatFileSize formats bytes into human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Add flags
	downloadCmd.Flags().BoolP("stats", "s", false, "Show performance statistics (cached vs fetched file sizes)")
}
