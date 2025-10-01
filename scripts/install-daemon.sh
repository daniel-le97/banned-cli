#!/bin/bash
set -e

# Banned CLI Daemon Installation Script

DAEMON_USER="banned"
DAEMON_GROUP="banned" 
SERVICE_NAME="banned-daemon"
BINARY_PATH="/usr/local/bin/banned"
LOG_DIR="/var/log/banned"
RUN_DIR="/var/run/banned"
DATA_DIR="/home/${DAEMON_USER}/.local/share/banned"

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

install_daemon() {
    info "Installing banned-cli daemon..."
    
    # Create system user
    if ! id "$DAEMON_USER" &>/dev/null; then
        info "Creating system user: $DAEMON_USER"
        useradd --system --home-dir "/home/$DAEMON_USER" --create-home --shell /bin/bash "$DAEMON_USER"
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
    cp "scripts/systemd/${SERVICE_NAME}.service" "/etc/systemd/system/"
    
    # Reload systemd and enable service
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    
    success "Daemon installed successfully!"
    
    echo
    echo "📋 Next steps:"
    echo "  1. Start the daemon: sudo systemctl start $SERVICE_NAME"
    echo "  2. Check status: sudo systemctl status $SERVICE_NAME"
    echo "  3. View logs: sudo journalctl -u $SERVICE_NAME -f"
    echo "  4. Web UI will be available at: http://localhost:8080"
    echo
    echo "🔧 Configuration:"
    echo "  • Edit service file: /etc/systemd/system/${SERVICE_NAME}.service"
    echo "  • Logs directory: $LOG_DIR"
    echo "  • Data directory: $DATA_DIR"
    echo "  • Run as user: $DAEMON_USER"
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
    echo "🚀 Banned CLI Daemon Manager"
    echo "============================"
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