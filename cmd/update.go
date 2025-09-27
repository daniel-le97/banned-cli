/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: fmt.Sprintf("Update %s to the latest version", AppName),
	Long: fmt.Sprintf(`Update %s to the latest version from the source repository.

This command uses 'go install' to fetch and build the latest version of %s
from the source repository. It requires Go to be installed on your system.

The update process will:
1. Check if Go is available on your system
2. Use 'go install' to fetch and build the latest version
3. Update the binary in your GOPATH/bin or GOBIN directory

Examples:
  %s app update              # Update to the latest version
  %s app update --source URL # Update from a specific repository URL`, AppName, AppName, AppName, AppName),
	Run: func(cmd *cobra.Command, args []string) {
		sourceURL, _ := cmd.Flags().GetString("source")

		if err := updateBinary(sourceURL); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func updateBinary(sourceURL string) error {
	// Check if Go is installed
	if err := checkGoInstalled(); err != nil {
		return err
	}

	// Determine the module path to use for go install
	modulePath := sourceURL
	if modulePath == "" {
		// Try to detect if we're using a binary that was installed via go install
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		// Resolve any symlinks to get the actual binary path
		execPath, err = filepath.EvalSymlinks(execPath)
		if err != nil {
			return fmt.Errorf("failed to resolve executable path: %w", err)
		}

		goBinPath := getGoBinPath()
		execDir := filepath.Dir(execPath)

		// Check if the executable is in the same directory as go.mod (local development)
		if _, err := os.Stat(filepath.Join(execDir, "go.mod")); err == nil {
			// Executable is in the source directory, suggest local rebuild
			return fmt.Errorf("detected local development environment. To update:\n"+
				"1. For latest changes: run 'go build -o %s' to rebuild locally\n"+
				"2. For remote updates: first publish to a Git repository, then use:\n"+
				"   %s update --source github.com/yourusername/%s@latest", AppName, AppName, AppName)
		}

		if strings.HasPrefix(execPath, goBinPath) {
			// Binary is in GOBIN/GOPATH, likely installed via go install
			// Default to the current module path (when published)
			modulePath = DefaultModulePath + "@latest"
		} else {
			// Binary is elsewhere, probably local build or symlinked
			return fmt.Errorf("this appears to be a local build or custom installation. To update:\n"+
				"1. If this is a local build: go build -o %s\n"+
				"2. If installed from a repository: %s update --source github.com/user/repo@latest\n"+
				"3. Binary location: %s", AppName, AppName, execPath)
		}
	}

	fmt.Printf("Updating %s from %s...\n", AppName, modulePath)

	// Run go install command
	cmd := exec.Command("go", "install", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run 'go install %s': %w", modulePath, err)
	}

	// Check where the binary was installed
	goBin := getGoBinPath()
	binaryPath := filepath.Join(goBin, AppName)

	// Verify the installation
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("update completed but binary not found at expected location %s: %w", binaryPath, err)
	}

	fmt.Printf("Successfully updated %s!\n", AppName)
	fmt.Printf("Binary location: %s\n", binaryPath)

	// Check if the binary location is in PATH
	if isInPath(goBin) {
		fmt.Printf("You can now run '%s' from anywhere.\n", AppName)
	} else {
		fmt.Printf("Note: %s is not in your PATH. You may need to add it or use the full path.\n", goBin)
		fmt.Printf("To add to PATH, run: export PATH=\"%s:$PATH\"\n", goBin)
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
	appCmd.AddCommand(updateCmd)

	// Add flags for the update command
	updateCmd.Flags().StringP("source", "s", "", "Specify a custom source repository URL (e.g., github.com/user/repo@latest)")
}
