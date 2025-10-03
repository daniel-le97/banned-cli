# Copilot Instructions for banned-cli

This guide enables AI coding agents to be immediately productive in the `banned-cli` codebase. It covers architecture, workflows, conventions, and integration points unique to this project.

## 🏗️ Big Picture Architecture

- **Monorepo CLI**: The project is a Go monorepo for a CLI tool that manages and downloads content from banned.video.
- **Major Components**:
  - `cmd/`: All CLI commands (root, fetch, sync, db, torrent, install, update, etc.)
  - `api/`: GraphQL client library (currently unused, may be integrated later)
  - `tui/`: Terminal UI logic for interactive browsing
  - `main.go`: Entry point, wires up Cobra CLI
- **Data Flow**:
  - CLI commands interact with SQLite DB and banned.video API
  - TUI and batch operations fetch, sync, and manage video/channel metadata
  - Database schema: channels, videos, downloads, settings
- **Why**: Designed for high-throughput archival, bulk operations, and interactive exploration of banned.video content.

## 🛠️ Developer Workflows

- **Build**: `go build -o banned .` (binary: `banned`)
- **Install**: `go install github.com/daniel-le97/banned-cli@latest` (module path matches repo)
- **Test**: `go test ./...` (unit/integration tests)
- **Release**: Use `./scripts/publish.sh` for GoReleaser-based multi-platform releases (includes installer scripts)
- **Installers**: All scripts stored in `/scripts/` - `install.sh` (bash), `install.ps1` (PowerShell), `install.bat` (batch)
- **Database**: SQLite file at `~/.local/share/banned/banned.db` (Linux/macOS) or `%APPDATA%\banned\banned.db` (Windows)

## 📦 Project-Specific Conventions

- **Command Structure**: All commands are in `cmd/`, registered in `root.go`. No `app` subcommand; `install` and `update` are root-level.
- **API Integration**: Currently uses direct HTTP/GraphQL calls; `api/` module exists but is not integrated yet.
- **Batching**: Fetches use 50-200 item batches for performance; sync commands only fetch new data since last update.
- **Error Handling**: Database locked errors are normal under load; handled with retry logic.
- **Performance**: Bulk fetches can process ~15K videos in ~2.4 minutes; use `--all` for recursive API calls.
- **Environment Variables**: `BANNED_API_URL`, `BANNED_USER_AGENT` for API customization.
- **Module Path**: All internal imports use `github.com/daniel-le97/banned-cli/...`.

## 🔗 Integration Points

- **GraphQL API**: All external data comes from banned.video's GraphQL endpoint.
- **SQLite**: Local database for caching and state.
- **GoReleaser**: `.goreleaser.yaml` for cross-platform builds, version injection, and release assets.
- **Installer Scripts**: Included in releases for all platforms.

## 🧩 Patterns & Examples

- **Add a new CLI command**: Create `cmd/yourcmd.go`, register in `init()` of that file with `rootCmd.AddCommand(...)`.
- **API Usage**: Currently uses direct HTTP calls to GraphQL endpoint; `api/` module available for future integration.
- **Database Usage**:
  ```go
  db, err := sql.Open("sqlite3", dbPath)
  // ...
  ```
- **Release a new version**: Run `./scripts/publish.sh` (auto-increments version, tags, builds, uploads)

## 📝 Key Files & Directories

- `cmd/root.go`: CLI entry and command registration
- `cmd/install.go`, `cmd/update.go`: Install/update logic
- `api/`: GraphQL client and types
- `tui/`: Terminal UI logic
- `main.go`: Entrypoint
- `.goreleaser.yaml`: Release config
- `scripts/`: All automation scripts (publish.sh, install.sh, install.ps1, install.bat, etc.)
- `README.md`: Full command and workflow documentation

---

If any conventions or workflows are unclear, please ask for clarification or examples from the codebase before proceeding with major changes.
