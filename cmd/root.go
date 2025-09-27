/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppDescription,
	Long:  AppLongDescription,
	Run: func(cmd *cobra.Command, args []string) {
		interactive, _ := cmd.Flags().GetBool("interactive")

		if interactive || len(args) == 0 {
			// Run interactive TUI when no subcommands are provided or --interactive flag is used
			if err := RunTUI(); err != nil {
				fmt.Printf("Error running interactive mode: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Show help if arguments are provided but no valid command
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s.yaml)", ConfigFileName))

	// Cobra also supports local flags which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("interactive", "i", false, "Run in interactive mode")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
