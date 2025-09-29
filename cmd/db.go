/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dbCmd represents the db command for database operations
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database status and tools",
	Long: `Database status and interactive tools for ` + AppName + `.

Available commands:
- status: Show database statistics and connection info
- view:   Interactive database table viewer`,
}

// dbStatusCmd shows database status
var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database status and statistics",
	Long:  `Show database location, connection status, and table record counts.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📊 %s Database Status\n", AppName)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Show database location
		dbPath := getDatabaseLocationSafe()
		fmt.Printf("📍 Location: %s\n", dbPath)

		// Try to connect
		_, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Status: Failed (%v)\n", err)
			return
		}

		fmt.Printf("✅ Status: Connected\n\n")

		// Show bucket statistics
		stats, err := GetBucketStats()
		if err != nil {
			fmt.Printf("❌ Failed to get bucket statistics: %v\n", err)
			return
		}

		fmt.Println("📋 Bucket Statistics:")
		buckets := []string{"channels", "videos", "downloads", "settings"}
		for _, bucket := range buckets {
			count := stats[bucket]
			fmt.Printf("   %s: %d records\n", bucket, count)
		}

		fmt.Println("\n💡 Use 'banned config list' to view application settings")
	},
}

// dbViewCmd shows database tables in an interactive TUI
var dbViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Interactive database table viewer",
	Long: `View database tables and records in an interactive Bubble Tea interface.

Note: Database viewer is temporarily disabled during bbolt migration.
Use 'db status' to view bucket statistics instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("� Database viewer is temporarily disabled during bbolt migration")
		fmt.Println("📊 Use 'banned db status' to view bucket statistics")
		fmt.Println("💡 Use 'banned config list' to view application settings")
	},
}

func getDatabaseLocationSafe() string {
	path, err := getDatabasePath()
	if err != nil {
		return fmt.Sprintf("unknown (%v)", err)
	}
	return path
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbViewCmd)
}
