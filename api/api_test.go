package api_test

import (
	"context"
	"testing"
	"time"

	"libertyarchive.com/banned/api"
)

func TestNewAPI(t *testing.T) {
	apiClient := api.New()

	if apiClient == nil {
		t.Fatal("API client should not be nil")
	}

	if apiClient.Client == nil {
		t.Fatal("HTTP client should not be nil")
	}

	if apiClient.Channels == nil {
		t.Fatal("Channels service should not be nil")
	}

	if apiClient.Videos == nil {
		t.Fatal("Videos service should not be nil")
	}

	info := apiClient.GetAPIInfo()
	if info["base_url"] != "https://api.banned.video/graphql" {
		t.Errorf("Expected base URL to be https://api.banned.video/graphql, got %s", info["base_url"])
	}
}

func TestNewAPIWithConfig(t *testing.T) {
	baseURL := "https://custom.api.url/graphql"
	userAgent := "test-agent/1.0"
	timeout := 60 * time.Second

	apiClient := api.NewWithConfig(baseURL, userAgent, timeout)

	info := apiClient.GetAPIInfo()
	if info["base_url"] != baseURL {
		t.Errorf("Expected base URL to be %s, got %s", baseURL, info["base_url"])
	}

	if info["user_agent"] != userAgent {
		t.Errorf("Expected user agent to be %s, got %s", userAgent, info["user_agent"])
	}

	if info["timeout"] != timeout {
		t.Errorf("Expected timeout to be %v, got %v", timeout, info["timeout"])
	}
}

func TestChannelVideosParams(t *testing.T) {
	params := api.ChannelVideosParams{
		ChannelID:          "test-channel-id",
		IncludeUnlisted:    true,
		IncludeUnpublished: false,
		IncludeLive:        true,
		Offset:             0,
		Limit:              50,
	}

	if params.ChannelID != "test-channel-id" {
		t.Errorf("Expected ChannelID to be test-channel-id, got %s", params.ChannelID)
	}

	if !params.IncludeUnlisted {
		t.Error("Expected IncludeUnlisted to be true")
	}

	if params.IncludeUnpublished {
		t.Error("Expected IncludeUnpublished to be false")
	}

	if params.Limit != 50 {
		t.Errorf("Expected Limit to be 50, got %d", params.Limit)
	}
}

func TestGraphQLRequest(t *testing.T) {
	req := &api.GraphQLRequest{
		Query: "query { test }",
		Variables: map[string]interface{}{
			"id":    "test-id",
			"limit": 10,
		},
	}

	if req.Query != "query { test }" {
		t.Errorf("Expected query to be 'query { test }', got %s", req.Query)
	}

	if req.Variables["id"] != "test-id" {
		t.Errorf("Expected id variable to be 'test-id', got %v", req.Variables["id"])
	}

	if req.Variables["limit"] != 10 {
		t.Errorf("Expected limit variable to be 10, got %v", req.Variables["limit"])
	}
}

// Mock test for API connectivity (would need real API for integration tests)
func TestAPIMethods(t *testing.T) {
	apiClient := api.New()
	ctx := context.Background()

	// Test that methods exist and can be called (will fail without real API)
	t.Run("channels service exists", func(t *testing.T) {
		if apiClient.Channels == nil {
			t.Fatal("Channels service should not be nil")
		}
	})

	t.Run("videos service exists", func(t *testing.T) {
		if apiClient.Videos == nil {
			t.Fatal("Videos service should not be nil")
		}
	})

	t.Run("context methods", func(t *testing.T) {
		// Test context usage
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		// These would fail without real API, but we're testing method signatures
		_ = ctxWithTimeout
	})
}

func TestVideoHelperMethods(t *testing.T) {
	videoService := api.NewVideoService(api.NewClient())

	// Test video with uploads
	video := &api.Video{
		ID:    "test-video",
		Title: "Test Video",
		VideoUpload: &api.Upload{
			Name:    "video.mp4",
			Encoding: "video.mp4",
		},
		PosterUpload: &api.Upload{
			Name:    "poster.jpg",
			Encoding: "poster.jpg",
		},
		AudioUpload: &api.Upload{
			Name:    "audio.mp3",
			Encoding: "audio.mp3",
		},
		LargeImage: "https://example.com/large.jpg",
	}

	// Test download URL
	downloadURL := videoService.GetVideoDownloadURL(video)
	if downloadURL != "video.mp4" {
		t.Errorf("Expected name, got %s", downloadURL)
	}

	// Test poster URL
	posterURL := videoService.GetVideoPosterURL(video)
	if posterURL != "poster.jpg" {
		t.Errorf("Expected name, got %s", posterURL)
	}

	// Test audio URL
	audioURL := videoService.GetVideoAudioURL(video)
	if audioURL != "audio.mp3" {
		t.Errorf("Expected name, got %s", audioURL)
	}

	// Test fallback to regular URL when CDN is not available
	video.VideoUpload.Encoding = ""
	downloadURL = videoService.GetVideoDownloadURL(video)
	if downloadURL != "video.mp4" {
		t.Errorf("Expected name fallback, got %s", downloadURL)
	}

	// Test fallback to large image when no poster upload
	video.PosterUpload = nil
	posterURL = videoService.GetVideoPosterURL(video)
	if posterURL != "https://example.com/large.jpg" {
		t.Errorf("Expected large image fallback, got %s", posterURL)
	}
}
