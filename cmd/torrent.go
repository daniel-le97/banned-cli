/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// torrentCmd represents the torrent command
var torrentCmd = &cobra.Command{
	Use:   "torrent",
	Short: "Create torrents from downloaded videos",
	Long: fmt.Sprintf(`Create .torrent files from downloaded .mp4 videos.

%s can convert your downloaded videos to torrent files using various methods:
- External tools (mktorrent, transmission-create)
- Custom torrent creation with configurable trackers
- Batch processing of multiple files

Examples:
  %s torrent create video.mp4              # Create torrent for single file
  %s torrent create *.mp4                  # Create torrents for all mp4 files  
  %s torrent create --tracker url video.mp4 # Create with custom tracker
  %s torrent batch /path/to/videos/       # Batch create from directory`,
		AppName, AppName, AppName, AppName, AppName),
}

// torrentCreateCmd creates torrent files
var torrentCreateCmd = &cobra.Command{
	Use:   "create [file...]",
	Short: "Create torrent file(s) from video file(s)",
	Long: `Create .torrent files from one or more video files.
	
Supports various creation methods and allows custom tracker configuration.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tracker, _ := cmd.Flags().GetString("tracker")
		outputDir, _ := cmd.Flags().GetString("output")
		pieceLength, _ := cmd.Flags().GetInt("piece-length")
		comment, _ := cmd.Flags().GetString("comment")
		private, _ := cmd.Flags().GetBool("private")

		for _, file := range args {
			if err := createTorrent(file, &TorrentConfig{
				Tracker:     tracker,
				OutputDir:   outputDir,
				PieceLength: pieceLength,
				Comment:     comment,
				Private:     private,
			}); err != nil {
				fmt.Printf("❌ Failed to create torrent for %s: %v\n", file, err)
			} else {
				fmt.Printf("✅ Created torrent for %s\n", file)
			}
		}
	},
}

// torrentBatchCmd creates torrents for all files in a directory
var torrentBatchCmd = &cobra.Command{
	Use:   "batch [directory]",
	Short: "Create torrents for all videos in a directory",
	Long:  `Recursively find all .mp4 files in a directory and create torrent files for them.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]
		tracker, _ := cmd.Flags().GetString("tracker")
		outputDir, _ := cmd.Flags().GetString("output")
		pieceLength, _ := cmd.Flags().GetInt("piece-length")
		private, _ := cmd.Flags().GetBool("private")

		files, err := findVideoFiles(directory)
		if err != nil {
			fmt.Printf("❌ Failed to find video files: %v\n", err)
			return
		}

		fmt.Printf("🔍 Found %d video files\n", len(files))

		successful := 0
		for _, file := range files {
			if err := createTorrent(file, &TorrentConfig{
				Tracker:     tracker,
				OutputDir:   outputDir,
				PieceLength: pieceLength,
				Private:     private,
			}); err != nil {
				fmt.Printf("❌ Failed: %s (%v)\n", filepath.Base(file), err)
			} else {
				fmt.Printf("✅ Created: %s\n", filepath.Base(file))
				successful++
			}
		}

		fmt.Printf("\n🎉 Successfully created %d/%d torrents\n", successful, len(files))
	},
}

// TorrentConfig holds configuration for torrent creation
type TorrentConfig struct {
	Tracker     string
	OutputDir   string
	PieceLength int
	Comment     string
	Private     bool
}

// createTorrent creates a torrent file for the given video file
func createTorrent(videoFile string, config *TorrentConfig) error {
	// Check if file exists and is a video file
	if err := validateVideoFile(videoFile); err != nil {
		return err
	}

	// Determine output directory
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(videoFile)
	}

	// Generate output torrent filename
	baseName := strings.TrimSuffix(filepath.Base(videoFile), filepath.Ext(videoFile))
	torrentFile := filepath.Join(outputDir, baseName+".torrent")

	// Use mktorrent as primary method since it's guaranteed to be available
	return createTorrentWithMktorrent(videoFile, torrentFile, config)
}

// createTorrentWithMktorrent uses the mktorrent command-line tool
func createTorrentWithMktorrent(videoFile, torrentFile string, config *TorrentConfig) error {
	args := []string{
		"-o", torrentFile, // Output file
		"-v", // Verbose output
	}

	// Add tracker - use default if none specified
	tracker := config.Tracker
	if tracker == "" {
		tracker = DefaultTrackerURL
	}
	args = append(args, "-a", tracker)

	// Add piece length - mktorrent expects exponent (2^n), not actual KB
	// Convert KB to exponent: 256KB = 2^18, 512KB = 2^19, etc.
	pieceLength := config.PieceLength
	if pieceLength <= 0 {
		pieceLength = DefaultPieceLength // 256 KB
	}

	// Calculate exponent: log2(pieceLength * 1024)
	var exponent int
	switch {
	case pieceLength <= 32:
		exponent = 15 // 32 KB
	case pieceLength <= 64:
		exponent = 16 // 64 KB
	case pieceLength <= 128:
		exponent = 17 // 128 KB
	case pieceLength <= 256:
		exponent = 18 // 256 KB
	case pieceLength <= 512:
		exponent = 19 // 512 KB
	case pieceLength <= 1024:
		exponent = 20 // 1 MB
	case pieceLength <= 2048:
		exponent = 21 // 2 MB
	default:
		exponent = 18 // Default to 256 KB
	}

	args = append(args, "-l", fmt.Sprintf("%d", exponent))

	// Add comment
	comment := config.Comment
	if comment == "" {
		comment = fmt.Sprintf("Created by %s - banned.video downloader", AppName)
	}
	args = append(args, "-c", comment)

	// Add private flag if requested
	if config.Private {
		args = append(args, "-p")
	}

	// Add the video file
	args = append(args, videoFile)

	fmt.Printf("🔨 Creating torrent with mktorrent...\n")
	fmt.Printf("📁 Input: %s\n", filepath.Base(videoFile))
	fmt.Printf("💾 Output: %s\n", filepath.Base(torrentFile))
	fmt.Printf("🌐 Tracker: %s\n", tracker)

	cmd := exec.Command("mktorrent", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mktorrent failed: %w\nOutput: %s", err, output)
	}

	// Log success with file info
	if info, err := os.Stat(torrentFile); err == nil {
		fmt.Printf("✅ Torrent created successfully (%.2f KB)\n", float64(info.Size())/1024)

		// Update database if the video file is tracked
		if err := updateTorrentStatus(videoFile, torrentFile); err != nil {
			fmt.Printf("⚠️  Warning: Could not update database: %v\n", err)
		}
	}

	return nil
}

// validateVideoFile checks if the file exists and is a video file
func validateVideoFile(filename string) error {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filename)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", filename)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("not a recognized video file: %s (extension: %s)", filename, ext)
}

// findVideoFiles recursively finds all video files in a directory
func findVideoFiles(directory string) ([]string, error) {
	var videoFiles []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if validateVideoFile(path) == nil {
				videoFiles = append(videoFiles, path)
			}
		}

		return nil
	})

	return videoFiles, err
}

// updateTorrentStatus updates the database to mark that a torrent was created
func updateTorrentStatus(videoFile, torrentFile string) error {
	// Try to find the video file in the downloads table and update it
	err := UpdateDownloadTorrentInfo(videoFile, torrentFile)
	if err != nil {
		// This is not critical, just log it
		fmt.Printf("⚠️ Could not update download record: %v\n", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(torrentCmd)
	torrentCmd.AddCommand(torrentCreateCmd)
	torrentCmd.AddCommand(torrentBatchCmd)

	// Flags for create command
	torrentCreateCmd.Flags().StringP("tracker", "t", "", fmt.Sprintf("Tracker announce URL (default: %s)", DefaultTrackerURL))
	torrentCreateCmd.Flags().StringP("output", "o", "", "Output directory for torrent files (default: same as video)")
	torrentCreateCmd.Flags().IntP("piece-length", "l", 0, fmt.Sprintf("Piece length in KB (default: %d)", DefaultPieceLength))
	torrentCreateCmd.Flags().StringP("comment", "c", "", "Torrent comment (default: auto-generated)")
	torrentCreateCmd.Flags().BoolP("private", "p", true, "Create private torrent (recommended for banned.video content)")

	// Flags for batch command
	torrentBatchCmd.Flags().StringP("tracker", "t", "", fmt.Sprintf("Tracker announce URL (default: %s)", DefaultTrackerURL))
	torrentBatchCmd.Flags().StringP("output", "o", "", "Output directory for torrent files (default: same as video)")
	torrentBatchCmd.Flags().IntP("piece-length", "l", 0, fmt.Sprintf("Piece length in KB (default: %d)", DefaultPieceLength))
	torrentBatchCmd.Flags().BoolP("private", "p", true, "Create private torrents (recommended for banned.video content)")
}
