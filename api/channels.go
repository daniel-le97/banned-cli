package api

import (
	"context"
	"fmt"
)

// ChannelService provides methods for interacting with channels
type ChannelService struct {
	client *Client
}

// NewChannelService creates a new channel service
func NewChannelService(client *Client) *ChannelService {
	return &ChannelService{client: client}
}

// GetAllChannels fetches all channels with pagination
func (s *ChannelService) GetAllChannels(ctx context.Context, offset, limit int) ([]Channel, error) {
	req := &GraphQLRequest{
		Query: GetAllChannelsQuery,
		Variables: map[string]interface{}{
			"offset": float64(offset),
			"limit":  float64(limit),
		},
	}

	var result struct {
		GetAllChannels []Channel `json:"getAllChannels"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get all channels: %w", err)
	}

	return result.GetAllChannels, nil
}

// GetChannel fetches a specific channel by ID
func (s *ChannelService) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	req := &GraphQLRequest{
		Query: GetChannelQuery,
		Variables: map[string]interface{}{
			"id": channelID,
		},
	}

	var result struct {
		GetChannel Channel `json:"getChannel"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get channel %s: %w", channelID, err)
	}

	return &result.GetChannel, nil
}

// GetChannelWithVideos fetches a channel with its videos
func (s *ChannelService) GetChannelWithVideos(ctx context.Context, params ChannelVideosParams) (*Channel, error) {
	req := &GraphQLRequest{
		Query: GetChannelWithVideosQuery,
		Variables: map[string]interface{}{
			"id":                 params.ChannelID,
			"includeUnlisted":    params.IncludeUnlisted,
			"includeUnpublished": params.IncludeUnpublished,
			"includeLive":        params.IncludeLive,
			"offset":             float64(params.Offset),
			"limit":              float64(params.Limit),
		},
	}

	var result struct {
		GetChannel Channel `json:"getChannel"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get channel with videos %s: %w", params.ChannelID, err)
	}

	return &result.GetChannel, nil
}

// GetChannelVideos fetches videos from a specific channel (convenience method)
func (s *ChannelService) GetChannelVideos(ctx context.Context, channelID string, offset, limit int) ([]Video, error) {
	params := ChannelVideosParams{
		ChannelID:          channelID,
		IncludeUnlisted:    false,
		IncludeUnpublished: false,
		IncludeLive:        true,
		Offset:             offset,
		Limit:              limit,
	}

	channel, err := s.GetChannelWithVideos(ctx, params)
	if err != nil {
		return nil, err
	}

	return channel.Videos, nil
}

// GetChannelHotVideos fetches hot/trending videos from a specific channel
func (s *ChannelService) GetChannelHotVideos(ctx context.Context, channelID string, offset, limit int) ([]Video, error) {
	req := &GraphQLRequest{
		Query: GetChannelHotVideosQuery,
		Variables: map[string]interface{}{
			"id":     channelID,
			"offset": float64(offset),
			"limit":  float64(limit),
		},
	}

	var result struct {
		GetChannel struct {
			HotVideos []Video `json:"hotVideos"`
		} `json:"getChannel"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get hot videos for channel %s: %w", channelID, err)
	}

	return result.GetChannel.HotVideos, nil
}

// FetchAllChannelVideos recursively fetches all videos from a channel with pagination
func (s *ChannelService) FetchAllChannelVideos(ctx context.Context, channelID string, batchSize int) ([]Video, error) {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	var allVideos []Video
	offset := 0

	for {
		videos, err := s.GetChannelVideos(ctx, channelID, offset, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos at offset %d: %w", offset, err)
		}

		if len(videos) == 0 {
			break // No more videos
		}

		allVideos = append(allVideos, videos...)
		offset += len(videos)

		// If we got fewer videos than requested, we've reached the end
		if len(videos) < batchSize {
			break
		}
	}

	return allVideos, nil
}

// FetchAllChannelVideosWithLimit recursively fetches videos with a maximum limit
func (s *ChannelService) FetchAllChannelVideosWithLimit(ctx context.Context, channelID string, maxVideos, batchSize int) ([]Video, error) {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	var allVideos []Video
	offset := 0

	for len(allVideos) < maxVideos {
		// Adjust batch size for the final request
		currentBatchSize := batchSize
		if len(allVideos)+currentBatchSize > maxVideos {
			currentBatchSize = maxVideos - len(allVideos)
		}

		videos, err := s.GetChannelVideos(ctx, channelID, offset, currentBatchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch videos at offset %d: %w", offset, err)
		}

		if len(videos) == 0 {
			break // No more videos
		}

		allVideos = append(allVideos, videos...)
		offset += len(videos)

		// If we got fewer videos than requested, we've reached the end
		if len(videos) < currentBatchSize {
			break
		}
	}

	return allVideos, nil
}
