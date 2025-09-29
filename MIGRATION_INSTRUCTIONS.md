# Migration Instructions: bbolt → SQLite

This document outlines the key improvements and features implemented during the bbolt migration that should be preserved when reverting to SQLite. All functionality and performance optimizations should be maintained with SQLite as the backend.

## 🎯 Overview of Improvements

During the bbolt migration, we implemented several critical performance and functionality improvements that significantly enhance the CLI's capabilities. These improvements are **database-agnostic** and should be preserved when reverting to SQLite.

## 🚀 Key Features to Preserve

### 1. **Smart Conditional Fetching Logic**

**Location**: `cmd/fetch.go` (lines 115-160)

**Description**: Intelligent video fetching that only downloads videos when the database count is less than the API's reported total videos.

**Key Implementation**:

```go
// Compare database count vs API total
currentVideoCount, err := CountVideosForChannel(channelID)
totalVideos := int(channel.TotalVideos)

if currentVideoCount < totalVideos {
    shouldFetch = true
    fetchReason = fmt.Sprintf("missing %d videos", totalVideos-currentVideoCount)
} else {
    fetchReason = "all videos already in database"
}
```

**Benefits**:

- Avoids unnecessary API calls when database is up-to-date
- Shows clear status: "Channel has X total videos (API) vs Y in database"
- Skips fetching with message: "Skipping video fetch: all videos already in database"

**SQLite Adaptation**: Ensure `CountVideosForChannel()` works efficiently with SQLite queries.

### 2. **High-Performance Worker Pool for File Sizes**

**Location**: `cmd/fetch.go` (`processFileSizesWithWorkerPool` function)

**Description**: 100 concurrent workers for fetching video file sizes, achieving ~21.9 videos/second performance.

**Key Metrics**:

- **Throughput**: 6,802 videos processed in 5m11s
- **Concurrency**: 100 workers
- **Error Handling**: Categorized errors (timeout, network, HTTP, database)
- **Real-time Progress**: Live updates with success/error counts

**Implementation Features**:

```go
const numWorkers = 100
// Worker pool with channels for job distribution
// Real-time progress reporting
// Error categorization and reporting
// Atomic counters for thread-safe statistics
```

**SQLite Adaptation**: Ensure database writes are properly batched and handle potential SQLite lock contention under high concurrency.

### 3. **Data Preservation During Updates**

**Location**: `db/videos.go` (`StoreVideos` function)

**Description**: Preserves existing file sizes when updating video records to prevent re-fetching overhead.

**Logic**:

```go
// Before storing new video data, check if video already exists
// If exists and has file_size, preserve the file_size value
// Only update other metadata fields
```

**Benefits**:

- Prevents file size data loss during video updates
- Avoids expensive re-fetching of file sizes
- Maintains data integrity across updates

**SQLite Adaptation**: Use `INSERT OR REPLACE` with careful field selection or `UPDATE` with conditional logic.

### 4. **Enhanced Interactive Database Viewer**

**Location**: `cmd/db_viewer.go`

**Description**: Full-screen database viewer with mouse support and clipboard functionality.

**Features**:

- **Mouse Support**: Click selection, drag scrolling
- **Clipboard Integration**: Copy data with Enter/Ctrl+C
- **Full Terminal Usage**: Utilizes entire terminal space
- **Table Navigation**: Switch between different data tables
- **Details Panel**: View complete record details
- **Search/Filter**: Real-time filtering capabilities

**Key Components**:

```go
// Mouse event handling
// Clipboard operations
// Full-screen terminal UI with Bubble Tea
// Data serialization for clean console output
```

**SQLite Adaptation**: Ensure queries work with SQLite schema (tables instead of buckets).

### 5. **Improved Error Reporting and Progress Tracking**

**Location**: Multiple files (`cmd/fetch.go`, `cmd/utils.go`)

**Description**: Comprehensive error categorization and real-time progress reporting.

**Features**:

- **Error Categories**: Timeout, Network, HTTP, Database errors
- **Real-time Updates**: Live progress counters
- **Performance Metrics**: Processing speed, success/failure rates
- **User Feedback**: Clear status messages and completion summaries

**Implementation**:

```go
// Atomic counters for thread-safe statistics
// Error type detection and categorization
// Progress reporting with time estimates
// Comprehensive logging and user feedback
```

**SQLite Adaptation**: Ensure error handling works with SQLite-specific errors.

## 🔄 Migration Strategy

### Phase 1: Database Layer Abstraction

1. **Preserve Interface**: Keep the same function signatures in `cmd/database.go`
2. **Update Implementation**: Change underlying calls from bbolt to SQLite
3. **Maintain Buckets→Tables Mapping**:
   - `channels` bucket → `channels` table
   - `videos` bucket → `videos` table
   - `downloads` bucket → `downloads` table
   - `settings` bucket → `settings` table

### Phase 2: Schema Updates

```sql
-- Ensure SQLite schema supports all features
CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    title TEXT,
    summary TEXT,
    total_videos INTEGER,
    -- ... other fields
);

CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY,
    channel_id TEXT,
    title TEXT,
    file_size INTEGER,  -- Critical for preservation logic
    -- ... other fields
    FOREIGN KEY (channel_id) REFERENCES channels(id)
);
```

### Phase 3: Performance Optimizations

1. **Indexing**: Ensure proper indexes for fast lookups
2. **Batch Operations**: Use transactions for bulk inserts
3. **Connection Pooling**: Handle concurrent access properly
4. **WAL Mode**: Enable Write-Ahead Logging for better concurrency

### Phase 4: Function Mapping

| bbolt Function               | SQLite Equivalent                                                 | Notes                          |
| ---------------------------- | ----------------------------------------------------------------- | ------------------------------ |
| `CountVideosForChannel()`    | `SELECT COUNT(*) FROM videos WHERE channel_id = ?`                | Ensure index on channel_id     |
| `GetVideosWithoutFileSize()` | `SELECT * FROM videos WHERE channel_id = ? AND file_size IS NULL` | For worker pool                |
| `StoreVideos()`              | `INSERT OR REPLACE` with file_size preservation                   | Critical for data preservation |
| `GetChannelByID()`           | `SELECT * FROM channels WHERE id = ?`                             | Standard query                 |

## 🎯 Critical Success Criteria

When reverting to SQLite, ensure these behaviors are preserved:

### ✅ Conditional Fetching

- Shows "Channel has X total videos (API) vs Y in database"
- Only fetches when database count < API count
- Displays "missing N videos" or "all videos already in database"

### ✅ Worker Pool Performance

- 100 concurrent workers for file size fetching
- Real-time progress with error categorization
- Maintains ~20+ videos/second processing speed
- Handles database concurrency without corruption

### ✅ Data Preservation

- File sizes preserved during video updates
- No data loss when running fetch commands multiple times
- Efficient updates that don't trigger unnecessary re-fetching

### ✅ Interactive Viewer

- Mouse support for selection and scrolling
- Clipboard functionality for data copying
- Full-screen utilization with proper table switching
- Clean, formatted data display

### ✅ Error Handling

- Categorized error reporting (timeout, network, HTTP, database)
- Graceful handling of SQLite lock contention
- Retry logic for failed operations
- Clear user feedback for all operations

## 🛠️ Implementation Notes

### Database Configuration

```go
// Recommended SQLite settings for high concurrency
db, err := sql.Open("sqlite3", "file:banned.db?cache=shared&mode=rwc&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL")
```

### Concurrency Handling

```go
// Use connection pooling for worker pools
db.SetMaxOpenConns(25)  // Limit concurrent connections
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Transaction Management

```go
// Batch operations for performance
tx, err := db.Begin()
defer tx.Rollback()
// ... batch operations
tx.Commit()
```

## 📋 Testing Checklist

Before completing the SQLite migration, verify:

- [ ] Conditional fetching works correctly
- [ ] Worker pool processes file sizes without errors
- [ ] File sizes are preserved during updates
- [ ] Database viewer displays all data correctly
- [ ] Mouse interactions work in the viewer
- [ ] Clipboard functionality works
- [ ] Error reporting shows proper categories
- [ ] Performance meets or exceeds previous benchmarks
- [ ] Concurrent operations don't cause database locks
- [ ] All existing commands work without regression

## 🎉 Success Metrics

The migration is successful when:

1. **Performance**: File size processing maintains 15+ videos/second
2. **Reliability**: Zero database corruption under high concurrency
3. **Functionality**: All interactive features work identically
4. **Efficiency**: Conditional fetching reduces unnecessary API calls by 80%+
5. **User Experience**: No breaking changes to existing workflows

---

**Note**: This document serves as a comprehensive guide for maintaining all improvements while reverting the database backend. The goal is to keep all the intelligent behaviors and performance optimizations we've achieved during the bbolt implementation.
