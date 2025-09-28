/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Could not connect to database: %v\n", err)
			return
		}

		var oldCount int
		db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&oldCount)
		fmt.Printf("📋 Current channels in database: %d\n", oldCount)

		if err := client.FetchAllChannels(); err != nil {
			fmt.Printf("❌ Failed to fetch channels: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully fetched and replaced all channel data!\n")

		// Show updated statistics
		fmt.Println("\n📊 Updated database statistics:")
		var newCount int
		db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&newCount)
		fmt.Printf("📋 Channels in database: %d\n", newCount)

		if newCount > oldCount {
			fmt.Printf("🆕 Added %d new channels\n", newCount-oldCount)
		} else if newCount < oldCount {
			fmt.Printf("🗑️  Removed %d obsolete channels\n", oldCount-newCount)
		} else {
			fmt.Printf("� Updated %d existing channels\n", newCount)
		}
	},
}

// fetchChannelCmd fetches a specific channel
var fetchChannelCmd = &cobra.Command{
	Use:   "channel <channel-id>",
	Short: "Fetch specific channel details and videos",
	Long: `Fetch detailed information about a specific channel and its videos from banned.video API.

By default, fetches the first 50 videos. Use --all flag to recursively fetch ALL videos.

Examples:
  banned fetch channel 5b885d33e6646a0015a6fa2d           # Fetch channel + first 50 videos
  banned fetch channel 5b885d33e6646a0015a6fa2d --all     # Fetch channel + ALL videos recursively`,
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

		// Fetch videos for this channel
		fetchAll, _ := cmd.Flags().GetBool("all")
		var videos []Video
		var err error

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

		// Fetch file sizes for videos in background
		if len(videos) > 0 {
			go func() {
				fmt.Printf("📏 Fetching file sizes for %d videos in background...\n", len(videos))
				fetchVideoFileSizes(videos)
				fmt.Printf("✅ File sizes updated!\n")
			}()
		}

		// Show updated statistics
		fmt.Println("\n📊 Updated database statistics:")
		db, err := GetDB()
		if err != nil {
			fmt.Printf("⚠️  Could not check database: %v\n", err)
			return
		}

		var channelCount, videoCount int
		db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&channelCount)
		db.QueryRow("SELECT COUNT(*) FROM videos").Scan(&videoCount)

		fmt.Printf("📋 Channels in database: %d\n", channelCount)
		fmt.Printf("📋 Videos in database: %d\n", videoCount)
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

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.AddCommand(fetchChannelsCmd)
	fetchCmd.AddCommand(fetchChannelCmd)
	rootCmd.AddCommand(listChannelsCmd)

	// Add flags for fetch channel command
	fetchChannelCmd.Flags().BoolP("all", "a", false, "Fetch ALL videos recursively (may take time for channels with many videos)")
}
