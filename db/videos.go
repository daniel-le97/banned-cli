package db

import (
	"fmt"
)

// StoreVideos stores multiple videos in the database
func StoreVideos(channelID string, videos []Video) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO videos (
			id, channel_id, title, summary, large_image, video_duration,
			created_at_api, direct_url, play_count, like_count, anger_count,
			embed_url, published, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare video insert statement: %w", err)
	}
	defer stmt.Close()

	for _, video := range videos {
		_, err = stmt.Exec(
			video.ID, channelID, video.Title, video.Summary, video.LargeImage,
			video.VideoDuration, video.CreatedAt, video.DirectURL, video.PlayCount,
			video.LikeCount, video.AngerCount, video.EmbedURL, video.Published,
		)
		if err != nil {
			return fmt.Errorf("failed to store video '%s': %w", video.ID, err)
		}
	}

	LogDebug("Stored %d videos for channel %s", len(videos), channelID)
	return nil
}

// GetChannelVideos retrieves videos for a specific channel from the database
func GetChannelVideos(channelID string, limit, offset int) ([]Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, title, summary, large_image, video_duration, created_at_api,
			   direct_url, play_count, like_count, anger_count, embed_url, published, file_size
		FROM videos 
		WHERE channel_id = ? 
		ORDER BY created_at_api DESC 
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query videos: %w", err)
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var video Video
		err := rows.Scan(
			&video.ID, &video.Title, &video.Summary, &video.LargeImage,
			&video.VideoDuration, &video.CreatedAt, &video.DirectURL,
			&video.PlayCount, &video.LikeCount, &video.AngerCount,
			&video.EmbedURL, &video.Published, &video.FileSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading video rows: %w", err)
	}

	return videos, nil
}

// UpdateVideoFileSize updates the file_size for a video in the database
func UpdateVideoFileSize(videoID string, fileSize int64) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE videos SET file_size = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := DB.Exec(query, fileSize, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video file size: %w", err)
	}

	LogDebug("Updated video %s file size to %d bytes", videoID, fileSize)
	return nil
}
