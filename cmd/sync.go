/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command (disabled during bbolt migration)
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync new data from banned.video API (temporarily disabled)",
	Long: `Sync functionality is temporarily disabled during the bbolt database migration.
Use 'fetch' commands instead for now.

Alternative commands:
- banned fetch channels     # Fetch all channels
- banned fetch videos <id>  # Fetch videos for a channel`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚧 Sync functionality is temporarily disabled during bbolt migration")
		fmt.Println("📡 Use 'banned fetch channels' to get channel data")
		fmt.Println("📹 Use 'banned fetch videos <channel-id>' to get video data")
		fmt.Println("💡 Use 'banned --help' to see available commands")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	// All sync subcommands are temporarily disabled during bbolt migration
}