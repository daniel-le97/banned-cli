/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "math/rand"
	"net/http"
	"os"

	// "strings"
	"sync"
	"time"

	// "github.com/charmbracelet/bubbles/progress"
	// "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daniel-le97/banned-cli/tui"

	// "github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	// "libertyarchive.com/banned/tui"
)

// syncCmd represents the sync command for incremental updates
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync new data from banned.video API",
	Long: `Sync only new data from banned.video API based on timestamps.

This command performs incremental updates by:
- Checking the latest timestamps in your local database
- Fetching only data newer than those timestamps
- Efficiently updating your local cache without re-downloading everything

This is much faster than full fetches and reduces API load.`,
}

// syncChannelsCmd syncs only new/updated channels
var syncChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Sync new/updated channels",
	Long:  `Sync only channels that have been created or updated since your last sync.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔄 Syncing new channels from banned.video...\n")

		client := NewClient("https://api.banned.video/graphql")

		if err := client.SyncChannels(); err != nil {
			fmt.Printf("❌ Failed to sync channels: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully synced channels!\n")
		showSyncStats()
	},
}

// syncChannelCmd syncs only new videos for a specific channel
var syncChannelCmd = &cobra.Command{
	Use:   "channel <channel-id>",
	Short: "Sync new videos for a specific channel",
	Long: `Sync only new videos for a specific channel based on the latest video timestamp in your database.

This will fetch only videos published after your most recent video for this channel.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channelID := args[0]
		skipFileSizes, _ := cmd.Flags().GetBool("skip-file-sizes")
		fmt.Printf("🔄 Syncing new videos for channel '%s'...\n", channelID)

		client := NewClient("https://api.banned.video/graphql")

		if err := client.SyncChannelVideosWithOptions(channelID, !skipFileSizes); err != nil {
			fmt.Printf("❌ Failed to sync channel videos: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully synced videos for channel!\n")
		showSyncStats()
	},
}

// syncCheckCmd checks video counts between API and local database
var syncCheckCmd = &cobra.Command{
	Use:   "check [channel-id]",
	Short: "Check video counts between API and local database",
	Long: `Compare video counts between the API and your local database.

If no channel-id is provided, checks all channels in your database.
This helps identify which channels need syncing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			// Check specific channel
			channelID := args[0]
			apiCount, dbCount, needsSync, err := CheckChannelVideoCount(channelID)
			if err != nil {
				fmt.Printf("❌ Error checking channel %s: %v\n", channelID, err)
				return
			}

			fmt.Printf("📊 Channel %s:\n", channelID)
			fmt.Printf("   API videos: %d\n", apiCount)
			fmt.Printf("   Local videos: %d\n", dbCount)
			if needsSync {
				fmt.Printf("   🔄 Sync needed (%d new videos)\n", apiCount-dbCount)
			} else if apiCount == dbCount {
				fmt.Printf("   ✅ Up to date\n")
			} else {
				fmt.Printf("   ⚠️  Local has more videos than API (%d extra)\n", dbCount-apiCount)
			}
		} else {
			// Check all channels
			fmt.Printf("🔍 Checking video counts for all channels...\n\n")

			channels, err := GetAllChannels()
			if err != nil {
				fmt.Printf("❌ Failed to get channels: %v\n", err)
				return
			}

			var needsSyncCount, upToDateCount int
			for i, channel := range channels {
				apiCount, dbCount, needsSync, err := CheckChannelVideoCount(channel.ID)
				if err != nil {
					fmt.Printf("❌ Error checking %s: %v\n", channel.Title, err)
					continue
				}

				status := "✅"
				if needsSync {
					status = "🔄"
					needsSyncCount++
				} else {
					upToDateCount++
				}

				fmt.Printf("%s %s: API=%d, Local=%d\n", status, channel.Title, apiCount, dbCount)

				// Progress indicator for large lists
				if (i+1)%10 == 0 {
					fmt.Printf("   ... checked %d/%d channels\n", i+1, len(channels))
				}
			}

			fmt.Printf("\n📊 Summary:\n")
			fmt.Printf("   Up to date: %d channels\n", upToDateCount)
			fmt.Printf("   Need sync: %d channels\n", needsSyncCount)
			if needsSyncCount > 0 {
				fmt.Printf("\n💡 Run 'banned sync all' to sync all channels\n")
			}
		}
	},
}

// syncAllCmd syncs all channels and their new videos
var syncAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Sync all channels and videos incrementally",
	Long: `Perform a complete incremental sync of all channels and videos.

This will:
1. Sync any new/updated channels
2. For each channel, sync only new videos published since the last sync
3. Show detailed progress and statistics`,
	Run: func(cmd *cobra.Command, args []string) {
		skipFileSizes, _ := cmd.Flags().GetBool("skip-file-sizes")
		fmt.Printf("🔄 Starting complete incremental sync...\n")

		client := NewClient("https://api.banned.video/graphql")

		// First sync channels
		fmt.Printf("📋 Step 1: Syncing channels...\n")
		if err := client.SyncChannels(); err != nil {
			fmt.Printf("❌ Failed to sync channels: %v\n", err)
			return
		}

		// Then sync videos for all channels
		fmt.Printf("📺 Step 2: Syncing videos for all channels...\n")
		if err := client.SyncAllChannelVideosWithOptions(!skipFileSizes); err != nil {
			fmt.Printf("❌ Failed to sync channel videos: %v\n", err)
			return
		}

		fmt.Printf("✅ Complete incremental sync finished!\n")
		showSyncStats()
	},
}

// SyncChannels fetches only new/updated channels
func (c *Client) SyncChannels() error {
	// Get the latest channel update timestamp from database
	lastUpdate, err := getLatestChannelTimestamp()
	if err != nil {
		fmt.Printf("📅 No existing channels found, performing full sync...\n")
		// If no channels exist, do a full fetch
		return c.FetchAllChannels()
	}

	fmt.Printf("📅 Last channel update: %s\n", lastUpdate.Format("2006-01-02 15:04:05"))

	// TODO: Implement API call with timestamp filter
	// For now, we'll fetch all channels and filter locally
	// This should be optimized to use API filtering when available
	fmt.Printf("⚠️  API timestamp filtering not yet implemented, performing full channel sync\n")
	return c.FetchAllChannels()
}

// SyncChannelVideos fetches only new videos for a specific channel
func (c *Client) SyncChannelVideos(channelID string) error {
	return c.SyncChannelVideosWithOptions(channelID, true)
}

// SyncChannelVideosWithOptions fetches only new videos for a specific channel with options
func (c *Client) SyncChannelVideosWithOptions(channelID string, fetchFileSizes bool) error {
	// Get the latest video timestamp for this channel
	lastUpdate, err := getLatestVideoTimestamp(channelID)
	if err != nil {
		fmt.Printf("📅 No existing videos found for channel %s, performing full sync...\n", channelID)
		// If no videos exist for this channel, do a full fetch
		videos, err := c.FetchAllVideos(channelID)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Synced %d videos (full sync)\n", len(videos))
		return nil
	}

	fmt.Printf("📅 Last video update for channel %s: %s\n", channelID, lastUpdate.Format("2006-01-02 15:04:05"))

	// Fetch videos published after the last update
	newVideos, err := c.fetchVideosAfterTimestamp(channelID, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to fetch new videos: %w", err)
	}

	if len(newVideos) == 0 {
		fmt.Printf("📭 No new videos found for channel %s\n", channelID)
		return nil
	}

	// Store new videos
	if err := StoreVideosBatch(channelID, newVideos); err != nil {
		return fmt.Errorf("failed to store new videos: %w", err)
	}

	// Conditionally fetch file sizes for new videos in background
	if fetchFileSizes && len(newVideos) > 0 {
		go func() {
			fmt.Printf("📏 Fetching file sizes for %d new videos...\n", len(newVideos))
			fetchVideoFileSizes(newVideos)
			fmt.Printf("✅ File sizes updated for new videos\n")
		}()
	}

	fmt.Printf("✅ Synced %d new videos for channel %s\n", len(newVideos), channelID)
	return nil
}

// SyncAllChannelVideos syncs videos for all channels incrementally
func (c *Client) SyncAllChannelVideos() error {
	return c.SyncAllChannelVideosWithOptions(true)
}

// SyncAllChannelVideosWithOptions syncs videos for all channels incrementally with options
func (c *Client) SyncAllChannelVideosWithOptions(fetchFileSizes bool) error {
	// Get all channels from database
	channels, err := GetAllChannels()
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	if len(channels) == 0 {
		fmt.Printf("📭 No channels found in database. Run 'banned fetch channels' first.\n")
		return nil
	}

	var totalNewVideos int
	for i, channel := range channels {
		fmt.Printf("📺 Syncing channel %d/%d: %s\n", i+1, len(channels), channel.Title)

		// Get the latest video timestamp for this channel
		lastUpdate, err := getLatestVideoTimestamp(channel.ID)
		if err != nil {
			// No videos exist for this channel yet, skip for incremental sync
			fmt.Printf("   📭 No existing videos, skipping incremental sync\n")
			continue
		}

		// Fetch only new videos
		newVideos, err := c.fetchVideosAfterTimestamp(channel.ID, lastUpdate)
		if err != nil {
			fmt.Printf("   ❌ Failed to sync: %v\n", err)
			continue
		}

		if len(newVideos) == 0 {
			fmt.Printf("   📭 No new videos\n")
			continue
		}

		// Store new videos
		if err := StoreVideosBatch(channel.ID, newVideos); err != nil {
			fmt.Printf("   ❌ Failed to store videos: %v\n", err)
			continue
		}

		// Conditionally fetch file sizes for new videos in background
		if fetchFileSizes {
			go fetchVideoFileSizes(newVideos)
		}

		totalNewVideos += len(newVideos)
		fmt.Printf("   ✅ Synced %d new videos\n", len(newVideos))
	}

	fmt.Printf("📊 Total new videos synced: %d\n", totalNewVideos)
	return nil
}

// Helper functions for timestamp queries

// getLatestChannelTimestamp returns the most recent channel update timestamp
func getLatestChannelTimestamp() (time.Time, error) {
	db, err := GetDB()
	if err != nil {
		return time.Time{}, err
	}

	var timestamp string
	query := `
		SELECT COALESCE(MAX(updated_at), MAX(created_at)) 
		FROM channels 
		WHERE updated_at IS NOT NULL OR created_at IS NOT NULL
		ORDER BY COALESCE(updated_at, created_at) DESC 
		LIMIT 1`

	err = db.QueryRow(query).Scan(&timestamp)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, timestamp)
}

// getLatestVideoTimestamp returns the most recent video timestamp for a channel
func getLatestVideoTimestamp(channelID string) (time.Time, error) {
	db, err := GetDB()
	if err != nil {
		return time.Time{}, err
	}

	var timestamp string
	query := `
		SELECT COALESCE(created_at_api, updated_at, created_at)
		FROM videos 
		WHERE channel_id = ? 
		AND (created_at_api IS NOT NULL OR updated_at IS NOT NULL OR created_at IS NOT NULL)
		ORDER BY COALESCE(created_at_api, updated_at, created_at) DESC
		LIMIT 1`

	err = db.QueryRow(query, channelID).Scan(&timestamp)
	if err != nil {
		return time.Time{}, err
	}

	// Handle both RFC3339 and other timestamp formats
	if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return parsedTime, nil
	}

	// Try other common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, timestamp); err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

// fetchVideosAfterTimestamp fetches videos published after a specific timestamp
func (c *Client) fetchVideosAfterTimestamp(channelID string, after time.Time) ([]Video, error) {
	// For now, we'll fetch recent videos and filter client-side
	// This should be optimized with API filtering when available

	// Fetch recent videos (first few pages)
	allVideos := []Video{}
	offset := 0
	limit := 200
	maxPages := 5 // Limit to first few pages for incremental updates

	for page := 0; page < maxPages; page++ {
		videos, err := c.fetchVideosPage(channelID, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos page %d: %w", page, err)
		}

		if len(videos) == 0 {
			break
		}

		// Filter videos published after the timestamp
		newVideos := []Video{}
		for _, video := range videos {
			videoTime, err := parseVideoTimestamp(video)
			if err != nil {
				// If we can't parse the timestamp, include it to be safe
				newVideos = append(newVideos, video)
				continue
			}

			if videoTime.After(after) {
				newVideos = append(newVideos, video)
			} else {
				// Videos are typically sorted by date, so if we hit an old one, we can stop
				fmt.Printf("   📅 Reached videos older than %s, stopping\n",
					after.Format("2006-01-02 15:04:05"))
				allVideos = append(allVideos, newVideos...)
				return allVideos, nil
			}
		}

		allVideos = append(allVideos, newVideos...)

		// If we got fewer videos than requested, we've reached the end
		if len(videos) < limit {
			break
		}

		offset += limit
	}

	return allVideos, nil
}

// parseVideoTimestamp extracts timestamp from video for comparison
func parseVideoTimestamp(video Video) (time.Time, error) {
	// Try different timestamp fields in order of preference
	timestamps := []string{
		video.CreatedAt, // This is the main timestamp field in Video struct
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	for _, timestamp := range timestamps {
		if timestamp == "" {
			continue
		}

		for _, format := range formats {
			if parsedTime, err := time.Parse(format, timestamp); err == nil {
				return parsedTime, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("no valid timestamp found for video %s", video.ID)
}

// showSyncStats displays sync statistics
func showSyncStats() {
	fmt.Println("\n📊 Updated database statistics:")
	db, err := GetDB()
	if err != nil {
		fmt.Printf("⚠️  Could not check database: %v\n", err)
		return
	}

	var channelCount, videoCount int
	db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&channelCount)
	db.QueryRow("SELECT COUNT(*) FROM videos").Scan(&videoCount)

	fmt.Printf("📋 Channels: %d\n", channelCount)
	fmt.Printf("📺 Videos: %d\n", videoCount)

	// Show latest timestamps
	if latestChannel, err := getLatestChannelTimestamp(); err == nil {
		fmt.Printf("📅 Latest channel update: %s\n", latestChannel.Format("2006-01-02 15:04:05"))
	}

	// Show most recent video across all channels
	var latestVideo string
	query := `SELECT COALESCE(created_at_api, updated_at, created_at) FROM videos 
		ORDER BY COALESCE(created_at_api, updated_at, created_at) DESC LIMIT 1`
	if err := db.QueryRow(query).Scan(&latestVideo); err == nil {
		if t, err := time.Parse(time.RFC3339, latestVideo); err == nil {
			fmt.Printf("📅 Latest video: %s\n", t.Format("2006-01-02 15:04:05"))
		}
	}
}

// fetchVideoFileSizes fetches file sizes for videos concurrently and updates database
func fetchVideoFileSizes(videos []Video) {
	if len(videos) == 0 {
		return
	}

	var wg sync.WaitGroup
	// Limit concurrent requests to avoid overwhelming servers
	semaphore := make(chan struct{}, 10)

	for _, video := range videos {
		if video.DirectURL == "" {
			continue // Skip videos without direct URLs
		}

		wg.Add(1)
		go func(v Video) {
			defer wg.Done()

			// Acquire semaphore for HTTP request
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Get file size from HTTP request
			size, err := getFileSize(v.DirectURL)
			if err != nil {
				// Don't log every error to avoid spam, just continue
				return
			}

			// Update database with file size if successful
			if size > 0 {
				if updateErr := UpdateVideoFileSize(v.ID, size); updateErr != nil {
					// Continue silently on database update errors
				}
			}
		}(video)
	}

	wg.Wait()
}

// GetChannelVideoCount fetches the video count for a channel from the API
func (c *Client) GetChannelVideoCount(channelID string) (int, error) {
	query := `
		query GetChannel($id: String!) {
			getChannel(id: $id) {
				_id
				totalVideos
			}
		}`

	variables := map[string]interface{}{
		"id": channelID,
	}

	request := GraphQLRequest{
		OperationName: "GetChannel",
		Variables:     variables,
		Query:         query,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			GetChannel struct {
				ID          string  `json:"_id"`
				TotalVideos float64 `json:"totalVideos"`
			} `json:"getChannel"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return int(response.Data.GetChannel.TotalVideos), nil
}

// CheckChannelVideoCount compares API video count with local database count
func CheckChannelVideoCount(channelID string) (apiCount, dbCount int, needsSync bool, err error) {
	// Get count from API
	client := NewClient("https://api.banned.video/graphql")
	apiCount, err = client.GetChannelVideoCount(channelID)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to get API count: %w", err)
	}

	// Get count from local database
	db, err := GetDB()
	if err != nil {
		return apiCount, 0, false, fmt.Errorf("failed to get database: %w", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM videos WHERE channel_id = ?", channelID).Scan(&dbCount)
	if err != nil {
		return apiCount, 0, false, fmt.Errorf("failed to query database: %w", err)
	}

	needsSync = apiCount > dbCount
	return apiCount, dbCount, needsSync, nil
}

// syncFileSizesCmd fetches file sizes for videos that don't have them
var syncFileSizesCmd = &cobra.Command{
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
  banned sync file-sizes                           # Update all videos missing file sizes
  banned sync file-sizes 5b885d33e6646a0015a6fa2d  # Update specific channels videos`,

	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var channelID string
		if len(args) > 0 {
			channelID = args[0]
		}

		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Could not connect to database: %v\n", err)
			return
		}

		// Build query to find videos without file sizes
		var query string
		var queryArgs []interface{}

		if channelID != "" {
			query = `SELECT id, direct_url FROM videos WHERE channel_id = ? AND (file_size = 0 OR file_size IS NULL) AND direct_url != ''`
			queryArgs = []interface{}{channelID}
			fmt.Printf("🔍 Finding videos without file sizes for channel %s...\n", channelID)
		} else {
			query = `SELECT id, direct_url FROM videos WHERE (file_size = 0 OR file_size IS NULL) AND direct_url != ''`
			fmt.Printf("🔍 Finding all videos without file sizes...\n")
		}

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			fmt.Printf("❌ Failed to query videos: %v\n", err)
			return
		}
		defer rows.Close()

		var videos []Video
		for rows.Next() {
			var video Video
			if err := rows.Scan(&video.ID, &video.DirectURL); err != nil {
				fmt.Printf("⚠️  Warning: failed to scan video: %v\n", err)
				continue
			}
			videos = append(videos, video)
		}

		if len(videos) == 0 {
			fmt.Printf("✅ All videos already have file sizes!\n")
			return
		}

		var pkgs []string
		for _, v := range videos {
			pkgs = append(pkgs, v.DirectURL)
		}

		fmt.Printf("📏 Found %d videos without file sizes. Starting fetch...\n", len(videos))
		if _, err := tea.NewProgram(tui.NewPackageManagerModel("fetch", pkgs, fetchFunc)).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
		// fetchVideoFileSizes(videos)
		fmt.Printf("✅ File size fetching completed!\n")
	},
}

func fetchFunc(pkg string) tea.Cmd {
	returnFunc := func() tea.Msg {
		return tui.InstalledPkgMsg(pkg)
	}
	size, err := getFileSize(pkg)
	if err != nil {
		return returnFunc
	}
	db, err := GetDB()
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		fetchFunc(pkg)
	}
	db.Exec("UPDATE videos SET file_size = ? WHERE direct_url = ?", size, pkg)
	return returnFunc
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncChannelsCmd)
	syncCmd.AddCommand(syncChannelCmd)
	syncCmd.AddCommand(syncCheckCmd)
	syncCmd.AddCommand(syncAllCmd)
	syncCmd.AddCommand(syncFileSizesCmd)

	// Add flags
	syncChannelCmd.Flags().BoolP("skip-file-sizes", "s", false, "Skip fetching file sizes for new videos")
	syncAllCmd.Flags().BoolP("skip-file-sizes", "s", false, "Skip fetching file sizes for new videos")
}
