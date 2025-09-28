package db

import (
	"database/sql"
	"fmt"
)

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
