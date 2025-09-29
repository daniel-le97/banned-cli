# Banned - banned.video Content Manager

A comprehensive command-line tool for downloading and managing content from https://banned.video. This tool provides an interactive Terminal User Interface (TUI) for browsing channels, fetching video metadata, and managing downloads with advanced features like recursive fetching, intelligent conditional fetching, and torrent creation.

**Latest Updates:**

- 🔄 **Database Migration**: Migrated from SQLite to bbolt for improved performance
- 🧠 **Smart Fetching**: Conditional video fetching only when database < API video counts
- ⚡ **Performance**: Worker pool optimization for file-sizes (~21.9 videos/second)
- 🖱️ **Enhanced UI**: Interactive database viewer with mouse support and clipboard functionality

## 🚀 Features

- **Interactive TUI**: Browse channels and videos with a beautiful terminal interface
- **Smart Conditional Fetching**: Only fetches videos when database count < API total (avoids unnecessary API calls)
- **High Performance**: Optimized for bulk operations with worker pools (~21.9 videos/second for file sizes)
- **Database Management**: bbolt key-value database for fast local storage and caching
- **Interactive Database Viewer**: Mouse-enabled browsing with clipboard support for data inspection
- **File Size Preservation**: Maintains existing file sizes during video updates to prevent re-fetching
- **Worker Pool Architecture**: 100 concurrent workers for high-throughput operations
- **Torrent Support**: Create .torrent files from downloaded videos
- **Smart Installation**: Multiple installation methods with PATH management
- **Self-Updating**: Built-in update mechanism from source repositories

## 📦 Installation

### One-Line Install (Recommended)

**Linux/macOS:**

```bash
curl -sSL https://github.com/daniel-le97/banned-cli/releases/latest/download/install.sh | bash
```

**Windows (PowerShell):**

```powershell
iwr -useb https://github.com/daniel-le97/banned-cli/releases/latest/download/install.ps1 | iex
```

### Manual Download

Download the latest release for your platform from:
[GitHub Releases](https://github.com/daniel-le97/banned-cli/releases/latest)

### From Go

```bash
go install github.com/daniel-le97/banned-cli@latest
```

### From Source

```bash
git clone https://github.com/daniel-le97/banned-cli
cd banned-cli
go build -o banned .
```

## 🎯 Quick Start

### Interactive Mode

```bash
# Launch interactive TUI (default when no arguments)
banned

# Or explicitly
banned --interactive
```

### Initialize Database

```bash
banned db init
```

### Fetch All Channels

```bash
banned fetch channels
```

### Browse Channels Interactively

```bash
banned list-channels
```

## 📚 Commands Reference

### Root Command

```bash
banned [--interactive]
```

- Launches interactive TUI when no subcommands provided
- Use `--interactive` flag to force interactive mode

### Fetch Commands

#### `banned fetch channels`

Fetch all available channels from banned.video API and store in database.

```bash
banned fetch channels
```

#### `banned fetch videos <channel-id>`

Fetch specific channel details and videos with intelligent conditional fetching.

```bash
# Fetch first 50 videos (default)
banned fetch videos 12345

# Fetch ALL videos with smart conditional logic (recommended)
banned fetch videos 12345 --all

# Examples
banned fetch videos 5b885d33e6646a0015a6fa2d --all    # Fetch Alex Jones channel
banned fetch videos 5cf7df690a17850012626701 --all    # Fetch Mike Adams channel
```

**Smart Conditional Fetching:**

- **Intelligence**: Only fetches when database video count < API totalVideos count
- **Efficiency**: Skips unnecessary API calls when database is up-to-date
- **Performance**: Shows "missing X videos" and only fetches the difference
- **Status Display**: Clear comparison of database vs API video counts

**Performance Notes:**

- Default: Fetches 50 videos per request
- `--all`: Conditionally fetches until database matches API count
- Optimized batching: 200-video API calls, efficient database storage
- Displays elapsed time and performance metrics
- Automatic file size fetching in background with worker pools

### Sync Commands (Incremental Updates)

#### `banned sync all`

Perform a complete incremental sync of all channels and videos.

```bash
banned sync all
```

This will:

1. Sync any new/updated channels
2. For each channel, sync only new videos published since the last sync
3. Show detailed progress and statistics

**Benefits:**

- Much faster than full fetches (only gets new data)
- Reduces API load and bandwidth usage
- Perfect for regular updates with pre-populated database

#### `banned sync channels`

Sync only new/updated channels.

```bash
banned sync channels
```

#### `banned sync channel <channel-id>`

Sync new videos for a specific channel based on timestamps.

```bash
# Sync only new videos for a channel
banned sync channel alex-jones
banned sync channel 5b885d33e6646a0015a6fa2d
```

**How it works:**

- Checks your database for the most recent video timestamp
- Fetches only videos published after that timestamp
- Automatically fetches file sizes for new videos in background
- Efficiently updates your local cache without re-downloading

**Options:**

```bash
# Skip file size fetching for faster sync (useful for large batches)
banned sync channel alex-jones --skip-file-sizes
banned sync all --skip-file-sizes
```

### Database Commands

#### `banned db init`

Initialize the database and create all necessary tables.

```bash
banned db init
```

#### `banned db status`

Show database connection status and statistics.

```bash
banned db status
```

Displays:

- Database location and connection status
- Bucket counts (channels, videos, downloads, settings)
- Storage usage and performance metrics

#### `banned db view`

Interactive database viewer with mouse support and clipboard functionality.

```bash
banned db view
```

Features:

- **Mouse Support**: Click to select items, drag to scroll
- **Clipboard Integration**: Copy data to clipboard with Enter/Ctrl+C
- **Bucket Navigation**: Use ←/→ to switch between data buckets
- **Row Navigation**: Use ↑/↓ to browse items
- **Details Panel**: Press Enter/Space to view full item details
- **Search & Filter**: Type 'd' to toggle details view

#### `banned db settings`

Display all current settings stored in the database.

```bash
banned db settings
```

### Interactive Browsing

#### `banned list-channels`

Launch interactive TUI for browsing channels with search functionality.

```bash
banned list-channels
```

Features:

- Search channels by name (type to filter)
- Navigate with arrow keys or vim-style (j/k)
- Press Enter to fetch videos for selected channel
- Press 'q' to quit, '/' to search

### Torrent Commands

#### `banned torrent create [files...]`

Create torrent files from video files.

```bash
# Single file
banned torrent create video.mp4

# Multiple files
banned torrent create video1.mp4 video2.mp4

# With custom tracker
banned torrent create video.mp4 --tracker "udp://custom.tracker.com:8080/announce"

# With custom piece size (default: 256KB)
banned torrent create video.mp4 --piece-length 512
```

#### `banned torrent batch [directory]`

Create torrents for all videos in a directory recursively.

```bash
# Current directory
banned torrent batch

# Specific directory
banned torrent batch /path/to/videos

# With custom tracker for all torrents
banned torrent batch /path/to/videos --tracker "udp://my.tracker.com:80/announce"
```

### Application Management

#### `banned app install`

Install the application to PATH with multiple methods.

```bash
# Interactive installation (choose method)
banned app install

# Force symlink method
banned app install --symlink

# Force shell profile method
banned app install --profile
```

#### `banned app update`

Update the application to the latest version.

```bash
# Update from default source
banned app update

# Update from specific repository
banned app update --source github.com/yourusername/cli-aj@latest

# Update from custom Git URL
banned app update --source "https://github.com/yourusername/cli-aj.git"
```

### Download Commands

#### `banned download`

Download video content (implementation varies based on requirements).

```bash
banned download [options]
```

## 🗂️ Database Schema

The application uses bbolt (key-value store) with the following buckets:

- **channels**: Channel metadata (id, name, slug, description, totalVideos, etc.)
- **videos**: Video metadata (id, title, channel_id, duration, file_size, etc.) with preserved file sizes
- **downloads**: Download tracking and status
- **settings**: Application configuration and preferences

**Performance Benefits of bbolt Migration:**

- Faster read/write operations for large datasets
- Better concurrent access handling
- Reduced memory footprint
- Eliminates database lock issues under high load

## ⚙️ Configuration

### Managing Settings

Use the `config` command to manage application settings:

```bash
# List all current settings
banned config list

# Set download directory
banned config set download_dir ~/Videos/banned

# Set torrent trackers
banned config set torrent_trackers "udp://tracker1.com:80/announce,udp://tracker2.com:80/announce"

# Set max concurrent downloads
banned config set max_concurrent_downloads 5

# Get current download directory
banned config get download_dir

# Reset a setting to default
banned config reset download_dir
```

### Available Settings

- **download_dir**: Directory where downloaded files are saved
- **max_concurrent_downloads**: Maximum concurrent downloads (1-10)
- **retry_attempts**: Number of retry attempts for failed downloads (0-10)
- **user_agent**: User agent string for HTTP requests
- **torrent_trackers**: Comma-separated list of torrent tracker URLs
- **torrent_piece_length**: Torrent piece length in KB (64-1024)

### Database Location

- **Linux/macOS**: `~/.config/banned/banned.db`
- **Windows**: `%APPDATA%\banned\banned.db`
- **Current Directory**: `./banned.db` (fallback)

**Note**: Database migrated from SQLite to bbolt for improved performance and reliability.

### Default Settings

- Download Directory: `~/Downloads/banned/`
- API Endpoint: `https://api.banned.video/graphql`
- User Agent: `banned-cli/1.0`
- Torrent Piece Length: 256KB
- Default Trackers: Multiple public BitTorrent trackers

## 🔧 Advanced Usage

### Batch Operations

```bash
# Fetch all channels, then recursively fetch all videos
banned fetch channels
banned fetch channel --all $(banned db status | grep -o '[0-9]* channels' | cut -d' ' -f1)

# Create torrents for all downloaded videos
banned torrent batch ~/Downloads/banned/
```

### Performance Optimization

```bash
# For large channels, use --all flag for optimal API usage with conditional fetching
banned fetch videos large-channel-id --all

# Monitor performance with timing (includes intelligent skipping)
time banned fetch videos 5b885d33e6646a0015a6fa2d --all

# Background file size fetching with 100 concurrent workers
banned fetch file-sizes 5cf7df690a17850012626701  # ~21.9 videos/second
```

**Worker Pool Performance:**

- File size fetching: 100 concurrent workers
- Throughput: ~21.9 videos/second (6,802 videos in 5m11s)
- Error categorization: timeouts, network failures, HTTP errors
- Real-time progress reporting with error statistics

### Scripting Integration

```bash
#!/bin/bash
# Automated content sync script
banned db init
banned fetch channels
for channel in $(banned list-channels --json | jq -r '.[] | .id'); do
    banned fetch channel $channel --all
done
```

## 🐛 Troubleshooting

### Common Issues

#### Database Locked Warnings

```
SQLITE_BUSY warnings during high-throughput operations
```

**Solution**: These warnings are normal under heavy load and are handled gracefully with retry logic.

#### Go Environment Issues

```bash
# Check Go installation
go version

# Verify GOPATH/GOBIN
echo $GOPATH
echo $GOBIN
```

#### Performance Concerns

- **15K videos in ~2.4 minutes** is normal performance
- Use `--all` flag for optimal API rate limit usage
- Database operations are optimized with batched transactions

### Debug Mode

```bash
# Enable verbose logging (if implemented)
banned --verbose fetch channel 12345 --all

# Check database status for issues
banned db status
```

## 🤝 Development

### Building from Source

```bash
git clone https://github.com/yourusername/cli-aj
cd cli-aj
go mod download
go build -o banned .
```

### Dependencies

- **Cobra CLI**: Command-line interface framework
- **Bubble Tea**: Terminal UI framework
- **SQLite**: Database storage
- **GraphQL**: API client for banned.video

### Project Structure

```
├── cmd/          # Command implementations
│   ├── root.go   # Root command and CLI setup
│   ├── fetch.go  # Data fetching commands
│   ├── db.go     # Database operations
│   ├── tui.go    # Interactive terminal UI
│   └── ...
├── main.go       # Application entry point
├── go.mod        # Go module definition
└── README.md     # This file
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **banned.video**: https://banned.video
- **API Documentation**: https://api.banned.video/graphql
- **Issues**: Report bugs and feature requests
- **Contributing**: Pull requests welcome

---

**Note**: This tool is for educational and archival purposes. Please respect the terms of service of banned.video and applicable copyright laws.
