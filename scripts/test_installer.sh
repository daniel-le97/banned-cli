#!/bin/bash

# Test installer script locally
echo "Testing installer script..."

# Test with a custom install directory (no sudo needed)
export INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

echo "Testing installer with INSTALL_DIR=$INSTALL_DIR"
./install.sh

# Check if it was installed
if [ -f "$INSTALL_DIR/banned" ]; then
    echo "✅ Installation successful!"
    echo "📍 Installed at: $INSTALL_DIR/banned"
    
    # Test the installed binary
    if "$INSTALL_DIR/banned" --version 2>/dev/null; then
        echo "✅ Binary is working!"
    else
        echo "⚠️ Binary installed but not working"
    fi
else
    echo "❌ Installation failed"
    exit 1
fi