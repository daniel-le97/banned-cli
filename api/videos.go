package api

import (
	"context"
	"fmt"
)

// VideoService provides methods for interacting with videos
type VideoService struct {
	client *Client
}

// NewVideoService creates a new video service
func NewVideoService(client *Client) *VideoService {
	return &VideoService{client: client}
}

// GetVideo fetches a specific video by ID
func (s *VideoService) GetVideo(ctx context.Context, videoID string) (*Video, error) {
	req := &GraphQLRequest{
		Query: GetVideoQuery,
		Variables: map[string]interface{}{
			"id": videoID,
		},
	}

	var result struct {
		GetVideo Video `json:"getVideo"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get video %s: %w", videoID, err)
	}

	return &result.GetVideo, nil
}

// GetVideos fetches multiple videos by their IDs
func (s *VideoService) GetVideos(ctx context.Context, videoIDs []string) ([]Video, error) {
	req := &GraphQLRequest{
		Query: GetVideosQuery,
		Variables: map[string]interface{}{
			"ids": videoIDs,
		},
	}

	var result struct {
		GetVideos []Video `json:"getVideos"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}

	return result.GetVideos, nil
}

// GetHotVideos fetches trending/hot videos
func (s *VideoService) GetHotVideos(ctx context.Context, offset, limit int) ([]Video, error) {
	req := &GraphQLRequest{
		Query: GetHotVideosQuery,
		Variables: map[string]interface{}{
			"offset": float64(offset),
			"limit":  float64(limit),
		},
	}

	var result struct {
		GetHotVideos []Video `json:"getHotVideos"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get hot videos: %w", err)
	}

	return result.GetHotVideos, nil
}

// GetNewVideos fetches the newest videos
func (s *VideoService) GetNewVideos(ctx context.Context, offset, limit int) ([]Video, error) {
	req := &GraphQLRequest{
		Query: GetNewVideosQuery,
		Variables: map[string]interface{}{
			"offset": float64(offset),
			"limit":  float64(limit),
		},
	}

	var result struct {
		GetNewVideos []Video `json:"getNewVideos"`
	}

	if err := s.client.ExecuteWithResult(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("failed to get new videos: %w", err)
	}

	return result.GetNewVideos, nil
}

// FetchAllHotVideos recursively fetches all hot videos with pagination
func (s *VideoService) FetchAllHotVideos(ctx context.Context, batchSize int) ([]Video, error) {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	var allVideos []Video
	offset := 0

	for {
		videos, err := s.GetHotVideos(ctx, offset, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch hot videos at offset %d: %w", offset, err)
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

// FetchAllNewVideos recursively fetches all new videos with pagination
func (s *VideoService) FetchAllNewVideos(ctx context.Context, batchSize int) ([]Video, error) {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	var allVideos []Video
	offset := 0

	for {
		videos, err := s.GetNewVideos(ctx, offset, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new videos at offset %d: %w", offset, err)
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

// FetchAllHotVideosWithLimit recursively fetches hot videos with a maximum limit
func (s *VideoService) FetchAllHotVideosWithLimit(ctx context.Context, maxVideos, batchSize int) ([]Video, error) {
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

		videos, err := s.GetHotVideos(ctx, offset, currentBatchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch hot videos at offset %d: %w", offset, err)
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

// GetVideoDownloadURL returns the best download URL for a video
func (s *VideoService) GetVideoDownloadURL(video *Video) string {
	// Note: The actual GraphQL schema doesn't expose direct URLs
	// This would need to be implemented based on the actual URL structure
	if video.VideoUpload != nil {
		// You might need to construct URLs based on name/key fields
		return video.VideoUpload.Name
	}
	return ""
}

// GetVideoPosterURL returns the poster/thumbnail URL for a video
func (s *VideoService) GetVideoPosterURL(video *Video) string {
	if video.PosterUpload != nil {
		// You might need to construct URLs based on name/key fields
		return video.PosterUpload.Name
	}
	return video.LargeImage
}

// GetVideoAudioURL returns the audio URL for a video (if available)
func (s *VideoService) GetVideoAudioURL(video *Video) string {
	if video.AudioUpload != nil {
		// You might need to construct URLs based on name/key fields
		return video.AudioUpload.Name
	}
	return ""
}
