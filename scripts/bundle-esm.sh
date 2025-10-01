#!/bin/bash
set -e

# ESM Bundler Script for Banned CLI PWA
# Downloads and bundles ESM imports for offline use

WEBUI_DIR="internal/webui/web"
VENDOR_DIR="$WEBUI_DIR/vendor"
ESM_VERSION="10.23.1"
HTM_VERSION="3.1.1"

echo "📦 Bundling ESM imports for offline PWA..."

# Create vendor directory
mkdir -p "$VENDOR_DIR"

# Download Preact
echo "⬇️  Downloading Preact v$ESM_VERSION..."
curl -s "https://esm.sh/preact@$ESM_VERSION" > "$VENDOR_DIR/preact.js"
curl -s "https://esm.sh/preact@$ESM_VERSION/hooks" > "$VENDOR_DIR/preact-hooks.js"

# Download HTM
echo "⬇️  Downloading HTM v$HTM_VERSION..."
curl -s "https://esm.sh/htm@$HTM_VERSION/preact?external=preact" > "$VENDOR_DIR/htm-preact.js"

# Create local import map for offline use
cat > "$WEBUI_DIR/importmap-offline.json" << 'EOF'
{
  "imports": {
    "preact": "./vendor/preact.js",
    "preact/hooks": "./vendor/preact-hooks.js", 
    "htm/preact": "./vendor/htm-preact.js"
  }
}
EOF

# Update service worker to include vendor files
echo "🔄 Updating service worker cache list..."

# Create updated static files list
cat > "$WEBUI_DIR/sw-config.js" << 'EOF'
// Auto-generated SW configuration
const VENDOR_FILES = [
  '/vendor/preact.js',
  '/vendor/preact-hooks.js',
  '/vendor/htm-preact.js'
];

// Export for use in service worker
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { VENDOR_FILES };
}
EOF

echo "✅ ESM imports bundled successfully!"
echo ""
echo "📋 What was created:"
echo "  • $VENDOR_DIR/preact.js - Preact library"
echo "  • $VENDOR_DIR/preact-hooks.js - Preact hooks"  
echo "  • $VENDOR_DIR/htm-preact.js - HTM template library"
echo "  • $WEBUI_DIR/importmap-offline.json - Offline import map"
echo ""
echo "🔧 To use bundled imports:"
echo "  1. Update index.html to use importmap-offline.json"
echo "  2. Update service worker to cache vendor files"
echo "  3. Add offline/online import map switching"