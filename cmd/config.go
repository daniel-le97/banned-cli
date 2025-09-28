/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path"
	"strings"
)

// Application configuration - change these values to customize the application
const (
	// DefaultModulePath is the Go module path for updates (when published)
	DefaultModulePath = "github.com/daniel-le97/banned-cli"
	APIEndpoint	   = "https://api.banned.video/graphql"

	// AppDescription is the short description shown in help
	AppDescription = "tool for downloading from https://banned.video"

	// AppLongDescription is the detailed description
	AppLongDescription = "a tool for downloading from https://banned.video"

	// Default torrent configuration
	DefaultTrackerURL  = "udp://tracker.openbittorrent.com:80/announce"
	DefaultPieceLength = 256 // KB
)

// Popular public trackers (can be used as fallbacks)
var DefaultTrackers = []string{
	"udp://tracker.openbittorrent.com:80/announce",
	"udp://tracker.publicbt.com:80/announce",
	"udp://tracker.ccc.de:80/announce",
	"udp://exodus.desync.com:6969/announce",
}

var (
	// AppName is derived from the module name (part after last /)
	AppName = getAppNameFromModule(DefaultModulePath)

	// ConfigFileName is the default config file name (without extension)
	ConfigFileName = "." + AppName
)

// getAppNameFromModule extracts the application name from the module path
func getAppNameFromModule(modulePath string) string {
	// Remove any version suffix (e.g., @latest, @v1.0.0)
	if idx := strings.Index(modulePath, "@"); idx != -1 {
		modulePath = modulePath[:idx]
	}

	// Get the base name (part after last /)
	return path.Base(modulePath)
}
