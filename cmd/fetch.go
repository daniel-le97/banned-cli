/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from banned.video API",
	Long: `Fetch channels and videos from banned.video API and store in database.

This command allows you to:
- Fetch all available channels
- Fetch channel details and videos
- Cache data in local database for offline access`,
}

// fetchChannelsCmd fetches all channels
var fetchChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Fetch all channels from API and replace local data",
	Long: `Fetch all available channels from banned.video API and completely replace the local database with fresh data.

This will:
- Fetch all current channels from the API
- Replace all existing channel data with fresh information
- Update video counts and channel metadata
- Preserve existing video records but update channel references`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔄 Fetching all channels from banned.video...\n")
		fmt.Printf("⚠️  This will replace all existing channel data with fresh API data.\n")

		client := NewClient("https://api.banned.video/graphql")

		// Get count before fetching
		_, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Could not connect to database: %v\n", err)
			return
		}

		oldCount, err := CountChannels()
		if err != nil {
			fmt.Printf("⚠️ Could not get current channel count: %v\n", err)
			oldCount = 0
		}
		fmt.Printf("📋 Current channels in database: %d\n", oldCount)

		if err := client.FetchAllChannels(); err != nil {
			fmt.Printf("❌ Failed to fetch channels: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully fetched and replaced all channel data!\n")

		// Show updated statistics
		fmt.Println("\n📊 Updated database statistics:")
		newCount, err := CountChannels()
		if err != nil {
			fmt.Printf("⚠️ Could not get updated channel count: %v\n", err)
		} else {
			fmt.Printf("📋 Channels in database: %d\n", newCount)
		}

		if newCount > oldCount {
			fmt.Printf("🆕 Added %d new channels\n", newCount-oldCount)
		} else if newCount < oldCount {
			fmt.Printf("🗑️  Removed %d obsolete channels\n", oldCount-newCount)
		} else {
			fmt.Printf("� Updated %d existing channels\n", newCount)
		}
	},
}

// fetchVideosCmd fetches videos from a specific channel
var fetchVideosCmd = &cobra.Command{
	Use:   "videos <channel-id>",
	Short: "Fetch videos from a specific channel",
	Long: `Fetch detailed information about a specific channel and its videos from banned.video API.

By default, fetches the first 50 videos. Use --all flag to recursively fetch ALL videos.

Examples:
  banned fetch videos 5b885d33e6646a0015a6fa2d           # Fetch channel + first 50 videos
  banned fetch videos 5b885d33e6646a0015a6fa2d --all     # Fetch channel + ALL videos recursively`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channelID := args[0]
		fmt.Printf("🔄 Fetching channel '%s' from banned.video...\n", channelID)

		client := NewClient("https://api.banned.video/graphql")

		// Fetch channel details
		if err := client.FetchChannelData(channelID); err != nil {
			fmt.Printf("❌ Failed to fetch channel details: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully fetched and stored channel details!\n")

		// Get channel details to check total video count
		channel, err := GetChannelByID(channelID)
		if err != nil {
			fmt.Printf("⚠️ Failed to get channel details from database: %v\n", err)
			return
		}

		// Count videos currently in database for this channel
		currentVideoCount, err := CountVideosForChannel(channelID)
		if err != nil {
			fmt.Printf("⚠️ Failed to count videos in database: %v\n", err)
			return
		}

		fmt.Printf("📊 Channel '%s' has %.0f total videos (API) vs %d in database\n",
			channel.Title, channel.TotalVideos, currentVideoCount)

		// Check if we need to fetch videos
		fetchAll, _ := cmd.Flags().GetBool("all")
		var shouldFetch bool
		var fetchReason string

		if fetchAll {
			// For --all flag, fetch if we have fewer videos than the channel's total
			totalVideos := int(channel.TotalVideos)
			if currentVideoCount < totalVideos {
				shouldFetch = true
				fetchReason = fmt.Sprintf("missing %d videos", totalVideos-currentVideoCount)
			} else {
				fetchReason = "all videos already in database"
			}
		} else {
			// For regular fetch (50 videos), always fetch to get latest
			shouldFetch = true
			fetchReason = "fetching latest 50 videos"
		}

		if !shouldFetch {
			fmt.Printf("✅ Skipping video fetch: %s\n", fetchReason)
			return
		}

		fmt.Printf("🔄 %s - proceeding with fetch...\n", fetchReason)

		// Fetch videos for this channel
		var videos []Video

		if fetchAll {
			// Fetch all videos recursively
			videos, err = client.FetchAllVideos(channelID)
		} else {
			// Fetch only first page (50 videos)
			fmt.Printf("🔄 Fetching videos for channel '%s'...\n", channelID)
			videos, err = client.FetchVideos(channelID, 0, 50)
		}

		if err != nil {
			fmt.Printf("❌ Failed to fetch videos: %v\n", err)
			return
		}

		if !fetchAll {
			fmt.Printf("✅ Successfully fetched and stored %d videos!\n", len(videos))
		}

		// Fetch file sizes for videos in background (only for videos without file sizes)
		if len(videos) > 0 {
			go func() {
				var videosToFetch []Video
				for _, v := range videos {
					if v.FileSize == 0 && v.DirectURL != "" {
						videosToFetch = append(videosToFetch, v)
					}
				}
				if len(videosToFetch) > 0 {
					fmt.Printf("📏 Fetching file sizes for %d videos in background...\n", len(videosToFetch))
					fetchVideoFileSizes(videosToFetch)
					fmt.Printf("✅ File sizes updated!\n")
				}
			}()
		}

		// Show updated statistics
		fmt.Println("\n📊 Updated database statistics:")
		_, err = GetDB()
		if err != nil {
			fmt.Printf("⚠️  Could not check database: %v\n", err)
			return
		}

		channelCount, err := CountChannels()
		if err != nil {
			fmt.Printf("⚠️  Could not count channels: %v\n", err)
		} else {
			fmt.Printf("📋 Channels in database: %d\n", channelCount)
		}

		videoCount, err := CountVideos()
		if err != nil {
			fmt.Printf("⚠️  Could not count videos: %v\n", err)
		} else {
			fmt.Printf("📋 Videos in database: %d\n", videoCount)
		}
	},
}

// listChannelsCmd lists channels from database
var listChannelsCmd = &cobra.Command{
	Use:   "list-channels",
	Short: "Browse channels with interactive TUI",
	Long: `Browse all channels stored in the local database using an interactive TUI.

Features:
- Navigate with arrow keys or vim keys (j/k)
- Press '/' or Ctrl+F to search/filter channels
- Type in search box to filter by title, description, or ID
- Press Enter to select a channel or apply search filter
- Press Esc to exit search mode or quit application
- View channel info, social links, and video counts
- Fetch latest videos or browse existing ones
- Press 'q' or Ctrl+C to exit`,
	Run: func(cmd *cobra.Command, args []string) {
		channels, err := GetAllChannelsFromDB()
		if err != nil {
			fmt.Printf("❌ Failed to get channels from database: %v\n", err)
			return
		}

		if len(channels) == 0 {
			fmt.Printf("📭 No channels found in database.\n")
			fmt.Printf("💡 Use 'banned fetch channels' to fetch from API.\n")
			return
		}

		// Launch Bubble Tea TUI for channel browsing
		if err := RunChannelBrowserTUI(channels); err != nil {
			fmt.Printf("❌ Failed to start channel browser: %v\n", err)
			return
		}
	},
}

// syncFileSizesCmd fetches file sizes for videos that don't have them
var fetchFileSizesCmd = &cobra.Command{
	Use:   "file-sizes [channel-id]",
	Short: "Fetch file sizes for videos missing size data",
	Long: `
	Fetch file sizes for videos that don't have file size data in the database.

This command will:
- Find all videos with missing file sizes (file_size = 0 or NULL)  
- Optionally filter by channel ID
- Fetch file sizes concurrently with rate limiting
- Update the database with the results

Examples:
  banned fetch file-sizes                           # Update all videos missing file sizes
  banned fetch file-sizes 5b885d33e6646a0015a6fa2d  # Update specific channels videos`,

	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var channelID string
		if len(args) > 0 {
			channelID = args[0]
		}

		_, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Could not connect to database: %v\n", err)
			return
		}

		// Get videos without file sizes
		if channelID != "" {
			fmt.Printf("🔍 Finding videos without file sizes for channel %s...\n", channelID)
		} else {
			fmt.Printf("🔍 Finding all videos without file sizes...\n")
		}

		videos, err := GetVideosWithoutFileSize(channelID)
		if err != nil {
			fmt.Printf("❌ Failed to get videos: %v\n", err)
			return
		}

		if len(videos) == 0 {
			fmt.Printf("✅ All videos already have file sizes!\n")
			return
		}

		// Use fast worker pool for maximum speed
		fmt.Printf("📏 Found %d videos without file sizes. Starting maximum speed fetch...\n", len(videos))

		var urls []string
		for _, v := range videos {
			if v.DirectURL != "" {
				urls = append(urls, v.DirectURL)
			}
		}

		// Process with worker pool
		processFileSizesWithWorkerPool(urls)
		fmt.Printf("✅ File size fetching completed!\n")
	},
}

// processFileSizesWithWorkerPool uses a worker pool for maximum concurrency and speed
func processFileSizesWithWorkerPool(urls []string) {
	const numWorkers = 100 // High concurrency for speed

	totalURLs := len(urls)
	if totalURLs == 0 {
		return
	}

	// Channels for work distribution and progress tracking
	urlChan := make(chan string, totalURLs)
	resultsChan := make(chan result, totalURLs)

	// Statistics tracking
	var (
		completed    int64
		failed       int64
		timeouts     int64
		networkErrs  int64
		databaseErrs int64
		httpErrs     int64
	)

	startTime := time.Now()

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for url := range urlChan {
				result := processURL(url)
				resultsChan <- result
			}
		}(i)
	}

	// Send URLs to workers
	go func() {
		for _, url := range urls {
			urlChan <- url
		}
		close(urlChan)
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Progress reporting
	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()

	processedCount := 0

	// Collect results and show progress
	for {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				// All results processed
				goto done
			}

			processedCount++

			// Update statistics based on result
			if result.Success {
				atomic.AddInt64(&completed, 1)
			} else {
				atomic.AddInt64(&failed, 1)
				switch result.ErrorType {
				case "timeout":
					atomic.AddInt64(&timeouts, 1)
				case "network":
					atomic.AddInt64(&networkErrs, 1)
				case "database":
					atomic.AddInt64(&databaseErrs, 1)
				case "http":
					atomic.AddInt64(&httpErrs, 1)
				}

				// Print error details immediately
				fmt.Printf("❌ %s: %s\n", result.ErrorType, result.URL)
			}

		case <-progressTicker.C:
			// Show progress every 2 seconds
			elapsed := time.Since(startTime)
			rate := float64(processedCount) / elapsed.Seconds()
			eta := time.Duration(float64(totalURLs-processedCount)/rate) * time.Second

			fmt.Printf("\r📏 Progress: %d/%d (%.1f%%) | ⚡ %.1f/sec | ⏱️ ETA: %v | ✅ %d | ❌ %d",
				processedCount, totalURLs,
				float64(processedCount)/float64(totalURLs)*100,
				rate, eta.Round(time.Second),
				atomic.LoadInt64(&completed), atomic.LoadInt64(&failed))
		}
	}

done:
	elapsed := time.Since(startTime)
	completedFinal := atomic.LoadInt64(&completed)
	failedFinal := atomic.LoadInt64(&failed)

	fmt.Printf("\n\n🎉 Completed in %v!\n", elapsed.Round(time.Second))
	fmt.Printf("✅ Successful: %d\n", completedFinal)
	fmt.Printf("❌ Failed: %d\n", failedFinal)

	if failedFinal > 0 {
		fmt.Printf("\nError breakdown:\n")
		if t := atomic.LoadInt64(&timeouts); t > 0 {
			fmt.Printf("  ⏱️ Timeouts: %d\n", t)
		}
		if n := atomic.LoadInt64(&networkErrs); n > 0 {
			fmt.Printf("  🌐 Network errors: %d\n", n)
		}
		if h := atomic.LoadInt64(&httpErrs); h > 0 {
			fmt.Printf("  🌍 HTTP errors: %d\n", h)
		}
		if d := atomic.LoadInt64(&databaseErrs); d > 0 {
			fmt.Printf("  💾 Database errors: %d\n", d)
		}
	}
}

// result represents the outcome of processing a URL
type result struct {
	URL       string
	Success   bool
	ErrorType string
	Error     error
}

// processURL handles fetching and updating file size for a single URL
func processURL(url string) result {
	size, err := getFileSize(url)
	if err != nil {
		// Quick retry on failure
		time.Sleep(50 * time.Millisecond)
		size, err = getFileSize(url)
		if err != nil {
			// Classify error type
			errorStr := err.Error()
			errorType := "network"
			if strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "deadline") {
				errorType = "timeout"
			} else if strings.Contains(errorStr, "HTTP") {
				errorType = "http"
			}

			return result{
				URL:       url,
				Success:   false,
				ErrorType: errorType,
				Error:     err,
			}
		}
	}

	if err == nil && size > 0 {
		if updateErr := UpdateVideoFileSizeByURL(url, size); updateErr != nil {
			return result{
				URL:       url,
				Success:   false,
				ErrorType: "database",
				Error:     updateErr,
			}
		}
	}

	return result{
		URL:     url,
		Success: true,
	}
}

// fetchVideoFileSizesConcurrent fetches file sizes for videos with high concurrency and batching
func fetchVideoFileSizesConcurrent(videos []Video) {
	const (
		maxConcurrency = 50  // Higher concurrency for speed
		batchSize      = 500 // Process in batches for progress updates
		maxRetries     = 2   // Retry failed requests
	)

	totalVideos := 0
	for _, video := range videos {
		if video.DirectURL != "" {
			totalVideos++
		}
	}

	if totalVideos == 0 {
		return
	}

	processed := 0
	failed := 0
	startTime := time.Now()

	// Process videos in batches
	for i := 0; i < len(videos); i += batchSize {
		end := i + batchSize
		if end > len(videos) {
			end = len(videos)
		}
		batch := videos[i:end]

		// Process current batch
		batchProcessed, batchFailed := processBatch(batch, maxConcurrency, maxRetries)
		processed += batchProcessed
		failed += batchFailed

		// Progress update
		elapsed := time.Since(startTime)
		rate := float64(processed) / elapsed.Seconds()
		eta := time.Duration(float64(totalVideos-processed)/rate) * time.Second

		fmt.Printf("\r📏 Progress: %d/%d (%.1f%%) | Rate: %.1f/sec | Failed: %d | ETA: %v",
			processed, totalVideos, float64(processed)/float64(totalVideos)*100,
			rate, failed, eta.Round(time.Second))
	}
	fmt.Printf("\n")
}

// processBatch processes a batch of videos concurrently
func processBatch(videos []Video, maxConcurrency, maxRetries int) (processed, failed int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)
	var processedCount, failedCount int64

	for _, video := range videos {
		if video.DirectURL == "" {
			continue
		}

		wg.Add(1)
		go func(v Video) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Try to get file size with retries
			var size int64
			var err error
			for retry := 0; retry <= maxRetries; retry++ {
				size, err = getFileSize(v.DirectURL)
				if err == nil {
					break
				}
				if retry < maxRetries {
					time.Sleep(time.Duration(retry+1) * 100 * time.Millisecond)
				}
			}

			if err == nil && size > 0 {
				if updateErr := UpdateVideoFileSizeByURL(v.DirectURL, size); updateErr == nil {
					atomic.AddInt64(&processedCount, 1)
				} else {
					atomic.AddInt64(&failedCount, 1)
				}
			} else {
				atomic.AddInt64(&failedCount, 1)
			}
		}(video)
	}

	wg.Wait()
	return int(atomic.LoadInt64(&processedCount)), int(atomic.LoadInt64(&failedCount))
}

// fetchVideoFileSizes fetches file sizes for videos concurrently (legacy function)
func fetchVideoFileSizes(videos []Video) {
	fetchVideoFileSizesConcurrent(videos)
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(listChannelsCmd)
	fetchCmd.AddCommand(fetchChannelsCmd)
	fetchCmd.AddCommand(fetchVideosCmd)
	fetchCmd.AddCommand(fetchFileSizesCmd)

	// Add flags for fetch channel command
	fetchVideosCmd.Flags().BoolP("all", "a", false, "Fetch ALL videos recursively (may take time for channels with many videos)")
}
