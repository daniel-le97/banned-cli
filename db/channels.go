package db

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// StoreChannel stores a channel in the database
func StoreChannel(channel Channel) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// Serialize channel to JSON
	data, err := json.Marshal(channel)
	if err != nil {
		return fmt.Errorf("failed to marshal channel '%s': %w", channel.ID, err)
	}

	// Store in bbolt
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return fmt.Errorf("channels bucket not found")
		}

		return bucket.Put([]byte(channel.ID), data)
	})

	if err != nil {
		return fmt.Errorf("failed to store channel '%s': %w", channel.ID, err)
	}

	return nil
}

// GetAllChannels retrieves all channels from the database
func GetAllChannels() ([]Channel, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var channels []Channel

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return fmt.Errorf("channels bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var channel Channel
			if err := json.Unmarshal(v, &channel); err != nil {
				return fmt.Errorf("failed to unmarshal channel %s: %w", string(k), err)
			}
			channels = append(channels, channel)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all channels: %w", err)
	}

	return channels, nil
}

// GetChannel retrieves a single channel by ID
func GetChannel(channelID string) (*Channel, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var channel Channel

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return fmt.Errorf("channels bucket not found")
		}

		data := bucket.Get([]byte(channelID))
		if data == nil {
			return fmt.Errorf("channel '%s' not found", channelID)
		}

		return json.Unmarshal(data, &channel)
	})

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// DeleteChannel removes a channel from the database
func DeleteChannel(channelID string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return fmt.Errorf("channels bucket not found")
		}

		return bucket.Delete([]byte(channelID))
	})

	if err != nil {
		return fmt.Errorf("failed to delete channel '%s': %w", channelID, err)
	}

	return nil
}

// ChannelExists checks if a channel exists in the database
func ChannelExists(channelID string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}

	var exists bool

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ChannelsBucket)
		if bucket == nil {
			return fmt.Errorf("channels bucket not found")
		}

		data := bucket.Get([]byte(channelID))
		exists = data != nil
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}