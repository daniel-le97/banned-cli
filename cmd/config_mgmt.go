/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage application configuration",
	Long: `Manage application configuration settings including download directories,
torrent trackers, and other preferences. Settings are stored in the database
and persist across application restarts.

Examples:
  banned config list                          # Show all settings
  banned config set download_dir ~/Downloads  # Set download directory
  banned config get download_dir              # Get current download directory
  banned config reset download_dir            # Reset to default value`,
}

// configListCmd shows all current configuration settings
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long:  "Display all current configuration settings and their values",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current Configuration Settings:")
		fmt.Println("================================")

		settings := getConfigSettings()

		for _, setting := range settings {
			value, err := GetSetting(setting.Key)
			if err != nil || value == "" {
				value = setting.DefaultValue
			}
			fmt.Printf("%-25s: %s\n", setting.Key, value)
			if setting.Description != "" {
				fmt.Printf("%-25s  %s\n", "", setting.Description)
			}
			fmt.Println()
		}

		return nil
	},
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Set a configuration setting to a specific value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Validate the setting key
		setting := findConfigSetting(key)
		if setting == nil {
			return fmt.Errorf("unknown setting '%s'. Use 'config list' to see available settings", key)
		}

		// Validate the value
		if err := validateSettingValue(setting, value); err != nil {
			return fmt.Errorf("invalid value for '%s': %w", key, err)
		}

		// Set the value
		if err := SetSetting(key, value); err != nil {
			return fmt.Errorf("failed to set '%s': %w", key, err)
		}

		fmt.Printf("Successfully set %s = %s\n", key, value)
		return nil
	},
}

// configGetCmd gets a configuration value
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  "Get the current value of a configuration setting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Validate the setting key
		setting := findConfigSetting(key)
		if setting == nil {
			return fmt.Errorf("unknown setting '%s'. Use 'config list' to see available settings", key)
		}

		value, err := GetSetting(key)
		if err != nil || value == "" {
			value = setting.DefaultValue
			fmt.Printf("%s = %s (default)\n", key, value)
		} else {
			fmt.Printf("%s = %s\n", key, value)
		}

		return nil
	},
}

// configResetCmd resets a configuration value to its default
var configResetCmd = &cobra.Command{
	Use:   "reset <key>",
	Short: "Reset a configuration value to its default",
	Long:  "Reset a configuration setting to its default value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Validate the setting key
		setting := findConfigSetting(key)
		if setting == nil {
			return fmt.Errorf("unknown setting '%s'. Use 'config list' to see available settings", key)
		}

		// Reset to default by setting the default value
		if err := SetSetting(key, setting.DefaultValue); err != nil {
			return fmt.Errorf("failed to reset '%s': %w", key, err)
		}

		fmt.Printf("Successfully reset %s = %s (default)\n", key, setting.DefaultValue)
		return nil
	},
}

// ConfigSetting represents a configuration setting with metadata
type ConfigSetting struct {
	Key          string
	DefaultValue string
	Description  string
	Validator    func(string) error
}

// getConfigSettings returns all available configuration settings
func getConfigSettings() []ConfigSetting {
	homeDir, _ := os.UserHomeDir()
	defaultDownloadDir := filepath.Join(homeDir, "Downloads", "banned")

	return []ConfigSetting{
		{
			Key:          "download_dir",
			DefaultValue: defaultDownloadDir,
			Description:  "Directory where downloaded files are saved",
			Validator:    validateDirectory,
		},
		{
			Key:          "max_concurrent_downloads",
			DefaultValue: "3",
			Description:  "Maximum number of concurrent downloads (1-10)",
			Validator:    validateConcurrentDownloads,
		},
		{
			Key:          "retry_attempts",
			DefaultValue: "3",
			Description:  "Number of retry attempts for failed downloads (0-10)",
			Validator:    validateRetryAttempts,
		},
		{
			Key:          "user_agent",
			DefaultValue: "banned-cli/1.0",
			Description:  "User agent string for HTTP requests",
			Validator:    validateUserAgent,
		},
		{
			Key:          "torrent_trackers",
			DefaultValue: strings.Join(DefaultTrackers, ","),
			Description:  "Comma-separated list of torrent tracker URLs",
			Validator:    validateTrackers,
		},
		{
			Key:          "torrent_piece_length",
			DefaultValue: strconv.Itoa(DefaultPieceLength),
			Description:  "Torrent piece length in KB (64-1024)",
			Validator:    validatePieceLength,
		},
	}
}

// findConfigSetting finds a setting by key
func findConfigSetting(key string) *ConfigSetting {
	settings := getConfigSettings()
	for i := range settings {
		if settings[i].Key == key {
			return &settings[i]
		}
	}
	return nil
}

// validateSettingValue validates a setting value using its validator
func validateSettingValue(setting *ConfigSetting, value string) error {
	if setting.Validator != nil {
		return setting.Validator(value)
	}
	return nil
}

// Validator functions

func validateDirectory(path string) error {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot expand home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Check if path is absolute or make it absolute
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot get current directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	// Try to create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	return nil
}

func validateConcurrentDownloads(value string) error {
	num, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if num < 1 || num > 10 {
		return fmt.Errorf("must be between 1 and 10")
	}
	return nil
}

func validateRetryAttempts(value string) error {
	num, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if num < 0 || num > 10 {
		return fmt.Errorf("must be between 0 and 10")
	}
	return nil
}

func validateUserAgent(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("cannot be empty")
	}
	if len(value) > 200 {
		return fmt.Errorf("must be less than 200 characters")
	}
	return nil
}

func validateTrackers(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("cannot be empty")
	}

	trackers := strings.Split(value, ",")
	for _, tracker := range trackers {
		tracker = strings.TrimSpace(tracker)
		if tracker == "" {
			continue
		}

		// Basic URL validation
		if !strings.HasPrefix(tracker, "http://") && !strings.HasPrefix(tracker, "https://") && !strings.HasPrefix(tracker, "udp://") {
			return fmt.Errorf("tracker URLs must start with http://, https://, or udp://: %s", tracker)
		}
	}

	return nil
}

func validatePieceLength(value string) error {
	num, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if num < 64 || num > 1024 {
		return fmt.Errorf("must be between 64 and 1024 KB")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configResetCmd)
}
