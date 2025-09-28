/*
Copyright © 2025 NAME ExaExamples:
  %s install           # Install using symlink method
  %s install --profile # Force install by modifying shell profile", AppName, AppName),es:

	%s app install           # Install using symlink method
	%s app install --profile # Force install by modifying shell profile", AppName, AppName, AppName),E <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: fmt.Sprintf("Install %s to PATH", AppName),
	Long: fmt.Sprintf(`Install %s by either creating a symlink in a PATH directory
or by adding the current binary location to your shell's PATH configuration.

The install command will:
1. First try to create a symlink in /usr/local/bin (requires sudo on most systems)
2. If that fails, it will add the binary location to your shell's profile (.bashrc, .zshrc, etc.)

Examples:
  %s install           # Install using symlink method
  %s install --profile # Force install by modifying shell profile`, AppName, AppName, AppName),
	Run: func(cmd *cobra.Command, args []string) {
		forceProfile, _ := cmd.Flags().GetBool("profile")

		if err := installBinary(forceProfile); err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func installBinary(forceProfile bool) error {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve any symlinks to get the actual binary path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	binaryName := AppName

	// Try symlink method first unless forcing profile method
	if !forceProfile {
		if err := installViaSymlink(execPath, binaryName); err == nil {
			return nil
		}
		fmt.Printf("Symlink installation failed, trying profile method...\n")
	}

	// Fall back to or force profile method
	return installViaProfile(execPath, binaryName)
}

func installViaSymlink(execPath, binaryName string) error {
	// Common PATH directories to try
	pathDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		filepath.Join(os.Getenv("HOME"), ".local/bin"),
		filepath.Join(os.Getenv("HOME"), "bin"),
	}

	for _, dir := range pathDirs {
		if !isDirectoryInPath(dir) {
			continue
		}

		symlinkPath := filepath.Join(dir, binaryName)

		// Check if directory exists and is writable
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				continue
			}
		}

		// Remove existing symlink/file if it exists
		if _, err := os.Lstat(symlinkPath); err == nil {
			os.Remove(symlinkPath)
		}

		// Create symlink
		if err := os.Symlink(execPath, symlinkPath); err != nil {
			continue
		}

		fmt.Printf("Successfully installed %s to %s\n", binaryName, symlinkPath)
		fmt.Printf("You can now run '%s' from anywhere in your terminal.\n", binaryName)
		return nil
	}

	return fmt.Errorf("failed to create symlink in any PATH directory")
}

func installViaProfile(execPath, binaryName string) error {
	binDir := filepath.Dir(execPath)

	// Detect shell and profile file
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash" // default
	}

	var profileFile string
	shellName := filepath.Base(shell)

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return fmt.Errorf("HOME environment variable not set")
	}

	switch shellName {
	case "zsh":
		profileFile = filepath.Join(homeDir, ".zshrc")
	case "fish":
		// Fish uses a different config structure
		configDir := filepath.Join(homeDir, ".config/fish/conf.d")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create fish config directory: %w", err)
		}
		profileFile = filepath.Join(configDir, AppName+".fish")
		return addToFishConfig(profileFile, binDir, binaryName)
	case "bash":
		fallthrough
	default:
		// Try .bashrc first, then .bash_profile, then .profile
		candidates := []string{".bashrc", ".bash_profile", ".profile"}
		for _, candidate := range candidates {
			path := filepath.Join(homeDir, candidate)
			if _, err := os.Stat(path); err == nil {
				profileFile = path
				break
			}
		}
		if profileFile == "" {
			profileFile = filepath.Join(homeDir, ".bashrc")
		}
	}

	return addToBashProfile(profileFile, binDir, binaryName)
}

func addToBashProfile(profileFile, binDir, binaryName string) error {
	// Check if the directory is already in PATH
	if isDirectoryInPath(binDir) {
		fmt.Printf("%s is already in your PATH.\n", binaryName)
		return nil
	}

	// Prepare the export statement
	exportLine := fmt.Sprintf("# Added by %s installer\nexport PATH=\"%s:$PATH\"", binaryName, binDir)

	// Read existing profile content
	var content []byte
	var err error
	if _, err = os.Stat(profileFile); err == nil {
		content, err = os.ReadFile(profileFile)
		if err != nil {
			return fmt.Errorf("failed to read profile file: %w", err)
		}
	}

	// Check if already added
	contentStr := string(content)
	if strings.Contains(contentStr, fmt.Sprintf("# Added by %s installer", binaryName)) {
		fmt.Printf("%s is already configured in %s\n", binaryName, profileFile)
		return nil
	}

	// Append the export line
	if len(content) > 0 && !strings.HasSuffix(contentStr, "\n") {
		exportLine = "\n" + exportLine
	}
	exportLine += "\n"

	// Write to profile file
	f, err := os.OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to profile file: %w", err)
	}

	fmt.Printf("Successfully added %s to PATH in %s\n", binaryName, profileFile)
	fmt.Printf("Please restart your terminal or run 'source %s' to use the new PATH.\n", profileFile)
	fmt.Printf("You can then run '%s' from anywhere in your terminal.\n", binaryName)

	return nil
}

func addToFishConfig(configFile, binDir, binaryName string) error {
	// Check if the directory is already in PATH
	if isDirectoryInPath(binDir) {
		fmt.Printf("%s is already in your PATH.\n", binaryName)
		return nil
	}

	fishConfig := fmt.Sprintf("# Added by %s installer\nset -gx PATH %s $PATH\n", binaryName, binDir)

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read fish config: %w", err)
		}
		if strings.Contains(string(content), fmt.Sprintf("# Added by %s installer", binaryName)) {
			fmt.Printf("%s is already configured in fish\n", binaryName)
			return nil
		}
	}

	// Write fish config
	if err := os.WriteFile(configFile, []byte(fishConfig), 0644); err != nil {
		return fmt.Errorf("failed to write fish config: %w", err)
	}

	fmt.Printf("Successfully added %s to PATH in fish configuration\n", binaryName)
	fmt.Printf("Please restart your terminal to use the new PATH.\n")
	fmt.Printf("You can then run '%s' from anywhere in your terminal.\n", binaryName)

	return nil
}

func isDirectoryInPath(dir string) bool {
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
	rootCmd.AddCommand(installCmd)

	// Add flags for the install command
	installCmd.Flags().BoolP("profile", "p", false, "Force installation by modifying shell profile instead of creating symlink")
}
