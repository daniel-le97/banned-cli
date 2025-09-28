#!/bin/bash
set -e

# banned CLI installer script
# Usage: curl -sSL https://github.com/daniel-le97/banned-cli/releases/latest/download/install.sh | bash

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="banned"
REPO="daniel-le97/banned-cli"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="Linux" ;;
        Darwin*)    os="Darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="Windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64" ;;
        i386|i686)      arch="i386" ;;
        aarch64|arm64)  arch="arm64" ;;
        armv7l)         arch="armv7" ;;
        armv6l)         arch="armv6" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        info "Fetching latest version..."
        local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
        
        if command -v curl >/dev/null 2>&1; then
            VERSION=$(curl -s "$latest_url" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
        elif command -v wget >/dev/null 2>&1; then
            VERSION=$(wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
        else
            error "Either curl or wget is required to download the installer"
        fi
        
        if [ -z "$VERSION" ]; then
            error "Failed to get latest version"
        fi
    fi
    
    info "Installing version: $VERSION"
}

# Download and install binary
install_binary() {
    local platform="$1"
    local tmpdir=$(mktemp -d)
    local archive_name="${BINARY_NAME}_${platform}"
    
    # Determine file extension
    if [[ "$platform" == *"Windows"* ]]; then
        archive_name="${archive_name}.zip"
    else
        archive_name="${archive_name}.tar.gz"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${archive_name}"
    local archive_path="${tmpdir}/${archive_name}"
    
    info "Downloading from: $download_url"
    
    # Download the archive
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$archive_path" "$download_url" || error "Failed to download $archive_name"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$archive_path" "$download_url" || error "Failed to download $archive_name"
    else
        error "Either curl or wget is required to download the installer"
    fi
    
    # Extract the archive
    if [[ "$archive_name" == *.tar.gz ]]; then
        tar -xzf "$archive_path" -C "$tmpdir" || error "Failed to extract $archive_name"
    elif [[ "$archive_name" == *.zip ]]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$archive_path" -d "$tmpdir" || error "Failed to extract $archive_name"
        else
            error "unzip is required to extract Windows archives"
        fi
    fi
    
    # Find the binary
    local binary_path
    if [[ "$platform" == *"Windows"* ]]; then
        binary_path="${tmpdir}/${BINARY_NAME}.exe"
    else
        binary_path="${tmpdir}/${BINARY_NAME}"
    fi
    
    if [ ! -f "$binary_path" ]; then
        error "Binary not found in archive"
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        warn "Install directory $INSTALL_DIR doesn't exist, creating it..."
        if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
            error "Failed to create install directory. Try running with sudo or set INSTALL_DIR to a writable location."
        fi
    fi
    
    # Install the binary
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [[ "$platform" == *"Windows"* ]]; then
        install_path="${install_path}.exe"
    fi
    
    info "Installing to: $install_path"
    
    if ! cp "$binary_path" "$install_path" 2>/dev/null; then
        error "Failed to install binary. Try running with sudo or set INSTALL_DIR to a writable location."
    fi
    
    # Clean up
    rm -rf "$tmpdir"
    
    success "Successfully installed $BINARY_NAME $VERSION to $install_path"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        success "$BINARY_NAME is available in PATH"
        info "Run '$BINARY_NAME --help' to get started"
        
        # Initialize database if it's the first install
        info "Initializing database..."
        if "$BINARY_NAME" db init >/dev/null 2>&1; then
            success "Database initialized successfully"
        else
            warn "Database initialization failed or already exists"
        fi
    else
        warn "$BINARY_NAME is not in PATH. You may need to add $INSTALL_DIR to your PATH"
        info "To add to PATH, add this line to your shell profile:"
        info "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# Main installation process
main() {
    echo "🚀 banned CLI Installer"
    echo "========================"
    
    # Check if running as root on Linux/macOS
    if [[ "$EUID" -eq 0 ]] && [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
        warn "Running as root. This is usually not necessary."
    fi
    
    # Check for required tools
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        error "Either curl or wget is required"
    fi
    
    local platform=$(detect_platform)
    info "Detected platform: $platform"
    
    get_latest_version
    install_binary "$platform"
    verify_installation
    
    echo ""
    success "Installation completed! 🎉"
    echo ""
    echo "Next steps:"
    echo "  1. Run 'banned --help' to see available commands"
    echo "  2. Run 'banned fetch channels' to populate the database"
    echo "  3. Run 'banned db view' to browse channels interactively"
    echo ""
    echo "For more information, visit: https://github.com/${REPO}"
}

# Run main function
main "$@"