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
	Short: "Database operations and status",
	Long: `Database operations and status for ` + AppName + `.

This command allows you to:
- Initialize the database
- Check database status  
- View database location
- Test database connectivity`,
}

// dbInitCmd initializes the database
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Initialize the database and create all necessary tables.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Initializing %s database...\n", AppName)

		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Failed to initialize database: %v\n", err)
			return
		}

		fmt.Printf("✅ Database initialized successfully!\n")
		fmt.Printf("📍 Location: %s\n", getDatabaseLocationSafe())

		// Test a simple query
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM downloads").Scan(&count)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not query downloads table: %v\n", err)
		} else {
			fmt.Printf("📊 Downloads in database: %d\n", count)
		}
	},
}

// dbStatusCmd shows database status
var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database status",
	Long:  `Show the current status of the database connection and statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📊 %s Database Status\n", AppName)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Show database location
		dbPath := getDatabaseLocationSafe()
		fmt.Printf("📍 Database Location: %s\n", dbPath)

		// Try to connect
		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Connection Status: Failed (%v)\n", err)
			return
		}

		fmt.Printf("✅ Connection Status: Connected\n")

		// Show table statistics
		tables := []string{"downloads", "settings", "channels", "videos"}
		for _, table := range tables {
			var count int
			err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err != nil {
				fmt.Printf("⚠️  %s table: Error (%v)\n", table, err)
			} else {
				fmt.Printf("📋 %s table: %d records\n", table, count)
			}
		}
	},
}

// dbSettingsCmd shows current settings
var dbSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show current database settings",
	Long:  `Display all current settings stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⚙️  %s Settings\n", AppName)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		db, err := GetDB()
		if err != nil {
			fmt.Printf("❌ Failed to connect to database: %v\n", err)
			return
		}

		rows, err := db.Query("SELECT key, value, updated_at FROM settings ORDER BY key")
		if err != nil {
			fmt.Printf("❌ Failed to query settings: %v\n", err)
			return
		}
		defer rows.Close()

		fmt.Println()
		for rows.Next() {
			var key, value, updatedAt string
			if err := rows.Scan(&key, &value, &updatedAt); err != nil {
				fmt.Printf("⚠️  Error reading setting: %v\n", err)
				continue
			}
			fmt.Printf("🔧 %s: %s (updated: %s)\n", key, value, updatedAt)
		}

		if err = rows.Err(); err != nil {
			fmt.Printf("⚠️  Error iterating settings: %v\n", err)
		}
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
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbSettingsCmd)
	dbCmd.AddCommand(dbViewCmd)
}
