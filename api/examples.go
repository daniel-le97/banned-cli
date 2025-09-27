package api

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Example demonstrates how to use the banned.video API client
func Example() {
	// Create a new API client with default settings
	api := New()

	// Or create with custom configuration
	apiCustom := NewWithConfig(
		"https://api.banned.video/graphql",
		"my-app/1.0",
		30*time.Second,
	)

	// Use the default API for examples
	ctx := context.Background()

	// Test API connectivity
	fmt.Println("Testing API connectivity...")
	if err := api.Ping(ctx); err != nil {
		log.Printf("API connectivity test failed: %v", err)
		return
	}
	fmt.Println("API is reachable!")

	// Fetch all channels
	fmt.Println("\nFetching channels...")
	channels, err := api.Channels.GetAllChannels(ctx, 0, 10)
	if err != nil {
		log.Printf("Failed to fetch channels: %v", err)
		return
	}

	fmt.Printf("Found %d channels:\n", len(channels))
	for i, channel := range channels {
		fmt.Printf("%d. %s (ID: %s) - %d videos\n",
			i+1, channel.Title, channel.ID, int(channel.TotalVideos))
	}

	// Get detailed info for the first channel
	if len(channels) > 0 {
		channelID := channels[0].ID
		fmt.Printf("\nFetching detailed info for channel: %s\n", channels[0].Title)

		channel, err := api.Channels.GetChannel(ctx, channelID)
		if err != nil {
			log.Printf("Failed to fetch channel details: %v", err)
		} else {
			fmt.Printf("Channel: %s\n", channel.Title)
			fmt.Printf("Summary: %s\n", channel.Summary)
			fmt.Printf("Total Videos: %.0f\n", channel.TotalVideos)
			fmt.Printf("Total Views: %.0f\n", channel.TotalVideoViews)
			fmt.Printf("Total Likes: %.0f\n", channel.TotalLikes)
		}

		// Fetch some videos from this channel
		fmt.Printf("\nFetching first 5 videos from channel: %s\n", channels[0].Title)
		videos, err := api.Channels.GetChannelVideos(ctx, channelID, 0, 5)
		if err != nil {
			log.Printf("Failed to fetch channel videos: %v", err)
		} else {
			fmt.Printf("Found %d videos:\n", len(videos))
			for i, video := range videos {
				duration := formatDuration(video.VideoDuration)
				fmt.Printf("%d. %s - %s (%.0f views, %.0f likes)\n",
					i+1, video.Title, duration, video.PlayCount, video.LikeCount)
			}
		}

		// Demonstrate recursive fetching with limit
		fmt.Printf("\nFetching up to 20 videos from channel: %s\n", channels[0].Title)
		allVideos, err := api.Channels.FetchAllChannelVideosWithLimit(ctx, channelID, 20, 5)
		if err != nil {
			log.Printf("Failed to fetch all channel videos: %v", err)
		} else {
			fmt.Printf("Fetched %d videos total\n", len(allVideos))
		}
	}

	// Fetch hot videos
	fmt.Println("\nFetching hot/trending videos...")
	hotVideos, err := api.Videos.GetHotVideos(ctx, 0, 5)
	if err != nil {
		log.Printf("Failed to fetch hot videos: %v", err)
	} else {
		fmt.Printf("Found %d hot videos:\n", len(hotVideos))
		for i, video := range hotVideos {
			duration := formatDuration(video.VideoDuration)
			fmt.Printf("%d. %s - %s (%.0f views)\n",
				i+1, video.Title, duration, video.PlayCount)
		}
	}

	// Fetch new videos
	fmt.Println("\nFetching newest videos...")
	newVideos, err := api.Videos.GetNewVideos(ctx, 0, 5)
	if err != nil {
		log.Printf("Failed to fetch new videos: %v", err)
	} else {
		fmt.Printf("Found %d new videos:\n", len(newVideos))
		for i, video := range newVideos {
			duration := formatDuration(video.VideoDuration)
			fmt.Printf("%d. %s - %s (%.0f views)\n",
				i+1, video.Title, duration, video.PlayCount)
		}
	}

	// Get detailed video information
	if len(newVideos) > 0 {
		videoID := newVideos[0].ID
		fmt.Printf("\nFetching detailed info for video: %s\n", newVideos[0].Title)

		video, err := api.Videos.GetVideo(ctx, videoID)
		if err != nil {
			log.Printf("Failed to fetch video details: %v", err)
		} else {
			fmt.Printf("Video: %s\n", video.Title)
			fmt.Printf("Summary: %s\n", video.Summary)
			fmt.Printf("Duration: %s\n", formatDuration(video.VideoDuration))
			fmt.Printf("Published: %t\n", video.Published)
			fmt.Printf("Live: %t\n", video.Live)
			fmt.Printf("Play Count: %.0f\n", video.PlayCount)
			fmt.Printf("Like Count: %.0f\n", video.LikeCount)
			fmt.Printf("Channel: %s\n", video.Channel.Title)

			// Get download URLs
			downloadURL := api.Videos.GetVideoDownloadURL(video)
			posterURL := api.Videos.GetVideoPosterURL(video)
			audioURL := api.Videos.GetVideoAudioURL(video)

			if downloadURL != "" {
				fmt.Printf("Download URL: %s\n", downloadURL)
			}
			if posterURL != "" {
				fmt.Printf("Poster URL: %s\n", posterURL)
			}
			if audioURL != "" {
				fmt.Printf("Audio URL: %s\n", audioURL)
			}

			// Show tags
			if len(video.Tags) > 0 {
				fmt.Print("Tags: ")
				for i, tag := range video.Tags {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Print(tag.Name)
				}
				fmt.Println()
			}
		}
	}

	// Search channels (basic client-side implementation)
	fmt.Println("\nSearching for channels containing 'alex'...")
	searchResults, err := api.SearchChannels(ctx, "alex")
	if err != nil {
		log.Printf("Failed to search channels: %v", err)
	} else {
		fmt.Printf("Found %d matching channels:\n", len(searchResults))
		for i, channel := range searchResults {
			fmt.Printf("%d. %s (%.0f videos)\n", i+1, channel.Title, channel.TotalVideos)
		}
	}

	// Bulk fetch multiple channels
	if len(channels) >= 2 {
		channelIDs := []string{channels[0].ID, channels[1].ID}
		fmt.Printf("\nBulk fetching data for %d channels...\n", len(channelIDs))

		bulkData, err := api.BulkFetchChannelData(ctx, channelIDs, 3)
		if err != nil {
			log.Printf("Failed to bulk fetch channel data: %v", err)
		} else {
			fmt.Printf("Successfully fetched data for %d channels:\n", len(bulkData))
			for _, channelData := range bulkData {
				fmt.Printf("- %s: %d videos loaded\n", channelData.Title, len(channelData.Videos))
			}
		}
	}

	// Print API information
	fmt.Println("\nAPI Client Information:")
	info := api.GetAPIInfo()
	for key, value := range info {
		fmt.Printf("- %s: %v\n", key, value)
	}

	fmt.Println("\nExample completed successfully!")

	// Don't forget to handle the unused apiCustom to avoid compiler warnings
	_ = apiCustom
}

// ExampleBatchProcessing shows how to process large amounts of data efficiently
func ExampleBatchProcessing() {
	api := New()
	ctx := context.Background()

	fmt.Println("Starting batch processing example...")

	// Fetch all channels
	channels, err := api.Channels.GetAllChannels(ctx, 0, 100)
	if err != nil {
		log.Printf("Failed to fetch channels: %v", err)
		return
	}

	fmt.Printf("Processing %d channels...\n", len(channels))

	totalVideos := 0
	for i, channel := range channels {
		fmt.Printf("Processing channel %d/%d: %s\n", i+1, len(channels), channel.Title)

		// Fetch up to 50 videos from each channel
		videos, err := api.Channels.FetchAllChannelVideosWithLimit(ctx, channel.ID, 50, 10)
		if err != nil {
			log.Printf("Failed to fetch videos for channel %s: %v", channel.Title, err)
			continue
		}

		totalVideos += len(videos)
		fmt.Printf("  - Fetched %d videos\n", len(videos))

		// Process videos (example: count by tags)
		tagCounts := make(map[string]int)
		for _, video := range videos {
			for _, tag := range video.Tags {
				tagCounts[tag.Name]++
			}
		}

		fmt.Printf("  - Found %d unique tags\n", len(tagCounts))
	}

	fmt.Printf("Batch processing completed: %d total videos processed\n", totalVideos)
}

// formatDuration converts seconds to a human-readable duration
func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "Unknown"
	}

	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
