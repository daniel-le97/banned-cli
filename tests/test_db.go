package main

import (
	"fmt"
	"os"
	"time"

	"github.com/daniel-le97/banned-cli/db"
)

func testDatabase() {
	fmt.Println("🔄 Testing bbolt database CRUD operations...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Test 1: Database connection
	fmt.Println("\n1. Testing database connection...")
	database, err := db.GetDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database connection successful")

	// Test 2: Settings CRUD operations
	fmt.Println("\n2. Testing Settings CRUD operations...")
	testSettings()

	// Test 3: Channel CRUD operations
	fmt.Println("\n3. Testing Channel CRUD operations...")
	testChannels()

	// Test 4: Video CRUD operations
	fmt.Println("\n4. Testing Video CRUD operations...")
	testVideos()

	// Test 5: Download CRUD operations
	fmt.Println("\n5. Testing Download CRUD operations...")
	testDownloads()

	// Test 6: Database statistics
	fmt.Println("\n6. Testing database statistics...")
	testStatistics()

	fmt.Printf("\n✅ All CRUD tests completed successfully!\n")
	database.Close()
}

func testSettings() {
	// CREATE: Set test settings
	testKey := "test_setting"
	testValue := "test_value_123"

	fmt.Printf("   Creating setting: %s = %s\n", testKey, testValue)
	if err := db.SetSetting(testKey, testValue); err != nil {
		fmt.Printf("❌ Failed to create setting: %v\n", err)
		return
	}

	// READ: Get the setting back
	fmt.Printf("   Reading setting: %s\n", testKey)
	retrievedValue, err := db.GetSetting(testKey)
	if err != nil {
		fmt.Printf("❌ Failed to read setting: %v\n", err)
		return
	}

	if retrievedValue != testValue {
		fmt.Printf("❌ Value mismatch: expected %s, got %s\n", testValue, retrievedValue)
		return
	}

	// UPDATE: Change the setting value
	updatedValue := "updated_value_456"
	fmt.Printf("   Updating setting: %s = %s\n", testKey, updatedValue)
	if err := db.SetSetting(testKey, updatedValue); err != nil {
		fmt.Printf("❌ Failed to update setting: %v\n", err)
		return
	}

	// READ: Verify the update
	retrievedValue, err = db.GetSetting(testKey)
	if err != nil {
		fmt.Printf("❌ Failed to read updated setting: %v\n", err)
		return
	}

	if retrievedValue != updatedValue {
		fmt.Printf("❌ Update failed: expected %s, got %s\n", updatedValue, retrievedValue)
		return
	}

	fmt.Println("   ✅ Settings CRUD operations successful")
}

func testChannels() {
	// CREATE: Store test channel
	testChannel := db.Channel{
		ID:              "test_channel_123",
		Title:           "Test Channel",
		Summary:         "A test channel for CRUD operations",
		TextInfo:        "Test channel description",
		Avatar:          "https://example.com/avatar.jpg",
		CoverImage:      "https://example.com/cover.jpg",
		IsLive:          false,
		TotalVideos:     5,
		TotalVideoViews: 1000,
		TotalLikes:      50,
		ShowInfo: &db.ShowInfo{
			Times: "Mon-Fri 9AM-5PM",
			Phone: "555-0123",
		},
		Links: &db.Links{
			Website:  "https://example.com",
			Twitter:  "@testchannel",
			Facebook: "testchannel",
		},
	}

	fmt.Printf("   Creating channel: %s\n", testChannel.ID)
	if err := db.StoreChannel(testChannel); err != nil {
		fmt.Printf("❌ Failed to create channel: %v\n", err)
		return
	}

	// READ: Get the channel back
	fmt.Printf("   Reading channel: %s\n", testChannel.ID)
	retrievedChannel, err := db.GetChannel(testChannel.ID)
	if err != nil {
		fmt.Printf("❌ Failed to read channel: %v\n", err)
		return
	}

	if retrievedChannel.Title != testChannel.Title {
		fmt.Printf("❌ Channel title mismatch: expected %s, got %s\n", testChannel.Title, retrievedChannel.Title)
		return
	}

	// UPDATE: Modify the channel
	testChannel.Title = "Updated Test Channel"
	testChannel.TotalVideos = 10
	fmt.Printf("   Updating channel: %s\n", testChannel.ID)
	if err := db.StoreChannel(testChannel); err != nil {
		fmt.Printf("❌ Failed to update channel: %v\n", err)
		return
	}

	// READ: Verify the update
	retrievedChannel, err = db.GetChannel(testChannel.ID)
	if err != nil {
		fmt.Printf("❌ Failed to read updated channel: %v\n", err)
		return
	}

	if retrievedChannel.Title != "Updated Test Channel" {
		fmt.Printf("❌ Channel update failed: expected 'Updated Test Channel', got %s\n", retrievedChannel.Title)
		return
	}

	// DELETE: Remove the test channel
	fmt.Printf("   Deleting channel: %s\n", testChannel.ID)
	if err := db.DeleteChannel(testChannel.ID); err != nil {
		fmt.Printf("❌ Failed to delete channel: %v\n", err)
		return
	}

	// Verify deletion
	_, err = db.GetChannel(testChannel.ID)
	if err == nil {
		fmt.Printf("❌ Channel deletion failed: channel still exists\n")
		return
	}

	fmt.Println("   ✅ Channel CRUD operations successful")
}

func testVideos() {
	// First create a test channel for the videos
	testChannel := db.Channel{
		ID:    "test_video_channel",
		Title: "Test Video Channel",
	}
	db.StoreChannel(testChannel)

	// CREATE: Store test videos
	testVideos := []db.Video{
		{
			ID:            "test_video_1",
			Title:         "Test Video 1",
			Summary:       "First test video",
			LargeImage:    "https://example.com/thumb1.jpg",
			VideoDuration: 120.5,
			CreatedAt:     "2025-01-01T00:00:00Z",
			DirectURL:     "https://example.com/video1.mp4",
			PlayCount:     100,
			LikeCount:     10,
			AngerCount:    1,
			Published:     true,
			FileSize:      1024000,
		},
		{
			ID:            "test_video_2",
			Title:         "Test Video 2",
			Summary:       "Second test video",
			LargeImage:    "https://example.com/thumb2.jpg",
			VideoDuration: 180.0,
			CreatedAt:     "2025-01-02T00:00:00Z",
			DirectURL:     "https://example.com/video2.mp4",
			PlayCount:     200,
			LikeCount:     20,
			AngerCount:    2,
			Published:     true,
			FileSize:      2048000,
		},
	}

	fmt.Printf("   Creating %d videos for channel: %s\n", len(testVideos), testChannel.ID)
	if err := db.StoreVideos(testChannel.ID, testVideos); err != nil {
		fmt.Printf("❌ Failed to create videos: %v\n", err)
		return
	}

	// READ: Get videos back
	fmt.Printf("   Reading videos for channel: %s\n", testChannel.ID)
	retrievedVideos, err := db.GetChannelVideos(testChannel.ID, 10, 0)
	if err != nil {
		fmt.Printf("❌ Failed to read videos: %v\n", err)
		return
	}

	if len(retrievedVideos) != len(testVideos) {
		fmt.Printf("❌ Video count mismatch: expected %d, got %d\n", len(testVideos), len(retrievedVideos))
		return
	}

	// UPDATE: Update video file size
	newFileSize := int64(3072000)
	fmt.Printf("   Updating video file size: %s -> %d bytes\n", testVideos[0].ID, newFileSize)
	if err := db.UpdateVideoFileSize(testVideos[0].ID, newFileSize); err != nil {
		fmt.Printf("❌ Failed to update video file size: %v\n", err)
		return
	}

	// READ: Verify the update
	updatedVideo, err := db.GetVideo(testVideos[0].ID)
	if err != nil {
		fmt.Printf("❌ Failed to read updated video: %v\n", err)
		return
	}

	if updatedVideo.FileSize != newFileSize {
		fmt.Printf("❌ Video file size update failed: expected %d, got %d\n", newFileSize, updatedVideo.FileSize)
		return
	}

	// Clean up test data
	db.DeleteChannel(testChannel.ID)
	fmt.Println("   ✅ Video CRUD operations successful")
}

func testDownloads() {
	// CREATE: Store test download record
	testDownload := db.Download{
		ID:             "test_download_123",
		URL:            "https://example.com/testvideo.mp4",
		Title:          "Test Download Video",
		Filename:       "testvideo.mp4",
		FilePath:       "/downloads/testvideo.mp4",
		FileSize:       1024000,
		Status:         "completed",
		CreatedAt:      time.Now(),
		CompletedAt:    &time.Time{},
		ErrorMessage:   "",
		TorrentCreated: false,
	}
	*testDownload.CompletedAt = time.Now()

	fmt.Printf("   Creating download record: %s\n", testDownload.ID)
	if err := db.StoreDownload(testDownload); err != nil {
		fmt.Printf("❌ Failed to create download: %v\n", err)
		return
	}

	// READ: Get the download back
	fmt.Printf("   Reading download record: %s\n", testDownload.ID)
	retrievedDownload, err := db.GetDownload(testDownload.ID)
	if err != nil {
		fmt.Printf("❌ Failed to read download: %v\n", err)
		return
	}

	if retrievedDownload.Title != testDownload.Title {
		fmt.Printf("❌ Download title mismatch: expected %s, got %s\n", testDownload.Title, retrievedDownload.Title)
		return
	}

	// UPDATE: Update torrent info
	torrentPath := "/downloads/testvideo.mp4.torrent"
	fmt.Printf("   Updating torrent info for download: %s\n", testDownload.ID)
	if err := db.UpdateDownloadTorrentInfo(testDownload.FilePath, torrentPath); err != nil {
		fmt.Printf("❌ Failed to update download torrent info: %v\n", err)
		return
	}

	// READ: Verify the update
	retrievedDownload, err = db.GetDownload(testDownload.ID)
	if err != nil {
		fmt.Printf("❌ Failed to read updated download: %v\n", err)
		return
	}

	if !retrievedDownload.TorrentCreated {
		fmt.Printf("❌ Download torrent update failed: torrent not marked as created\n")
		return
	}

	// DELETE: Remove test download
	fmt.Printf("   Deleting download record: %s\n", testDownload.ID)
	if err := db.DeleteDownload(testDownload.ID); err != nil {
		fmt.Printf("❌ Failed to delete download: %v\n", err)
		return
	}

	fmt.Println("   ✅ Download CRUD operations successful")
}

func testStatistics() {
	// Test bucket statistics
	fmt.Println("   Getting database statistics...")
	stats, err := db.GetBucketStats()
	if err != nil {
		fmt.Printf("❌ Failed to get bucket stats: %v\n", err)
		return
	}

	fmt.Println("   Current bucket statistics:")
	for bucket, count := range stats {
		fmt.Printf("     %s: %d records\n", bucket, count)
	}

	// Test count functions
	channelCount, err := db.CountChannels()
	if err != nil {
		fmt.Printf("❌ Failed to count channels: %v\n", err)
		return
	}

	videoCount, err := db.CountVideos()
	if err != nil {
		fmt.Printf("❌ Failed to count videos: %v\n", err)
		return
	}

	fmt.Printf("   Total channels: %d\n", channelCount)
	fmt.Printf("   Total videos: %d\n", videoCount)
	fmt.Println("   ✅ Statistics operations successful")
}

func main() {
	testDatabase()
}
