package db

import (
	"fmt"
)

// createTables creates all necessary database tables
func createTables() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Downloads table to track download history
	downloadsTable := `
	CREATE TABLE IF NOT EXISTS downloads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		title TEXT,
		filename TEXT,
		file_path TEXT,
		file_size INTEGER,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		error_message TEXT,
		torrent_created BOOLEAN DEFAULT FALSE,
		torrent_path TEXT,
		torrent_created_at DATETIME
	);`

	// Settings table for application configuration
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Channels table for storing channel information
	channelsTable := `
	CREATE TABLE IF NOT EXISTS channels (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		summary TEXT,
		text_info TEXT,
		avatar TEXT,
		cover_image TEXT,
		is_live BOOLEAN DEFAULT FALSE,
		total_videos INTEGER DEFAULT 0,
		total_video_views INTEGER DEFAULT 0,
		total_likes INTEGER DEFAULT 0,
		show_times TEXT,
		show_phone TEXT,
		website TEXT,
		facebook TEXT,
		twitter TEXT,
		gab TEXT,
		minds TEXT,
		telegram TEXT,
		subscribe_star TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Videos table for storing video information
	videosTable := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		channel_id TEXT NOT NULL,
		title TEXT NOT NULL,
		summary TEXT,
		large_image TEXT,
		video_duration REAL,
		created_at_api TEXT,
		direct_url TEXT,
		play_count INTEGER DEFAULT 0,
		like_count INTEGER DEFAULT 0,
		anger_count INTEGER DEFAULT 0,
		embed_url TEXT,
		published BOOLEAN DEFAULT TRUE,
		downloaded BOOLEAN DEFAULT FALSE,
		file_path TEXT,
		file_size INTEGER DEFAULT 0,
		torrent_created BOOLEAN DEFAULT FALSE,
		torrent_path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (channel_id) REFERENCES channels(id)
	);`

	// Execute table creation
	if _, err := DB.Exec(downloadsTable); err != nil {
		return fmt.Errorf("failed to create downloads table: %w", err)
	}

	if _, err := DB.Exec(settingsTable); err != nil {
		return fmt.Errorf("failed to create settings table: %w", err)
	}

	if _, err := DB.Exec(channelsTable); err != nil {
		return fmt.Errorf("failed to create channels table: %w", err)
	}

	if _, err := DB.Exec(videosTable); err != nil {
		return fmt.Errorf("failed to create videos table: %w", err)
	}

	// Insert default settings if they don't exist
	if err := insertDefaultSettings(); err != nil {
		return fmt.Errorf("failed to insert default settings: %w", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations handles database schema migrations
func runMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check if file_size column exists, if not, add it
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('videos') WHERE name='file_size'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for file_size column: %w", err)
	}

	if count == 0 {
		_, err := DB.Exec("ALTER TABLE videos ADD COLUMN file_size INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add file_size column: %w", err)
		}
	}

	// Check if total_videos column exists in channels table, if not, add it
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('channels') WHERE name='total_videos'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for total_videos column: %w", err)
	}

	if count == 0 {
		_, err := DB.Exec("ALTER TABLE channels ADD COLUMN total_videos INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add total_videos column: %w", err)
		}
	}

	// Check if total_video_views column exists in channels table, if not, add it
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('channels') WHERE name='total_video_views'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for total_video_views column: %w", err)
	}

	if count == 0 {
		_, err := DB.Exec("ALTER TABLE channels ADD COLUMN total_video_views INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add total_video_views column: %w", err)
		}
	}

	// Check if total_likes column exists in channels table, if not, add it
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('channels') WHERE name='total_likes'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for total_likes column: %w", err)
	}

	if count == 0 {
		_, err := DB.Exec("ALTER TABLE channels ADD COLUMN total_likes INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add total_likes column: %w", err)
		}
	}

	return nil
}

// insertDefaultSettings inserts default application settings
func insertDefaultSettings() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Default settings to insert
	defaultSettings := map[string]string{
		"download_directory": "~/Downloads/banned/",
		"last_sync":          "2023-01-01T00:00:00Z",
		"api_endpoint":       "https://api.banned.video/graphql",
		"user_agent":         "banned/1.0",
	}

	// Insert settings only if they don't exist
	for key, value := range defaultSettings {
		var exists int
		err := DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = ?", key).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if setting exists: %w", err)
		}

		if exists == 0 {
			_, err = DB.Exec("INSERT INTO settings (key, value) VALUES (?, ?)", key, value)
			if err != nil {
				return fmt.Errorf("failed to insert default setting %s: %w", key, err)
			}
		}
	}

	return nil
}
