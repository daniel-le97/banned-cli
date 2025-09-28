/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: fmt.Sprintf("Update %s to the latest version", AppName),
	Long: fmt.Sprintf(`Update %s to the latest version from GitHub releases.

This command downloads and installs the latest binary release for your platform
from the GitHub repository. It automatically detects your OS and architecture.

The update process will:
1. Check the latest version available on GitHub
2. Download the appropriate binary for your platform
3. Replace the current binary with the new version
4. Verify the update was successful

Examples:
  %s update              # Update to the latest version
  %s update --check      # Check for updates without installing
  %s update --version v1.2.3  # Update to a specific version`, AppName, AppName, AppName, AppName),
	Run: func(cmd *cobra.Command, args []string) {
		checkOnly, _ := cmd.Flags().GetBool("check")
		version, _ := cmd.Flags().GetString("version")

		if checkOnly {
			if err := checkForUpdates(); err != nil {
				fmt.Printf("Failed to check for updates: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if err := updateBinary(version); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// GitHub API structs
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	Body       string `json:"body"`
}

const (
	GitHubOwner = "daniel-le97"
	GitHubRepo  = "banned-cli"
)

// Version variables are defined in version.go and set by GoReleaser

func updateBinary(targetVersion string) error {
	fmt.Printf("🔍 Checking for %s updates...\n", AppName)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get latest release info
	var release *GitHubRelease
	if targetVersion == "" {
		release, err = getLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to get latest release: %w", err)
		}
	} else {
		release, err = getSpecificRelease(targetVersion)
		if err != nil {
			return fmt.Errorf("failed to get release %s: %w", targetVersion, err)
		}
	}

	fmt.Printf("📦 Latest version: %s\n", release.TagName)

	// Check if we're already on the latest version
	currentVersion := getCurrentVersion()
	if currentVersion != "" && currentVersion == release.TagName {
		fmt.Printf("✅ You're already running the latest version (%s)\n", currentVersion)
		return nil
	}

	// Find the appropriate asset for current OS/arch
	assetName := getAssetName(runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	var fileSize int64

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			downloadURL = asset.BrowserDownloadURL
			fileSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("📥 Downloading %s (%s)...\n", filepath.Base(downloadURL), formatBytes(fileSize))

	// Download the new binary
	tempDir, err := os.MkdirTemp("", "banned-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, filepath.Base(downloadURL))
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract the binary
	binaryPath := filepath.Join(tempDir, AppName)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	if err := extractBinary(archivePath, binaryPath); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Verify the binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("extracted binary not found: %w", err)
	}

	// Make it executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	fmt.Printf("🧪 Testing new binary...\n")

	// Test the new binary before replacing the old one
	if err := testBinary(binaryPath, release.TagName); err != nil {
		return fmt.Errorf("new binary verification failed: %w", err)
	}

	fmt.Printf("✅ New binary verified successfully!\n")
	fmt.Printf("🔄 Replacing current binary...\n")

	// Replace the current binary with verification
	if err := replaceBinaryWithVerification(binaryPath, execPath, release.TagName); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("✅ Successfully updated %s to version %s!\n", AppName, release.TagName)
	return nil
}

func checkForUpdates() error {
	fmt.Printf("🔍 Checking for %s updates...\n", AppName)

	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	currentVersion := getCurrentVersion()
	fmt.Printf("📦 Current version: %s\n", currentVersion)
	fmt.Printf("📦 Latest version: %s\n", release.TagName)

	if currentVersion != "" && currentVersion == release.TagName {
		fmt.Printf("✅ You're running the latest version!\n")
	} else {
		fmt.Printf("🆕 A new version is available!\n")
		fmt.Printf("Run '%s update' to install the latest version.\n", AppName)

		if release.Body != "" {
			fmt.Printf("\n📝 Release Notes:\n%s\n", release.Body)
		}
	}

	return nil
}

func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)
	return fetchRelease(url)
}

func getSpecificRelease(version string) (*GitHubRelease, error) {
	// Ensure version starts with 'v'
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", GitHubOwner, GitHubRepo, version)
	return fetchRelease(url)
}

func fetchRelease(url string) (*GitHubRelease, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Skip prereleases and drafts for latest release
	if release.Prerelease || release.Draft {
		return nil, fmt.Errorf("latest release is a prerelease or draft")
	}

	return &release, nil
}

func getCurrentVersion() string {
	// Return the version that was set by GoReleaser during build
	// If not set (development build), it will be "dev"
	return version
}

func getAssetName(goos, goarch string) string {
	// Convert Go arch names to GoReleaser names
	arch := goarch
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}

	// Convert Go OS names to GoReleaser names
	osName := strings.Title(goos)

	return fmt.Sprintf("%s_%s_%s", AppName, osName, arch)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func downloadFile(url, filepath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archivePath, targetPath string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractFromZip(archivePath, targetPath)
	} else if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractFromTarGz(archivePath, targetPath)
	}
	return fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractFromZip(zipPath, targetPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	binaryName := AppName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	for _, f := range r.File {
		if f.Name == binaryName || strings.HasSuffix(f.Name, "/"+binaryName) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in zip archive", binaryName)
}

func extractFromTarGz(tarGzPath, targetPath string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	binaryName := AppName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Name == binaryName || strings.HasSuffix(header.Name, "/"+binaryName) {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tr)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in tar.gz archive", binaryName)
}

// testBinary verifies that the new binary works correctly
func testBinary(binaryPath, expectedVersion string) error {
	// Test 1: Check if binary is executable and responds to version command
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("binary failed to execute: %w", err)
	}

	// Test 2: Check if version output contains expected version
	outputStr := string(output)
	if !strings.Contains(outputStr, expectedVersion) {
		return fmt.Errorf("version mismatch: expected %s, got %s", expectedVersion, outputStr)
	}

	// Test 3: Check if help command works (basic functionality test)
	cmd = exec.Command(binaryPath, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("binary help command failed: %w", err)
	}

	return nil
}

// replaceBinaryWithVerification replaces the binary with backup and rollback capability
func replaceBinaryWithVerification(newBinaryPath, currentBinaryPath, expectedVersion string) error {
	backupPath := currentBinaryPath + ".backup"

	// Step 1: Create backup of current binary
	if err := copyFile(currentBinaryPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Step 2: Replace with new binary
	if err := os.Rename(newBinaryPath, currentBinaryPath); err != nil {
		// Cleanup backup on failure
		os.Remove(backupPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Step 3: Test the installed binary
	fmt.Printf("🧪 Verifying installed binary...\n")
	if err := testBinary(currentBinaryPath, expectedVersion); err != nil {
		fmt.Printf("⚠️  Installed binary verification failed, rolling back...\n")

		// Rollback: restore backup
		if restoreErr := os.Rename(backupPath, currentBinaryPath); restoreErr != nil {
			return fmt.Errorf("verification failed AND rollback failed: %w (original error: %v)", restoreErr, err)
		}

		return fmt.Errorf("binary verification failed after installation, rolled back to previous version: %w", err)
	}

	// Step 4: Cleanup backup on success
	os.Remove(backupPath)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func replaceBinary(newBinaryPath, currentBinaryPath string) error {
	// On Windows, we might need to move the current binary before replacing
	if runtime.GOOS == "windows" {
		backupPath := currentBinaryPath + ".old"
		// Remove old backup if exists
		os.Remove(backupPath)

		// Move current binary to backup
		if err := os.Rename(currentBinaryPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}

		// Move new binary to current location
		if err := os.Rename(newBinaryPath, currentBinaryPath); err != nil {
			// Try to restore backup
			os.Rename(backupPath, currentBinaryPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}

		// Remove backup on success
		os.Remove(backupPath)
	} else {
		// On Unix-like systems, we can replace directly
		if err := os.Rename(newBinaryPath, currentBinaryPath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags for the update command
	updateCmd.Flags().BoolP("check", "c", false, "Check for updates without installing")
	updateCmd.Flags().StringP("version", "v", "", "Update to a specific version")
}
