package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// GetSetting retrieves a setting value by key
func GetSetting(key string) (string, error) {
	db, err := GetDB()
	if err != nil {
		return "", err
	}

	var value string
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(SettingsBucket)
		if bucket == nil {
			return fmt.Errorf("settings bucket not found")
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("setting '%s' not found", key)
		}

		value = string(data)
		return nil
	})

	if err != nil {
		return "", err
	}

	return value, nil
}

// SetSetting updates or inserts a setting
func SetSetting(key, value string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(SettingsBucket)
		if bucket == nil {
			return fmt.Errorf("settings bucket not found")
		}

		return bucket.Put([]byte(key), []byte(value))
	})

	if err != nil {
		return fmt.Errorf("failed to set setting '%s': %w", key, err)
	}

	return nil
}
