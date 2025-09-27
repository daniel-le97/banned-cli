/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/tar"
	"archive/zip"
	"bufio"
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

	fmt.Printf("🔄 Replacing current binary...\n")

	// Replace the current binary
	if err := replaceBinary(binaryPath, execPath); err != nil {
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

func checkGoInstalled() error {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		// Go is not installed, prompt user for installation
		fmt.Println("Go is not installed or not available in PATH.")
		fmt.Println()

		if promptYesNo("Would you like me to help you install Go?") {
			return installGo()
		}

		return fmt.Errorf("go installation is required but was declined")
	}
	return nil
}

func promptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", question)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response == "y" || response == "yes" {
			return true
		}
		if response == "n" || response == "no" || response == "" {
			return false
		}

		fmt.Println("Please answer 'y' or 'n'")
	}
}

func installGo() error {
	fmt.Println("Attempting to install Go...")

	switch runtime.GOOS {
	case "linux":
		return installGoLinux()
	case "darwin":
		return installGoMacOS()
	case "windows":
		return installGoWindows()
	default:
		return showManualInstallInstructions()
	}
}

func installGoLinux() error {
	// Try different package managers
	packageManagers := []struct {
		cmd  string
		args []string
		name string
	}{
		{"apt", []string{"update", "&&", "apt", "install", "-y", "golang-go"}, "apt (Ubuntu/Debian)"},
		{"yum", []string{"install", "-y", "golang"}, "yum (RHEL/CentOS)"},
		{"dnf", []string{"install", "-y", "golang"}, "dnf (Fedora)"},
		{"pacman", []string{"-S", "--noconfirm", "go"}, "pacman (Arch)"},
		{"zypper", []string{"install", "-y", "go"}, "zypper (openSUSE)"},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.cmd); err == nil {
			fmt.Printf("Found %s, attempting installation...\n", pm.name)

			var cmd *exec.Cmd
			if pm.cmd == "apt" {
				// Special handling for apt update && install
				cmd = exec.Command("bash", "-c", "sudo apt update && sudo apt install -y golang-go")
			} else {
				args := append([]string{pm.cmd}, pm.args...)
				cmd = exec.Command("sudo", args...)
			}

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to install Go using %s: %v\n", pm.name, err)
				continue
			}

			// Verify installation
			if checkGoInstallation() {
				fmt.Println("Go installed successfully!")
				return nil
			}
		}
	}

	// If all package managers failed, try manual installation
	fmt.Println("Package manager installation failed. Trying manual installation...")
	return installGoManualLinux()
}

func installGoMacOS() error {
	// Try Homebrew first
	if _, err := exec.LookPath("brew"); err == nil {
		fmt.Println("Found Homebrew, attempting installation...")
		cmd := exec.Command("brew", "install", "go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err == nil && checkGoInstallation() {
			fmt.Println("Go installed successfully via Homebrew!")
			return nil
		}
	}

	// Fallback to manual installation
	return installGoManualMacOS()
}

func installGoWindows() error {
	// Try Chocolatey first
	if _, err := exec.LookPath("choco"); err == nil {
		fmt.Println("Found Chocolatey, attempting installation...")
		cmd := exec.Command("choco", "install", "golang", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err == nil && checkGoInstallation() {
			fmt.Println("Go installed successfully via Chocolatey!")
			return nil
		}
	}

	// Try Scoop
	if _, err := exec.LookPath("scoop"); err == nil {
		fmt.Println("Found Scoop, attempting installation...")
		cmd := exec.Command("scoop", "install", "go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err == nil && checkGoInstallation() {
			fmt.Println("Go installed successfully via Scoop!")
			return nil
		}
	}

	return showManualInstallInstructions()
}

func installGoManualLinux() error {
	fmt.Println("Downloading and installing Go manually...")

	// Get latest Go version (simplified - using a known recent version)
	goVersion := "1.21.5"
	arch := runtime.GOARCH

	downloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", goVersion, arch)

	// Download Go
	fmt.Printf("Downloading Go from %s...\n", downloadURL)
	cmd := exec.Command("wget", "-O", "/tmp/go.tar.gz", downloadURL)
	if err := cmd.Run(); err != nil {
		// Try curl if wget fails
		cmd = exec.Command("curl", "-L", "-o", "/tmp/go.tar.gz", downloadURL)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to download Go: %w", err)
		}
	}

	// Remove old Go installation
	exec.Command("sudo", "rm", "-rf", "/usr/local/go").Run()

	// Extract Go
	fmt.Println("Extracting Go...")
	cmd = exec.Command("sudo", "tar", "-C", "/usr/local", "-xzf", "/tmp/go.tar.gz")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract Go: %w", err)
	}

	// Add to PATH
	fmt.Println("Adding Go to PATH...")
	pathLine := "export PATH=$PATH:/usr/local/go/bin"

	// Try to add to .bashrc or .profile
	home := os.Getenv("HOME")
	profiles := []string{".bashrc", ".bash_profile", ".profile", ".zshrc"}

	for _, profile := range profiles {
		profilePath := filepath.Join(home, profile)
		if _, err := os.Stat(profilePath); err == nil {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("echo '%s' >> %s", pathLine, profilePath))
			cmd.Run()
			break
		}
	}

	fmt.Println("Go installed! Please run 'source ~/.bashrc' or restart your terminal.")
	fmt.Println("You can also run: export PATH=$PATH:/usr/local/go/bin")

	return nil
}

func installGoManualMacOS() error {
	fmt.Println("Please install Go manually:")
	fmt.Println("1. Go to https://golang.org/dl/")
	fmt.Println("2. Download the macOS installer (.pkg file)")
	fmt.Println("3. Run the installer")
	fmt.Println("4. Restart your terminal")

	return fmt.Errorf("manual installation required")
}

func checkGoInstallation() bool {
	cmd := exec.Command("go", "version")
	return cmd.Run() == nil
}

func showManualInstallInstructions() error {
	fmt.Println("\nPlease install Go manually:")
	fmt.Println("1. Visit: https://golang.org/dl/")
	fmt.Printf("2. Download the appropriate version for %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("3. Follow the installation instructions for your operating system")
	fmt.Println("4. Restart your terminal and try the update command again")
	fmt.Println()

	return fmt.Errorf("manual Go installation required")
}

func getGoBinPath() string {
	// Check GOBIN first
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin
	}

	// Check GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Default GOPATH
		home := os.Getenv("HOME")
		if home != "" {
			gopath = filepath.Join(home, "go")
		}
	}

	if gopath != "" {
		return filepath.Join(gopath, "bin")
	}

	// Fallback - this shouldn't happen in normal cases
	return "/usr/local/bin"
}

func isInPath(dir string) bool {
	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, string(os.PathListSeparator))

	for _, pathDir := range pathDirs {
		if pathDir == dir {
			return true
		}
		// Also check if they resolve to the same directory
		if abs1, err1 := filepath.Abs(pathDir); err1 == nil {
			if abs2, err2 := filepath.Abs(dir); err2 == nil {
				if abs1 == abs2 {
					return true
				}
			}
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags for the update command
	updateCmd.Flags().BoolP("check", "c", false, "Check for updates without installing")
	updateCmd.Flags().StringP("version", "v", "", "Update to a specific version")
}
