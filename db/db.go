package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	// Global bbolt database instance
	DB *bolt.DB

	// Ensure database is initialized only once
	dbOnce sync.Once

	// Database initialization error
	dbInitError error
)

// Bucket names for different data types
var (
	ChannelsBucket  = []byte("channels")
	VideosBucket    = []byte("videos")
	DownloadsBucket = []byte("downloads")
	SettingsBucket  = []byte("settings")
)

// GetDB returns the global database instance, initializing it if necessary
func GetDB() (*bolt.DB, error) {
	dbOnce.Do(func() {
		var err error
		DB, err = initDatabase()
		if err != nil {
			dbInitError = fmt.Errorf("failed to initialize database: %w", err)
			return
		}

		// Create buckets if they don't exist
		if err := createBuckets(); err != nil {
			dbInitError = fmt.Errorf("failed to create buckets: %w", err)
			return
		}

		// Insert default settings
		if err := insertDefaultSettings(); err != nil {
			dbInitError = fmt.Errorf("failed to insert default settings: %w", err)
			return
		}
	})

	if dbInitError != nil {
		return nil, dbInitError
	}

	return DB, nil
}

// initDatabase initializes the bbolt database connection
func initDatabase() (*bolt.DB, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open bbolt database
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

// createBuckets creates all necessary buckets
func createBuckets() error {
	return DB.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			ChannelsBucket,
			VideosBucket,
			DownloadsBucket,
			SettingsBucket,
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", string(bucket), err)
			}
		}

		return nil
	})
}

// getDatabasePath returns the path where the database should be stored
func getDatabasePath() (string, error) {
	// Try to get user config directory first
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		dbPath := filepath.Join(userConfigDir, "banned", "banned.db")
		return dbPath, nil
	}

	// Fallback to user data directory
	userDataDir, err := getUserDataDir()
	if err == nil {
		dbPath := filepath.Join(userDataDir, "banned", "banned.db")
		return dbPath, nil
	}

	// Final fallback to current directory
	return "banned.db", nil
}

// getUserDataDir returns the user data directory
func getUserDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use appropriate data directory for the OS
	switch {
	case os.Getenv("XDG_DATA_HOME") != "":
		return os.Getenv("XDG_DATA_HOME"), nil
	default:
		return filepath.Join(homeDir, ".local", "share"), nil
	}
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetBucketStats returns statistics for all buckets
func GetBucketStats() (map[string]int, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)

	err = db.View(func(tx *bolt.Tx) error {
		buckets := map[string][]byte{
			"channels":  ChannelsBucket,
			"videos":    VideosBucket,
			"downloads": DownloadsBucket,
			"settings":  SettingsBucket,
		}

		for name, bucketName := range buckets {
			bucket := tx.Bucket(bucketName)
			if bucket == nil {
				stats[name] = 0
				continue
			}

			count := 0
			bucket.ForEach(func(k, v []byte) error {
				count++
				return nil
			})
			stats[name] = count
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get bucket stats: %w", err)
	}

	return stats, nil
}

// CountChannels returns the number of channels in the database
func CountChannels() (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	var count int
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
		return nil
	})

	return count, err
}

// CountVideos returns the number of videos in the database
func CountVideos() (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	var count int
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
		return nil
	})

	return count, err
}

// CountVideosForChannel returns the number of videos for a specific channel
func CountVideosForChannel(channelID string) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	var count int
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(VideosBucket)
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record VideoRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return nil // Skip invalid records
			}

			if record.ChannelID == channelID {
				count++
			}
			return nil
		})
	})

	return count, err
}

// GetDatabasePath returns the current database path (for compatibility)
func GetDatabasePath() (string, error) {
	return getDatabasePath()
}

// insertDefaultSettings inserts default application settings
func insertDefaultSettings() error {
	return DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(SettingsBucket)
		if bucket == nil {
			return fmt.Errorf("settings bucket not found")
		}

		// Default settings to insert
		defaultSettings := map[string]string{
			"download_directory": "~/Downloads/banned/",
			"last_sync":          "2023-01-01T00:00:00Z",
			"api_endpoint":       "https://api.banned.video/graphql",
			"user_agent":         "banned-cli/1.0",
		}

		// Insert settings only if they don't exist
		for key, value := range defaultSettings {
			existing := bucket.Get([]byte(key))
			if existing == nil {
				if err := bucket.Put([]byte(key), []byte(value)); err != nil {
					return fmt.Errorf("failed to insert default setting %s: %w", key, err)
				}
			}
		}

		return nil
	})
}