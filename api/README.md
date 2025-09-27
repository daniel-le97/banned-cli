# banned.video API Client

A comprehensive Go client library for the banned.video GraphQL API. This package provides a clean, type-safe interface for interacting with banned.video's content management system.

## Features

- **Full GraphQL API Coverage**: Complete client for banned.video's GraphQL API
- **Type-Safe Operations**: Strongly typed Go structs for all API responses
- **Automatic Pagination**: Built-in support for recursive data fetching
- **Concurrent Safe**: Thread-safe client suitable for concurrent operations
- **Comprehensive Error Handling**: Detailed error reporting with GraphQL error support
- **Flexible Configuration**: Customizable timeouts, user agents, and base URLs
- **Helper Methods**: Convenience methods for common operations
- **Extensive Examples**: Complete usage examples and documentation

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "libertyarchive.com/banned/api"
)

func main() {
    // Create a new API client
    client := api.New()

    ctx := context.Background()

    // Fetch all channels
    channels, err := client.Channels.GetAllChannels(ctx, 0, 10)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d channels\n", len(channels))

    // Fetch videos from the first channel
    if len(channels) > 0 {
        videos, err := client.Channels.GetChannelVideos(ctx, channels[0].ID, 0, 5)
        if err != nil {
            log.Fatal(err)
        }

        fmt.Printf("Channel '%s' has %d videos\n", channels[0].Title, len(videos))
    }
}
```

## API Structure

The client is organized into services for different resource types:

### Core Client

- **`Client`**: Low-level GraphQL client with connection management
- **`API`**: High-level interface combining all services

### Services

- **`ChannelService`**: Channel management and video fetching
- **`VideoService`**: Video operations and metadata retrieval

## Channel Operations

### Basic Channel Fetching

```go
// Get all channels with pagination
channels, err := client.Channels.GetAllChannels(ctx, 0, 50)

// Get specific channel by ID
channel, err := client.Channels.GetChannel(ctx, "channel-id")

// Get channel with videos
params := api.ChannelVideosParams{
    ChannelID: "channel-id",
    IncludeUnlisted: false,
    IncludeUnpublished: false,
    IncludeLive: true,
    Offset: 0,
    Limit: 20,
}
channel, err := client.Channels.GetChannelWithVideos(ctx, params)
```

### Recursive Video Fetching

```go
// Fetch ALL videos from a channel (automatic pagination)
allVideos, err := client.Channels.FetchAllChannelVideos(ctx, "channel-id", 50)

// Fetch up to a specific limit
limitedVideos, err := client.Channels.FetchAllChannelVideosWithLimit(ctx, "channel-id", 1000, 50)

// Get hot videos from a channel
hotVideos, err := client.Channels.GetChannelHotVideos(ctx, "channel-id", 0, 10)
```

## Video Operations

### Basic Video Fetching

```go
// Get specific video by ID
video, err := client.Videos.GetVideo(ctx, "video-id")

// Get multiple videos by IDs
videos, err := client.Videos.GetVideos(ctx, []string{"id1", "id2", "id3"})

// Get hot/trending videos
hotVideos, err := client.Videos.GetHotVideos(ctx, 0, 20)

// Get newest videos
newVideos, err := client.Videos.GetNewVideos(ctx, 0, 20)
```

### Helper Methods

```go
// Get download URLs from video
downloadURL := client.Videos.GetVideoDownloadURL(video)
posterURL := client.Videos.GetVideoPosterURL(video)
audioURL := client.Videos.GetVideoAudioURL(video)
```

## Advanced Usage

### Custom Configuration

```go
import "time"

// Create client with custom settings
client := api.NewWithConfig(
    "https://api.banned.video/graphql",  // Base URL
    "my-app/2.0",                        // User Agent
    60 * time.Second,                    // Timeout
)

// Update settings after creation
client.SetTimeout(30 * time.Second)
client.SetUserAgent("updated-agent/1.0")
```

### Bulk Operations

```go
// Fetch multiple channels with their videos
channelIDs := []string{"id1", "id2", "id3"}
bulkData, err := client.BulkFetchChannelData(ctx, channelIDs, 10)

for _, channelData := range bulkData {
    fmt.Printf("Channel: %s, Videos: %d\n",
        channelData.Title, len(channelData.Videos))
}
```

### Error Handling

```go
channels, err := client.Channels.GetAllChannels(ctx, 0, 10)
if err != nil {
    // Handle GraphQL errors
    fmt.Printf("API Error: %v\n", err)
    return
}

// Check for empty results
if len(channels) == 0 {
    fmt.Println("No channels found")
    return
}
```

## Data Types

### Core Types

- **`Channel`**: Complete channel information with metadata
- **`Video`**: Full video details with uploads and statistics
- **`AdminUser`**: Channel creator/admin information
- **`Tag`**: Video categorization tags
- **`Upload`**: File upload details (video, audio, poster)
- **`Playlist`**: Channel playlist information

### Utility Types

- **`ChannelVideosParams`**: Parameters for channel video queries
- **`Pagination`**: Pagination helper
- **`VideosResponse`**: Video collection with metadata
- **`ChannelsResponse`**: Channel collection with metadata

## Configuration

### Environment Variables

```bash
# Optional: Override default API endpoint
BANNED_API_URL="https://api.banned.video/graphql"

# Optional: Set default user agent
BANNED_USER_AGENT="my-app/1.0"
```

### Client Configuration

```go
// Default configuration
client := api.New()

// Custom configuration
client := api.NewWithConfig(
    "https://custom.api.url/graphql",
    "custom-agent/1.0",
    30 * time.Second,
)

// Get current configuration
info := client.GetAPIInfo()
fmt.Printf("Base URL: %s\n", info["base_url"])
fmt.Printf("User Agent: %s\n", info["user_agent"])
fmt.Printf("Timeout: %v\n", info["timeout"])
```

## Performance Considerations

### Pagination Best Practices

```go
// Use appropriate batch sizes for your use case
const (
    SmallBatch  = 20   // Interactive applications
    MediumBatch = 50   // General purpose (recommended)
    LargeBatch  = 200  // Bulk operations
)

// For large datasets, use recursive fetching
videos, err := client.Channels.FetchAllChannelVideos(ctx, channelID, MediumBatch)
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

// Implement rate limiting for bulk operations
limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests per second

for _, channelID := range channelIDs {
    limiter.Wait(ctx) // Wait for rate limit
    videos, err := client.Channels.GetChannelVideos(ctx, channelID, 0, 50)
    // Process videos...
}
```

## Testing

### Unit Tests

```bash
go test ./api
```

### Integration Tests

```bash
# Set environment variable for integration tests
export BANNED_INTEGRATION_TEST=true
go test ./api -tags=integration
```

## Error Types

The client provides detailed error information:

```go
_, err := client.Videos.GetVideo(ctx, "invalid-id")
if err != nil {
    // GraphQL errors contain detailed information
    fmt.Printf("Error: %v\n", err)
}
```

### Common Error Scenarios

- **Network Errors**: Connection timeouts, DNS failures
- **HTTP Errors**: 4xx/5xx status codes
- **GraphQL Errors**: Invalid queries, missing resources
- **Validation Errors**: Invalid parameters, malformed IDs

## Examples

See [`examples.go`](examples.go) for comprehensive usage examples including:

- Basic API operations
- Recursive data fetching
- Batch processing
- Error handling patterns
- Performance optimization

## Dependencies

This package uses only Go standard library dependencies:

- `context`: Context handling
- `encoding/json`: JSON marshaling/unmarshaling
- `net/http`: HTTP client
- `time`: Timeout management

## License

This API client is part of the banned-cli project and follows the same MIT license.

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Ensure backward compatibility when possible

## Changelog

### v1.0.0

- Initial release
- Complete GraphQL API coverage
- Channel and video services
- Recursive pagination support
- Comprehensive error handling
- Helper methods for common operations
