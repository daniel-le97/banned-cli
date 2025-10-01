package db

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
		return filepath.Join(currentDir, "banned.db"), nil
	}

	// Get user's data directory
	userDataDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user directories: %w", err)
		}
		userDataDir = homeDir
	}

	// Create the application data directory
	appDataDir := filepath.Join(userDataDir, "banned")
	return filepath.Join(appDataDir, "banned.db"), nil
}

// isDevelopmentMode detects if we're running in development mode
func isDevelopmentMode() bool {
	// Check if we're in a git repository (common development indicator)
	if _, err := os.Stat(".git"); err == nil {
		return true
	}

	// Check if main.go exists in current directory
	if _, err := os.Stat("main.go"); err == nil {
		return true
	}

	// Check if go.mod exists in current directory
	if _, err := os.Stat("go.mod"); err == nil {
		return true
	}

	return false
}

// copyBundledDatabase attempts to copy a bundled database to the target path
func copyBundledDatabase(dbPath string) error {
	// Look for bundled database in several locations
	possibleSources := []string{
		"banned.db",                   // Current directory
		"data/banned.db",              // Data directory
		"assets/banned.db",            // Assets directory
		"/usr/share/banned/banned.db", // System installation
	}

	for _, src := range possibleSources {
		if _, err := os.Stat(src); err == nil {
			// Source exists, try to copy it
			if err := copyFile(src, dbPath); err == nil {
				return nil // Success
			}
		}
	}

	return fmt.Errorf("no bundled database found in any of the expected locations")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDatabasePath returns the path where the database is stored
func GetDatabasePath() (string, error) {
	return getDatabasePath()
}

// LogDebug prints debug messages when in development mode
func LogDebug(message string, args ...interface{}) {
	if isDevelopmentMode() {
		log.Printf("[DEBUG] "+message, args...)
	}
}
