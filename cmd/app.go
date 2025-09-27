/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// appCmd represents the app command for application management
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application management commands",
	Long: `Application management commands for ` + AppName + `.

This command provides utilities for managing the application itself:
- Install the application to your system PATH
- Update the application to the latest version`,
}

func init() {
	rootCmd.AddCommand(appCmd)
}
