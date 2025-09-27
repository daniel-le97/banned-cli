package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client represents the API client
type Client struct {
	APIEndpoint string
}

// NewClient creates a new API client
func NewClient(apiEndpoint string) *Client {
	return &Client{
		APIEndpoint: apiEndpoint,
	}
}

// Video represents a video structure
type Video struct {
	ID            string  `json:"_id"`
	Title         string  `json:"title"`
	Summary       string  `json:"summary"`
	LargeImage    string  `json:"largeImage"`
	VideoDuration float64 `json:"videoDuration"`
	CreatedAt     string  `json:"createdAt"`
	DirectURL     string  `json:"directUrl"`
	PlayCount     int     `json:"playCount,omitempty"`
	LikeCount     int     `json:"likeCount,omitempty"`
	AngerCount    int     `json:"angerCount,omitempty"`
	EmbedURL      string  `json:"embedUrl,omitempty"`
	Published     bool    `json:"published,omitempty"`
}

// Channel represents a channel structure
type Channel struct {
	ID         string    `json:"_id"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary,omitempty"`
	TextInfo   string    `json:"textInfo,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
	CoverImage string    `json:"coverImage,omitempty"`
	IsLive     bool      `json:"isLive,omitempty"`
	Videos     []Video   `json:"videos,omitempty"`
	ShowInfo   *ShowInfo `json:"showInfo,omitempty"`
	Links      *Links    `json:"links,omitempty"`
}

type ShowInfo struct {
	Times string `json:"times"`
	Phone string `json:"phone"`
}

type Links struct {
	Website       string `json:"website"`
	Facebook      string `json:"facebook"`
	Twitter       string `json:"twitter"`
	Gab           string `json:"gab"`
	Minds         string `json:"minds"`
	Telegram      string `json:"telegram"`
	SubscribeStar string `json:"subscribeStar"`
}

// GraphQL request/response structures
type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
}

type GetChannelVideosResponse struct {
	Data struct {
		GetChannel struct {
			ID     string  `json:"_id"`
			Videos []Video `json:"videos"`
		} `json:"getChannel"`
	} `json:"data"`
}

type GetChannelResponse struct {
	Data struct {
		GetChannelByIDOrTitle Channel `json:"getChannelByIdOrTitle"`
	} `json:"data"`
}

type GetAllChannelsResponse struct {
	Data struct {
		GetAllChannels []Channel `json:"getAllChannels"`
	} `json:"data"`
}

// DownloadVideo downloads a video from a URL and saves it to a file
func DownloadVideo(url, filename string) error {
	if url == "" || filename == "" {
		return fmt.Errorf("URL and filename are required")
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch video: status %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FetchVideos fetches videos from the API and stores them in the database
func (c *Client) FetchVideos(channelID string, offset, limit int) ([]Video, error) {
	// First try to get videos from database if we have them
	if offset == 0 { // Only check database for fresh requests
		dbVideos, err := GetChannelVideos(channelID, limit, offset)
		if err == nil && len(dbVideos) > 0 {
			return dbVideos, nil
		}
	}

	// Fetch from API if not in database or offset > 0
	query := `
		query GetChannelVideos($id: String!, $limit: Float, $offset: Float) {
			getChannel(id: $id) {
				_id
				videos(limit: $limit, offset: $offset) {
					_id
					title
					summary
					largeImage
					videoDuration
					createdAt
					directUrl
					playCount
					likeCount
					angerCount
					embedUrl
					published
				}
			}
		}`

	variables := map[string]interface{}{
		"id":     channelID,
		"limit":  float64(limit),
		"offset": float64(offset),
	}

	request := GraphQLRequest{
		OperationName: "GetChannelVideos",
		Variables:     variables,
		Query:         query,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response GetChannelVideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := response.Data.GetChannel.Videos

	// Store videos in database
	if len(videos) > 0 {
		if err := StoreVideos(channelID, videos); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to store videos in database: %v\n", err)
		}
	}

	return videos, nil
}

// FetchAllVideos recursively fetches all videos for a channel using pagination
func (c *Client) FetchAllVideos(channelID string) ([]Video, error) {
	startTime := time.Now()
	var allVideos []Video
	offset := 0
	limit := 200    // Larger page size to minimize API calls and avoid rate limiting
	batchSize := 50 // Smaller batch size for database storage
	var totalStored int

	fmt.Printf("🔄 Fetching all videos for channel %s...\n", channelID)

	for {
		// Fetch this page of videos (bypass cache for pagination)
		videos, err := c.fetchVideosPage(channelID, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos at offset %d: %w", offset, err)
		}

		// If no videos returned, we've reached the end
		if len(videos) == 0 {
			break
		}

		allVideos = append(allVideos, videos...)

		// Store videos in smaller batches to avoid database locks
		if len(videos) > 0 {
			stored := 0
			for i := 0; i < len(videos); i += batchSize {
				end := i + batchSize
				if end > len(videos) {
					end = len(videos)
				}
				batch := videos[i:end]

				if err := StoreVideosBatch(channelID, batch); err != nil {
					fmt.Printf("⚠️  Warning: failed to store batch %d-%d: %v\n", i+1, end, err)
					// Continue with next batch even if this one fails
				} else {
					stored += len(batch)
				}
			}
			totalStored += stored
		}

		fmt.Printf("📥 Fetched %d videos (total: %d, stored: %d)\n", len(videos), len(allVideos), totalStored)

		// If we got fewer videos than requested, we've reached the end
		if len(videos) < limit {
			break
		}

		// Move to next page
		offset += len(videos)
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("✅ Completed! Fetched %d total videos (%d stored) for channel %s in %v\n", len(allVideos), totalStored, channelID, elapsedTime)
	return allVideos, nil
}

// FetchAllVideosWithLimit is like FetchAllVideos but stops after maxVideos for testing
func (c *Client) FetchAllVideosWithLimit(channelID string, maxVideos int) ([]Video, error) {
	startTime := time.Now()
	var allVideos []Video
	offset := 0
	limit := 200    // Larger page size to minimize API calls
	batchSize := 50 // Smaller batch size for database storage
	var totalStored int

	fmt.Printf("🔄 Fetching up to %d videos for channel %s...\n", maxVideos, channelID)

	for len(allVideos) < maxVideos {
		// Adjust limit for last page if needed
		remainingNeeded := maxVideos - len(allVideos)
		currentLimit := limit
		if currentLimit > remainingNeeded {
			currentLimit = remainingNeeded
		}

		// Fetch this page of videos
		videos, err := c.fetchVideosPage(channelID, offset, currentLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos at offset %d: %w", offset, err)
		}

		// If no videos returned, we've reached the end
		if len(videos) == 0 {
			break
		}

		allVideos = append(allVideos, videos...)

		// Store videos in smaller batches to avoid database locks
		if len(videos) > 0 {
			stored := 0
			for i := 0; i < len(videos); i += batchSize {
				end := i + batchSize
				if end > len(videos) {
					end = len(videos)
				}
				batch := videos[i:end]

				if err := StoreVideosBatch(channelID, batch); err != nil {
					fmt.Printf("⚠️  Warning: failed to store batch %d-%d: %v\n", i+1, end, err)
				} else {
					stored += len(batch)
				}
			}
			totalStored += stored
		}

		fmt.Printf("📥 Fetched %d videos (total: %d, stored: %d)\n", len(videos), len(allVideos), totalStored)

		// If we got fewer videos than requested, we've reached the end
		if len(videos) < currentLimit {
			break
		}

		// Move to next page
		offset += len(videos)
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("✅ Completed! Fetched %d total videos (%d stored) for channel %s in %v\n", len(allVideos), totalStored, channelID, elapsedTime)
	return allVideos, nil
}

// fetchVideosPage fetches a single page of videos (internal helper, bypasses cache)
func (c *Client) fetchVideosPage(channelID string, offset, limit int) ([]Video, error) {
	query := `
		query GetChannelVideos($id: String!, $limit: Float, $offset: Float) {
			getChannel(id: $id) {
				_id
				videos(limit: $limit, offset: $offset) {
					_id
					title
					summary
					largeImage
					videoDuration
					createdAt
					directUrl
					playCount
					likeCount
					angerCount
					embedUrl
					published
				}
			}
		}`

	variables := map[string]interface{}{
		"id":     channelID,
		"limit":  float64(limit),
		"offset": float64(offset),
	}

	request := GraphQLRequest{
		OperationName: "GetChannelVideos",
		Variables:     variables,
		Query:         query,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response GetChannelVideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.GetChannel.Videos, nil
}

// StoreVideosBatch stores videos in a batch with proper error handling
func StoreVideosBatch(channelID string, videos []Video) error {
	if len(videos) == 0 {
		return nil
	}

	// Use the existing StoreVideos function which handles transactions properly
	// This avoids nested transaction issues that were causing the SQLite errors
	return StoreVideos(channelID, videos)
}

// FetchChannelData fetches detailed channel information
func (c *Client) FetchChannelData(channelID string) error {
	query := `
		query GetChannel($id: String!) {
			getChannelByIdOrTitle(id: $id) {
				_id
				title
				summary
				textInfo
				avatar
				coverImage
				isLive
				showInfo {
					times
					phone
					__typename
				}
				links {
					website
					facebook
					twitter
					gab
					minds
					telegram
					subscribeStar
					__typename
				}
				featuredVideo {
					...DisplayVideoFields
					__typename
				}
				liveStreamVideo {
					...DisplayVideoFields
					streamUrl
					__typename
				}
				videos {
					_id
					__typename
				}
				playlists {
					_id
					title
					videos {
						_id
						__typename
					}
					__typename
				}
				__typename
			}
		}

		fragment DisplayVideoFields on Video {
			_id
			title
			summary
			playCount
			likeCount
			angerCount
			largeImage
			embedUrl
			published
			videoDuration
			channel {
				_id
				title
				avatar
				__typename
			}
			createdAt
			__typename
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
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response GetChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Store channel data in database
	channel := response.Data.GetChannelByIDOrTitle
	if err := StoreChannel(channel); err != nil {
		return fmt.Errorf("failed to store channel in database: %w", err)
	}

	// Also write response to file for backup/debugging
	outputData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := os.WriteFile("data/channel.json", outputData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FetchAllChannels fetches all available channels and stores them in database
func (c *Client) FetchAllChannels() error {
	query := `
		query {
			getAllChannels {
				_id
				title
				summary
				textInfo
				avatar
				coverImage
				isLive
				showInfo {
					times
					phone
					__typename
				}
				links {
					website
					facebook
					twitter
					gab
					minds
					telegram
					subscribeStar
					__typename
				}
				__typename
			}
		}`

	request := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response GetAllChannelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Store complete channel info in database
	for _, channel := range response.Data.GetAllChannels {
		if err := StoreChannel(channel); err != nil {
			fmt.Printf("Warning: failed to store channel %s in database: %v\n", channel.ID, err)
		}
	}

	// Write response to file for backup/debugging
	outputData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := os.WriteFile("data/all_channels.json", outputData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetAllChannelsFromDB retrieves all channels from the database
func GetAllChannelsFromDB() ([]Channel, error) {
	return GetAllChannels()
}
