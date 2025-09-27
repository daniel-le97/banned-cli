package api

import (
	"context"
	"time"
)

// API provides a high-level interface to the banned.video GraphQL API
type API struct {
	Client   *Client
	Channels *ChannelService
	Videos   *VideoService
}

// New creates a new API instance with default configuration
func New() *API {
	client := NewClient()
	return &API{
		Client:   client,
		Channels: NewChannelService(client),
		Videos:   NewVideoService(client),
	}
}

// NewWithConfig creates a new API instance with custom configuration
func NewWithConfig(baseURL, userAgent string, timeout time.Duration) *API {
	client := NewClientWithConfig(baseURL, userAgent, timeout)
	return &API{
		Client:   client,
		Channels: NewChannelService(client),
		Videos:   NewVideoService(client),
	}
}

// Ping tests the API connection
func (a *API) Ping(ctx context.Context) error {
	// Try to fetch a small amount of channels as a connectivity test
	_, err := a.Channels.GetAllChannels(ctx, 0, 1)
	return err
}

// GetAPIInfo returns information about the API client
func (a *API) GetAPIInfo() map[string]interface{} {
	return map[string]interface{}{
		"base_url":    a.Client.BaseURL,
		"user_agent":  a.Client.UserAgent,
		"timeout":     a.Client.HTTPClient.Timeout,
		"version":     "1.0.0",
		"description": "banned.video GraphQL API Client",
	}
}

// SetUserAgent updates the user agent for API requests
func (a *API) SetUserAgent(userAgent string) {
	a.Client.UserAgent = userAgent
}

// SetTimeout updates the HTTP timeout for API requests
func (a *API) SetTimeout(timeout time.Duration) {
	a.Client.HTTPClient.Timeout = timeout
}

// BulkFetchChannelData fetches multiple channels with their videos in parallel
func (a *API) BulkFetchChannelData(ctx context.Context, channelIDs []string, videoLimit int) (map[string]*Channel, error) {
	result := make(map[string]*Channel)

	// For simplicity, we'll fetch sequentially here
	// In a production environment, you might want to implement concurrent fetching
	for _, channelID := range channelIDs {
		params := ChannelVideosParams{
			ChannelID:          channelID,
			IncludeUnlisted:    false,
			IncludeUnpublished: false,
			IncludeLive:        true,
			Offset:             0,
			Limit:              videoLimit,
		}

		channel, err := a.Channels.GetChannelWithVideos(ctx, params)
		if err != nil {
			// Log error but continue with other channels
			continue
		}

		result[channelID] = channel
	}

	return result, nil
}

// SearchChannels searches for channels by title (basic implementation)
// Note: This might need to be adjusted based on actual search capabilities in the API
func (a *API) SearchChannels(ctx context.Context, query string) ([]Channel, error) {
	// For now, we'll fetch all channels and filter client-side
	// In a real implementation, you'd want a proper search endpoint
	channels, err := a.Channels.GetAllChannels(ctx, 0, 1000)
	if err != nil {
		return nil, err
	}

	var filtered []Channel
	for _, channel := range channels {
		if containsIgnoreCase(channel.Title, query) || containsIgnoreCase(channel.Summary, query) {
			filtered = append(filtered, channel)
		}
	}

	return filtered, nil
}

// Helper function for case-insensitive string searching
func containsIgnoreCase(text, search string) bool {
	if len(search) == 0 {
		return true
	}
	if len(text) == 0 {
		return false
	}

	// Simple case-insensitive contains check
	textLower := make([]rune, 0, len(text))
	for _, r := range text {
		if r >= 'A' && r <= 'Z' {
			r += 32 // Convert to lowercase
		}
		textLower = append(textLower, r)
	}

	searchLower := make([]rune, 0, len(search))
	for _, r := range search {
		if r >= 'A' && r <= 'Z' {
			r += 32 // Convert to lowercase
		}
		searchLower = append(searchLower, r)
	}

	textStr := string(textLower)
	searchStr := string(searchLower)

	for i := 0; i <= len(textStr)-len(searchStr); i++ {
		if textStr[i:i+len(searchStr)] == searchStr {
			return true
		}
	}

	return false
}
