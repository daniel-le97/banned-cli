package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

type InstallDaemonConfig struct {
	Generate     bool
	Execute      bool
	OutputPath   string
	ServiceUser  string
	ServiceGroup string
	Port         string
	AutoSync     bool
	SyncInterval string
	LogDir       string
	RunDir       string
	DataDir      string
}

var installDaemonConfig = InstallDaemonConfig{
	Generate:     false,
	Execute:      false,
	OutputPath:   "./install-daemon.sh",
	ServiceUser:  "banned",
	ServiceGroup: "banned",
	Port:         "8080",
	AutoSync:     true,
	SyncInterval: "1h",
	LogDir:       "/var/log/banned",
	RunDir:       "/var/run/banned",
	DataDir:      "/home/banned/.local/share/banned",
}

// installDaemonCmd represents the install-daemon command
var installDaemonCmd = &cobra.Command{
	Use:   "install-daemon",
	Short: "Generate or install daemon service scripts",
	Long: `Generate systemd service files and installation scripts for running banned-cli as a daemon.

This command can:
- Generate installation script only (--generate)
- Generate and execute installation (--execute) 
- Create systemd service files
- Set up proper users, directories, and permissions

Examples:
  banned install-daemon --generate                     # Generate script only
  banned install-daemon --execute                      # Generate and run (requires sudo)
  banned install-daemon --generate --port 3000         # Custom port
  banned install-daemon --output /tmp/install.sh       # Custom output path`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("daemon installation is only supported on Linux and macOS")
		}

		if installDaemonConfig.Generate || installDaemonConfig.Execute {
			if err := generateInstallScript(); err != nil {
				return fmt.Errorf("failed to generate install script: %w", err)
			}

			fmt.Printf("✅ Generated installation script: %s\n", installDaemonConfig.OutputPath)

			if installDaemonConfig.Execute {
				return executeInstallScript()
			} else {
				fmt.Printf("\n💡 To install the daemon, run:\n")
				fmt.Printf("   sudo bash %s install\n", installDaemonConfig.OutputPath)
			}
		} else {
			return fmt.Errorf("specify either --generate or --execute")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installDaemonCmd)

	// Generation flags
	installDaemonCmd.Flags().BoolVar(&installDaemonConfig.Generate, "generate", false, "Generate installation script")
	installDaemonCmd.Flags().BoolVar(&installDaemonConfig.Execute, "execute", false, "Generate and execute installation (requires sudo)")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.OutputPath, "output", "./install-daemon.sh", "Output path for installation script")

	// Service configuration
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.ServiceUser, "user", "banned", "System user for daemon service")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.ServiceGroup, "group", "banned", "System group for daemon service")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.Port, "port", "8080", "Port for web UI service")
	installDaemonCmd.Flags().BoolVar(&installDaemonConfig.AutoSync, "auto-sync", true, "Enable automatic syncing")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.SyncInterval, "interval", "1h", "Auto-sync interval")

	// Directory configuration
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.LogDir, "log-dir", "/var/log/banned", "Log directory")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.RunDir, "run-dir", "/var/run/banned", "Runtime directory")
	installDaemonCmd.Flags().StringVar(&installDaemonConfig.DataDir, "data-dir", "/home/banned/.local/share/banned", "Data directory")
}

func generateInstallScript() error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(installDaemonConfig.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate systemd service content
	serviceContent := generateSystemdService()

	// Generate install script content
	scriptContent := generateInstallScriptContent()

	// Write install script
	file, err := os.Create(installDaemonConfig.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create script file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(scriptContent); err != nil {
		return fmt.Errorf("failed to write script content: %w", err)
	}

	// Make script executable
	if err := os.Chmod(installDaemonConfig.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Write systemd service file alongside the script
	serviceFile := filepath.Join(outputDir, "banned-daemon.service")
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	fmt.Printf("📄 Generated systemd service: %s\n", serviceFile)

	return nil
}

func generateSystemdService() string {
	daemonArgs := fmt.Sprintf("--webui --port %s", installDaemonConfig.Port)
	if installDaemonConfig.AutoSync {
		daemonArgs += fmt.Sprintf(" --auto-sync --interval %s", installDaemonConfig.SyncInterval)
	}
	daemonArgs += fmt.Sprintf(" --log-file %s/daemon.log --pid-file %s/banned.pid",
		installDaemonConfig.LogDir, installDaemonConfig.RunDir)

	return fmt.Sprintf(`[Unit]
Description=Banned CLI Daemon
Documentation=https://github.com/daniel-le97/banned-cli
After=network.target
Wants=network.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=HOME=%s

# Daemon configuration
ExecStart=/usr/local/bin/banned daemon %s
ExecReload=/bin/kill -HUP $MAINPID

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=%s %s %s

# Restart policy
Restart=always
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# Resource limits
MemoryLimit=512M
TasksMax=50

# Logging
StandardOutput=append:%s/daemon.log
StandardError=append:%s/daemon.log

[Install]
WantedBy=multi-user.target
`, installDaemonConfig.ServiceUser, installDaemonConfig.ServiceGroup,
		filepath.Dir(installDaemonConfig.DataDir), filepath.Dir(installDaemonConfig.DataDir),
		daemonArgs, installDaemonConfig.LogDir, installDaemonConfig.RunDir, installDaemonConfig.DataDir,
		installDaemonConfig.LogDir, installDaemonConfig.LogDir)
}

func generateInstallScriptContent() string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Generated Banned CLI Daemon Installation Script
# Created by: banned install-daemon --generate

DAEMON_USER="%s"
DAEMON_GROUP="%s"
SERVICE_NAME="banned-daemon"
BINARY_PATH="/usr/local/bin/banned"
LOG_DIR="%s"
RUN_DIR="%s"
DATA_DIR="%s"
PORT="%s"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✅${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}❌${NC} $1"
    exit 1
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root. Use: sudo $0"
    fi
}

check_binary() {
    if [ ! -f "$BINARY_PATH" ]; then
        error "banned binary not found at $BINARY_PATH. Please install banned-cli first."
    fi
}

install_daemon() {
    info "Installing banned-cli daemon..."
    
    check_binary
    
    # Create system user
    if ! id "$DAEMON_USER" &>/dev/null; then
        info "Creating system user: $DAEMON_USER"
        useradd --system --home-dir "$(dirname "$DATA_DIR")" --create-home --shell /bin/bash "$DAEMON_USER"
        success "Created system user: $DAEMON_USER"
    else
        info "System user $DAEMON_USER already exists"
    fi
    
    # Create directories
    info "Creating directories..."
    mkdir -p "$LOG_DIR" "$RUN_DIR" "$DATA_DIR"
    
    # Set permissions
    chown "$DAEMON_USER:$DAEMON_GROUP" "$LOG_DIR" "$RUN_DIR" "$DATA_DIR"
    chmod 755 "$LOG_DIR" "$RUN_DIR" "$DATA_DIR"
    
    # Install systemd service
    info "Installing systemd service..."
    cp "banned-daemon.service" "/etc/systemd/system/"
    
    # Reload systemd and enable service
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    
    success "Daemon installed successfully!"
    
    echo
    echo "📋 Next steps:"
    echo "  1. Start the daemon: sudo systemctl start $SERVICE_NAME"
    echo "  2. Check status: sudo systemctl status $SERVICE_NAME"  
    echo "  3. View logs: sudo journalctl -u $SERVICE_NAME -f"
    echo "  4. PWA Web UI: http://localhost:$PORT"
    echo
    echo "🔧 Configuration:"
    echo "  • Service file: /etc/systemd/system/${SERVICE_NAME}.service"
    echo "  • Logs: $LOG_DIR"
    echo "  • Data: $DATA_DIR"
    echo "  • User: $DAEMON_USER"
}

uninstall_daemon() {
    info "Uninstalling banned-cli daemon..."
    
    # Stop and disable service
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        systemctl stop "$SERVICE_NAME"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        systemctl disable "$SERVICE_NAME"
    fi
    
    # Remove service file
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
    
    # Optionally remove user and directories
    read -p "Remove system user '$DAEMON_USER' and data directories? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        userdel -r "$DAEMON_USER" 2>/dev/null || true
        rm -rf "$LOG_DIR" "$RUN_DIR"
        success "Removed user and directories"
    fi
    
    success "Daemon uninstalled successfully!"
}

show_status() {
    info "Banned CLI Daemon Status"
    echo "========================"
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        success "Service is running"
        echo "🌐 PWA Web UI: http://localhost:$PORT"
    else
        warn "Service is not running"
    fi
    
    echo
    systemctl status "$SERVICE_NAME" --no-pager || true
    
    echo
    info "Recent logs:"
    journalctl -u "$SERVICE_NAME" --no-pager -n 10 || true
}

main() {
    echo "🚀 Banned CLI Daemon Installer"
    echo "Generated with: banned install-daemon --generate"
    echo "==============================="
    echo
    
    case "${1:-}" in
        install)
            check_root
            install_daemon
            ;;
        uninstall)
            check_root
            uninstall_daemon
            ;;
        status)
            show_status
            ;;
        *)
            echo "Usage: $0 {install|uninstall|status}"
            echo
            echo "Commands:"
            echo "  install    Install daemon service (requires root)"
            echo "  uninstall  Remove daemon service (requires root)"
            echo "  status     Show daemon status"
            exit 1
            ;;
    esac
}

main "$@"
`, installDaemonConfig.ServiceUser, installDaemonConfig.ServiceGroup,
		installDaemonConfig.LogDir, installDaemonConfig.RunDir, installDaemonConfig.DataDir,
		installDaemonConfig.Port)
}

func executeInstallScript() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("installation requires root privileges. Run: sudo banned install-daemon --execute")
	}

	// Execute the install script
	fmt.Printf("🚀 Executing installation script...\n\n")

	cmd := exec.Command("bash", installDaemonConfig.OutputPath, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// Helper function to check if a command exists
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
