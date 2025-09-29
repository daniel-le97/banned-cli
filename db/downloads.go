package db

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Download represents a download record
type Download struct {
	ID               string    `json:"id"`
	URL              string    `json:"url"`
	Title            string    `json:"title"`
	Filename         string    `json:"filename"`
	FilePath         string    `json:"file_path"`
	FileSize         int64     `json:"file_size"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	TorrentCreated   bool      `json:"torrent_created"`
	TorrentPath      string    `json:"torrent_path,omitempty"`
	TorrentCreatedAt *time.Time `json:"torrent_created_at,omitempty"`
}

// StoreDownload stores a download record in the database
func StoreDownload(download Download) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	data, err := json.Marshal(download)
	if err != nil {
		return fmt.Errorf("failed to marshal download: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(DownloadsBucket)
		if bucket == nil {
			return fmt.Errorf("downloads bucket not found")
		}

		return bucket.Put([]byte(download.ID), data)
	})

	if err != nil {
		return fmt.Errorf("failed to store download: %w", err)
	}

	return nil
}

// UpdateDownloadTorrentInfo updates torrent information for a download by file path
func UpdateDownloadTorrentInfo(filePath, torrentPath string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	now := time.Now()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(DownloadsBucket)
		if bucket == nil {
			return fmt.Errorf("downloads bucket not found")
		}

		// Find the download with matching file path
		return bucket.ForEach(func(k, v []byte) error {
			var download Download
			if err := json.Unmarshal(v, &download); err != nil {
				return nil // Skip invalid records
			}

			if download.FilePath == filePath {
				// Update torrent information
				download.TorrentCreated = true
				download.TorrentPath = torrentPath
				download.TorrentCreatedAt = &now

				// Marshal and save back
				updatedData, err := json.Marshal(download)
				if err != nil {
					return fmt.Errorf("failed to marshal updated download: %w", err)
				}

				return bucket.Put(k, updatedData)
			}

			return nil
		})
	})

	if err != nil {
		return fmt.Errorf("failed to update download torrent info: %w", err)
	}

	return nil
}

// GetDownload retrieves a download by ID
func GetDownload(downloadID string) (*Download, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var download Download

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(DownloadsBucket)
		if bucket == nil {
			return fmt.Errorf("downloads bucket not found")
		}

		data := bucket.Get([]byte(downloadID))
		if data == nil {
			return fmt.Errorf("download '%s' not found", downloadID)
		}

		return json.Unmarshal(data, &download)
	})

	if err != nil {
		return nil, err
	}

	return &download, nil
}

// GetAllDownloads retrieves all download records
func GetAllDownloads() ([]Download, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var downloads []Download

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(DownloadsBucket)
		if bucket == nil {
			return fmt.Errorf("downloads bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var download Download
			if err := json.Unmarshal(v, &download); err != nil {
				return fmt.Errorf("failed to unmarshal download %s: %w", string(k), err)
			}
			downloads = append(downloads, download)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all downloads: %w", err)
	}

	return downloads, nil
}