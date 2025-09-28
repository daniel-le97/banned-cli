package cmd

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	// Global database instance
	DB *sql.DB

	// Ensure database is initialized only once
	dbOnce sync.Once

	// Database initialization error
	dbInitError error
)

// GetDB returns the global database instance, initializing it if necessary
func GetDB() (*sql.DB, error) {
	dbOnce.Do(func() {
		var err error
		DB, err = initDatabase()
		if err != nil {
			dbInitError = fmt.Errorf("failed to initialize database: %w", err)
			return
		}

		// Create tables if they don't exist
		if err := createTables(); err != nil {
			dbInitError = fmt.Errorf("failed to create tables: %w", err)
			return
		}
	})

	if dbInitError != nil {
		return nil, dbInitError
	}

	return DB, nil
}

// initDatabase initializes the SQLite database connection
func initDatabase() (*sql.DB, error) {
	// Get the database path
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}

	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if database exists, if not try to copy bundled database
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := copyBundledDatabase(dbPath); err != nil {
			// If copying bundled database fails, continue with creating empty database
			fmt.Printf("📦 Could not copy bundled database (%v), creating empty database...\n", err)
		} else {
			fmt.Printf("📦 Initialized with bundled database containing sample data!\n")
		}
	}

	// Open database connection
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	return db, nil
}

// getDatabasePath returns the path where the database should be stored
func getDatabasePath() (string, error) {
	// For development: use current directory if DEV environment variable is set
	// or if we can detect we're in development mode
	if os.Getenv("DEV") != "" || os.Getenv("DEVELOPMENT") != "" || isDevelopmentMode() {
		// Use current directory for development
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		dbPath := filepath.Join(currentDir, AppName+".db")
		return dbPath, nil
	}

	// Production: use XDG_DATA_HOME first, then fallback to home directory
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, ".local", "share")
	}

	appDataDir := filepath.Join(dataDir, AppName)
	dbPath := filepath.Join(appDataDir, AppName+".db")

	return dbPath, nil
}

// isDevelopmentMode detects if we're running in development mode
func isDevelopmentMode() bool {
	// Check if we're in a directory that looks like a development environment
	currentDir, err := os.Getwd()
	if err != nil {
		return false
	}

	// Look for development indicators
	devIndicators := []string{"go.mod", "main.go", ".git", "cmd/"}
	for _, indicator := range devIndicators {
		if _, err := os.Stat(filepath.Join(currentDir, indicator)); err == nil {
			return true
		}
	}

	return false
}

// copyBundledDatabase copies the bundled database.db file to the user's data directory
func copyBundledDatabase(dbPath string) error {
	// Look for bundled database in possible locations
	bundledPaths := []string{
		"database.db",                            // Same directory as executable
		"./database.db",                          // Current directory
		filepath.Join(os.Args[0], "database.db"), // Next to executable (for packaged releases)
	}

	// Get the directory containing the executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		bundledPaths = append(bundledPaths, filepath.Join(execDir, "database.db"))
	}

	var bundledDB string
	for _, path := range bundledPaths {
		if _, err := os.Stat(path); err == nil {
			bundledDB = path
			break
		}
	}

	if bundledDB == "" {
		return fmt.Errorf("no bundled database found in expected locations")
	}

	// Copy the bundled database to the user location
	return copyFile(bundledDB, dbPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return destFile.Sync()
}

// createTables creates the necessary database tables
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

	return nil
}

// insertDefaultSettings inserts default application settings
func insertDefaultSettings() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	defaultSettings := map[string]string{
		"download_dir":             filepath.Join(os.Getenv("HOME"), "Downloads", AppName),
		"max_concurrent_downloads": "3",
		"retry_attempts":           "3",
		"user_agent":               fmt.Sprintf("%s/1.0", AppName),
	}

	stmt, err := DB.Prepare(`
		INSERT OR IGNORE INTO settings (key, value) 
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range defaultSettings {
		if _, err := stmt.Exec(key, value); err != nil {
			return fmt.Errorf("failed to insert setting %s: %w", key, err)
		}
	}

	return nil
}

// CloseDB closes the global database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Helper functions for database operations

// GetSetting retrieves a setting value by key
func GetSetting(key string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}

	var value string
	err = db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("setting '%s' not found", key)
		}
		return "", fmt.Errorf("failed to get setting '%s': %w", key, err)
	}

	return value, nil
}

// SetSetting updates or inserts a setting
func SetSetting(key, value string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO settings (key, value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, key, value)

	if err != nil {
		return fmt.Errorf("failed to set setting '%s': %w", key, err)
	}

	return nil
}

// LogDebug logs database operations for debugging
func LogDebug(message string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Printf("[DB] "+message, args...)
	}
}

// StoreChannel stores or updates channel information in the database
func StoreChannel(channel Channel) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	showTimes := ""
	showPhone := ""
	if channel.ShowInfo != nil {
		showTimes = channel.ShowInfo.Times
		showPhone = channel.ShowInfo.Phone
	}

	website, facebook, twitter, gab, minds, telegram, subscribeStar := "", "", "", "", "", "", ""
	if channel.Links != nil {
		website = channel.Links.Website
		facebook = channel.Links.Facebook
		twitter = channel.Links.Twitter
		gab = channel.Links.Gab
		minds = channel.Links.Minds
		telegram = channel.Links.Telegram
		subscribeStar = channel.Links.SubscribeStar
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO channels (
			id, title, summary, text_info, avatar, cover_image, is_live,
			total_videos, total_video_views, total_likes,
			show_times, show_phone, website, facebook, twitter, gab, minds,
			telegram, subscribe_star, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, channel.ID, channel.Title, channel.Summary, channel.TextInfo, channel.Avatar,
		channel.CoverImage, channel.IsLive, int(channel.TotalVideos), 0, 0,
		showTimes, showPhone, website, facebook, twitter, gab, minds, telegram, subscribeStar)

	if err != nil {
		return fmt.Errorf("failed to store channel '%s': %w", channel.ID, err)
	}

	LogDebug("Stored channel: %s (%s)", channel.ID, channel.Title)
	return nil
}

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

// GetAllChannels retrieves all channels from the database
func GetAllChannels() ([]Channel, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, title, summary, text_info, avatar, cover_image, is_live,
			   COALESCE(total_videos, 0), COALESCE(total_video_views, 0), COALESCE(total_likes, 0),
			   show_times, show_phone, website, facebook, twitter, gab, minds,
			   telegram, subscribe_star
		FROM channels 
		ORDER BY title
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var channel Channel
		var totalVideos, totalVideoViews, totalLikes int
		var showTimes, showPhone, website, facebook, twitter, gab, minds, telegram, subscribeStar sql.NullString

		err := rows.Scan(
			&channel.ID, &channel.Title, &channel.Summary, &channel.TextInfo,
			&channel.Avatar, &channel.CoverImage, &channel.IsLive,
			&totalVideos, &totalVideoViews, &totalLikes,
			&showTimes, &showPhone, &website, &facebook, &twitter, &gab, &minds,
			&telegram, &subscribeStar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}

		// Assign the video counts
		channel.TotalVideos = float64(totalVideos)
		channel.TotalVideoViews = float64(totalVideoViews)
		channel.TotalLikes = float64(totalLikes)

		// Populate nested structs if data exists
		if showTimes.Valid || showPhone.Valid {
			channel.ShowInfo = &ShowInfo{
				Times: showTimes.String,
				Phone: showPhone.String,
			}
		}

		if website.Valid || facebook.Valid || twitter.Valid || gab.Valid || minds.Valid || telegram.Valid || subscribeStar.Valid {
			channel.Links = &Links{
				Website:       website.String,
				Facebook:      facebook.String,
				Twitter:       twitter.String,
				Gab:           gab.String,
				Minds:         minds.String,
				Telegram:      telegram.String,
				SubscribeStar: subscribeStar.String,
			}
		}

		channels = append(channels, channel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading channel rows: %w", err)
	}

	return channels, nil
}
