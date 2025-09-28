# Banned - banned.video Content Manager

A comprehensive command-line tool for downloading and managing content from https://banned.video. This tool provides an interactive Terminal User Interface (TUI) for browsing channels, fetching video metadata, and managing downloads with advanced features like recursive fetching and torrent creation.

## 🚀 Features

- **Interactive TUI**: Browse channels and videos with a beautiful terminal interface
- **Recursive Video Fetching**: Automatically fetch all videos from channels with pagination support
- **High Performance**: Optimized for bulk operations (~103 videos/second)
- **Database Management**: SQLite database for local storage and caching
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
go install libertyarchive.com/banned@latest
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

#### `banned fetch channel <channel-id>`

Fetch specific channel details and videos with options for recursive fetching.

```bash
# Fetch first 50 videos (default)
banned fetch channel 12345

# Fetch ALL videos recursively (recommended for complete data)
banned fetch channel 12345 --all

# Fetch specific number of videos
banned fetch channel 12345 --limit 1000

# Examples
banned fetch channel 12345 --all              # Fetch all videos
banned fetch channel 607 --limit 500          # Fetch up to 500 videos
banned fetch channel alex-jones --all         # Works with channel slugs too
```

**Performance Notes:**

- Default: Fetches 50 videos per request
- `--all`: Recursively fetches until no new videos (recommended)
- `--limit N`: Sets maximum videos to fetch
- Optimized batching: 200-video API calls, 50-video DB batches
- Displays elapsed time and performance metrics

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
- Table counts (channels, videos, downloads)
- Storage usage and performance metrics

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

The application uses SQLite with the following tables:

- **channels**: Channel metadata (id, name, slug, description, etc.)
- **videos**: Video metadata (id, title, channel_id, duration, etc.)
- **downloads**: Download tracking and status
- **settings**: Application configuration and preferences

## ⚙️ Configuration

### Database Location

- **Linux/macOS**: `~/.local/share/banned/banned.db`
- **Windows**: `%APPDATA%\banned\banned.db`
- **Current Directory**: `./banned.db` (fallback)

### Default Settings

- Download Directory: `~/Downloads/banned/`
- API Endpoint: `https://api.banned.video/graphql`
- User Agent: `banned/1.0`
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
# For large channels, use --all flag for optimal API usage
banned fetch channel large-channel-id --all

# Monitor performance with timing
time banned fetch channel alex-jones --all
```

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
