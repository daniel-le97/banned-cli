package db

import (
	"encoding/json"
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"
)

// VideoRecord represents a video record stored in the database
type VideoRecord struct {
	Video     Video  `json:"video"`
	ChannelID string `json:"channel_id"`
}

// StoreVideos stores multiple videos in the database
func StoreVideos(channelID string, videos []Video) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		for _, video := range videos {
			record := VideoRecord{
				Video:     video,
				ChannelID: channelID,
			}

			data, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("failed to marshal video '%s': %w", video.ID, err)
			}

			if err := bucket.Put([]byte(video.ID), data); err != nil {
				return fmt.Errorf("failed to store video '%s': %w", video.ID, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to store videos for channel '%s': %w", channelID, err)
	}

	return nil
}

// StoreVideo stores a single video in the database
func StoreVideo(channelID string, video Video) error {
	return StoreVideos(channelID, []Video{video})
}

// GetChannelVideos retrieves videos for a specific channel from the database
func GetChannelVideos(channelID string, limit, offset int) ([]Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var videos []Video
	count := 0

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return fmt.Errorf("failed to unmarshal video %s: %w", string(k), err)
			}

			// Filter by channel ID
			if record.ChannelID == channelID {
				// Apply offset
				if count < offset {
					count++
					return nil
				}

				// Apply limit
				if limit > 0 && len(videos) >= limit {
					return nil
				}

				videos = append(videos, record.Video)
			}

			count++
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get videos for channel '%s': %w", channelID, err)
	}

	return videos, nil
}

// GetVideo retrieves a single video by ID
func GetVideo(videoID string) (*Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var video Video

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		data := bucket.Get([]byte(videoID))
		if data == nil {
			return fmt.Errorf("video '%s' not found", videoID)
		}

		var record VideoRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return fmt.Errorf("failed to unmarshal video: %w", err)
		}

		video = record.Video
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &video, nil
}

// DeleteVideo removes a video from the database
func DeleteVideo(videoID string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		return bucket.Delete([]byte(videoID))
	})

	if err != nil {
		return fmt.Errorf("failed to delete video '%s': %w", videoID, err)
	}

	return nil
}

// VideoExists checks if a video exists in the database
func VideoExists(videoID string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}

	var exists bool

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		data := bucket.Get([]byte(videoID))
		exists = data != nil
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetAllVideos retrieves all videos from the database
func GetAllVideos() ([]Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var videos []Video

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return fmt.Errorf("failed to unmarshal video %s: %w", string(k), err)
			}
			videos = append(videos, record.Video)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all videos: %w", err)
	}

	return videos, nil
}

// SearchVideos searches for videos by title (simple substring search)
func SearchVideos(query string) ([]Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var videos []Video
	query = strings.ToLower(query)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return fmt.Errorf("failed to unmarshal video %s: %w", string(k), err)
			}

			// Simple case-insensitive substring search
			if strings.Contains(strings.ToLower(record.Video.Title), query) ||
				strings.Contains(strings.ToLower(record.Video.Summary), query) {
				videos = append(videos, record.Video)
			}

			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	return videos, nil
}

// UpdateVideoFileSize updates the file size for a specific video
func UpdateVideoFileSize(videoID string, fileSize int64) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		data := bucket.Get([]byte(videoID))
		if data == nil {
			return fmt.Errorf("video '%s' not found", videoID)
		}

		var record VideoRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return fmt.Errorf("failed to unmarshal video: %w", err)
		}

		// Update the file size
		record.Video.FileSize = fileSize

		// Marshal and save back
		updatedData, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal updated video: %w", err)
		}

		return bucket.Put([]byte(videoID), updatedData)
	})

	if err != nil {
		return fmt.Errorf("failed to update file size for video '%s': %w", videoID, err)
	}

	return nil
}

// GetVideosWithoutFileSize returns videos that don't have file size information
func GetVideosWithoutFileSize(channelID string) ([]Video, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var videos []Video

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return nil // Skip invalid records
			}

			// Filter by channel if specified
			if channelID != "" && record.ChannelID != channelID {
				return nil
			}

			// Check if video has no file size and has a direct URL
			if record.Video.FileSize == 0 && record.Video.DirectURL != "" {
				videos = append(videos, record.Video)
			}

			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get videos without file size: %w", err)
	}

	return videos, nil
}

// UpdateVideoFileSizeByURL updates the file size for a video by its direct URL
func UpdateVideoFileSizeByURL(directURL string, fileSize int64) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return fmt.Errorf("videos bucket not found")
		}

		// Find the video with the matching direct URL
		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return nil // Skip invalid records
			}

			if record.Video.DirectURL == directURL {
				// Update the file size
				record.Video.FileSize = fileSize

				// Marshal and save back
				updatedData, err := json.Marshal(record)
				if err != nil {
					return fmt.Errorf("failed to marshal updated video: %w", err)
				}

				return bucket.Put(k, updatedData)
			}

			return nil
		})
	})

	if err != nil {
		return fmt.Errorf("failed to update file size for video with URL '%s': %w", directURL, err)
	}

	return nil
}
