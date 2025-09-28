package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (set by goreleaser)
var (
	version = "v0.3.7"
	commit  = "none"
	date    = "unknown"
	builtBy = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for this application.

This command shows the current version, build commit, and build date
if the application was built with GoReleaser. Otherwise, it shows 
development version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func showVersion() {
	fmt.Printf("%s version %s\n", AppName, getVersionFromMain())

	// If built with GoReleaser, show additional info
	commit, date := getBuildInfo()
	if commit != "none" && commit != "" {
		fmt.Printf("Commit: %s\n", commit)
	}
	if date != "unknown" && date != "" {
		fmt.Printf("Built: %s\n", date)
	}

	fmt.Printf("Runtime: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func getVersionFromMain() string {
	return version
}

func getBuildInfo() (string, string) {
	return commit, date
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
