/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Status: Failed (%v)\n", err)
			return
		}

		fmt.Printf("✅ Status: Connected\n\n")

		// Show table statistics
		tables := []string{"channels", "videos", "downloads", "settings"}
		fmt.Println("📋 Table Statistics:")
		for _, table := range tables {
			var count int
			err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err != nil {
				fmt.Printf("   %s: Error (%v)\n", table, err)
			} else {
				fmt.Printf("   %s: %d records\n", table, count)
			}
		}

		fmt.Println("\n💡 Use 'banned config list' to view application settings")
	},
}

// dbViewCmd shows database tables in an interactive TUI
var dbViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Interactive database table viewer",
	Long: `View database tables and records in an interactive Bubble Tea interface.

Navigation:
  ←/→ or Tab - Switch between tables
  ↑/↓        - Navigate rows  
  q          - Quit

Tables available: Downloads, Settings, Channels, Videos`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔗 Connecting to database...")
		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Failed to connect to database: %v\n", err)
			return
		}

		fmt.Println("📊 Starting database viewer...")
		fmt.Println("💡 Use ←/→ or Tab to switch tables, ↑/↓ to navigate, 'q' to quit")

		model := newDatabaseViewModel(db)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("❌ Error running database viewer: %v\n", err)
		}

		fmt.Println("👋 Database viewer closed")
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
	dbCmd.AddCommand(dbEditorHtmxCmd)
	dbEditorHtmxCmd.Flags().IntP("port", "p", 8080, "Port for the web server")
}
