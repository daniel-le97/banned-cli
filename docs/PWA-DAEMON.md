# 📱 PWA Daemon Mode

This document explains how to run banned-cli as a Progressive Web App (PWA) daemon service.

## 🚀 Quick Start

### Basic PWA Server
```bash
# Start PWA web interface
banned daemon --webui --port 8080

# Access at: http://localhost:8080
# Install as app using browser's install prompt
```

### Full Daemon with Auto-Sync
```bash
# Start with both web UI and automatic syncing
banned daemon --webui --auto-sync --interval 1h

# Background daemon with logging
banned daemon --webui --auto-sync --interval 30m \
  --log-file /var/log/banned/daemon.log \
  --pid-file /var/run/banned.pid
```

## 📱 PWA Features

### Installation
- **Desktop**: Install via browser prompt or address bar icon
- **Mobile**: "Add to Home Screen" from browser menu
- **Offline**: Works offline with cached data
- **Native Feel**: Standalone app experience

### PWA Capabilities
- 🔄 **Background Sync** - Updates data when connection restored
- 📱 **App Installation** - Install like native mobile/desktop app
- 🌐 **Offline Support** - Browse cached data without internet
- 🔔 **Push Notifications** - Daemon status updates (future)
- 📊 **Native UI** - Looks and feels like native app

## ⚙️ Daemon Services

### Web UI Service (`--webui`)
- **Progressive Web App** interface
- **Database browser** with full SQLite editor
- **Real-time status** monitoring  
- **Responsive design** for all devices
- **Offline functionality** via service worker

### Auto-Sync Service (`--auto-sync`)
- **Periodic syncing** from banned.video API
- **Configurable intervals** (30m, 1h, 2h, etc.)
- **Error handling** with retry logic
- **Background processing** without blocking

## 🔧 Installation as System Service

### Linux (systemd)
```bash
# Install daemon service
sudo ./scripts/install-daemon.sh install

# Start service
sudo systemctl start banned-daemon

# Enable auto-start
sudo systemctl enable banned-daemon

# Check status
sudo systemctl status banned-daemon
```

### Manual Installation
```bash
# Create service user
sudo useradd --system --home-dir /home/banned --create-home banned

# Copy binary
sudo cp banned /usr/local/bin/

# Create directories
sudo mkdir -p /var/log/banned /var/run/banned
sudo chown banned:banned /var/log/banned /var/run/banned

# Run daemon
sudo -u banned banned daemon --webui --auto-sync \
  --log-file /var/log/banned/daemon.log \
  --pid-file /var/run/banned/banned.pid
```

## 📊 API Endpoints

The PWA daemon exposes REST API endpoints:

### Sync Endpoints
```bash
GET /api/sync/channels    # Trigger channel sync
GET /api/sync/videos      # Trigger video sync
GET /api/daemon/status    # Get daemon status
```

### Database API
```bash
GET /api/tables           # List database tables
POST /api/query          # Execute SQL query
GET /api/stats           # Database statistics
```

## 🌐 PWA Manifest

The app includes a full PWA manifest with:
- **App name**: "Banned CLI Dashboard"
- **Theme colors**: Banned.video branding
- **Icons**: Multiple sizes for all platforms
- **Display mode**: Standalone app experience
- **Shortcuts**: Quick access to features

## 🔒 Security Considerations

### Network Security
- **Local access only** by default (localhost)
- **No authentication** - intended for personal use
- **Firewall rules** if exposing externally
- **HTTPS recommended** for production deployment

### System Security
- **Dedicated user** for daemon process
- **Limited permissions** via systemd
- **Resource limits** to prevent abuse
- **Log rotation** to prevent disk fill

## 📱 Mobile Experience

### Installation on Mobile
1. **Open browser** to daemon URL
2. **Tap share button** or browser menu
3. **Select "Add to Home Screen"**
4. **App appears** on home screen like native app

### Mobile Features
- **Touch-optimized** interface
- **Swipe gestures** for navigation
- **Responsive layout** for all screen sizes
- **Offline browsing** of cached data

## 🛠️ Troubleshooting

### Common Issues

**PWA won't install**
- Ensure HTTPS (or localhost)
- Check manifest.json is accessible
- Verify service worker registration

**Service worker errors**
- Check browser developer console
- Clear browser cache and reload
- Verify sw.js file is served correctly

**Daemon won't start**
- Check port availability (`lsof -i :8080`)
- Verify file permissions
- Check daemon logs for errors

### Debug Commands
```bash
# Check daemon status
banned daemon --webui &
curl http://localhost:8080/api/daemon/status

# Test API endpoints
curl http://localhost:8080/api/sync/channels
curl http://localhost:8080/api/sync/videos

# View logs
tail -f /var/log/banned/daemon.log
journalctl -u banned-daemon -f
```

## 🚧 Future Enhancements

### Planned Features
- 🔔 **Push notifications** for sync completion
- 🔐 **Authentication system** for multi-user access
- 📊 **Advanced analytics** dashboard
- 🌍 **Remote access** with secure tunneling
- 🎨 **Theme customization** options
- 🔄 **Real-time updates** via WebSocket

### API Expansion
- WebSocket connections for real-time updates
- Authentication and user management
- Bulk operations and batch processing
- Advanced query builder interface
- Export/import functionality

---

For more information, see:
- [PWA Documentation](https://web.dev/progressive-web-apps/)
- [Service Workers](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API)
- [systemd Services](https://www.freedesktop.org/software/systemd/man/systemd.service.html)