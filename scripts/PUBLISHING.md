# Publishing Scripts Documentation

This repository includes automated publishing scripts to streamline the release process.

## Scripts

### 🚀 `publish.sh` - Complete Release Script

A comprehensive script that handles the entire release process with smart defaults:

- **Smart defaults**: Auto-increments patch version when no arguments provided
- **Pre-flight checks**: Validates git repo, dependencies, and working directory
- **Version management**: Auto-increments versions or accepts custom versions
- **Testing**: Runs tests before release (optional)
- **Git operations**: Creates tags, commits, and pushes changes
- **GoReleaser integration**: Builds and publishes releases
- **Release summary**: Shows URLs and install instructions

#### Usage

```bash
# Auto-increment patch version (v0.3.0 -> v0.3.1) - DEFAULT
./publish.sh

# Explicit version increments
./publish.sh patch      # v0.3.0 -> v0.3.1
./publish.sh minor      # v0.3.0 -> v0.4.0
./publish.sh major      # v0.3.0 -> v1.0.0

# Use specific version
./publish.sh v1.2.5

# Options
./publish.sh --dry-run          # Test without publishing
./publish.sh --skip-tests       # Skip test execution
./publish.sh --force           # Skip clean working directory check
./publish.sh minor --dry-run   # Test minor release
```

## Release Process

1. **Prepare**: Ensure all changes are committed and pushed
2. **Test**: Run `./publish.sh --dry-run` to validate
3. **Release**: Run `./publish.sh` for patch, or `./publish.sh minor`/`./publish.sh major` as needed
4. **Verify**: Check the GitHub release page

### Quick Release Workflow

```bash
# Most common workflow (patch release)
./publish.sh --dry-run    # Test first
./publish.sh              # Create release (auto-increments patch)

# For feature releases
./publish.sh minor --dry-run    # Test minor release
./publish.sh minor              # Create minor release
```

## What Gets Published

- ✅ Cross-platform binaries (Linux, macOS, Windows)
- ✅ Archives with binary + database + install script
- ✅ Standalone `install.sh` for easy installation
- ✅ Standalone `banned.db` database file
- ✅ Checksums and signatures
- ✅ Homebrew formula (macOS)
- ✅ Scoop manifest (Windows)

## Requirements

- `git` - Version control
- `go` - Go compiler
- `goreleaser` - Release automation
- GitHub token in `$GITHUB_TOKEN` environment variable

## Install GoReleaser

```bash
# macOS
brew install goreleaser

# Linux
echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | sudo tee /etc/apt/sources.list.d/goreleaser.list
sudo apt update
sudo apt install goreleaser

# Or download from GitHub releases
```

## Environment Setup

```bash
# Set GitHub token (get from GitHub > Settings > Developer settings > Personal access tokens)
export GITHUB_TOKEN="your_github_token_here"

# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
echo 'export GITHUB_TOKEN="your_token"' >> ~/.bashrc
```

## Troubleshooting

### "Working directory is not clean"

Commit or stash your changes before releasing:

```bash
git add -A
git commit -m "your commit message"
```

### "goreleaser command not found"

Install GoReleaser following the instructions above.

### "GitHub token not found"

Set the `GITHUB_TOKEN` environment variable with your GitHub personal access token.

### Release fails

- Check GitHub token permissions (needs repo access)
- Ensure you have push access to the repository
- Verify GoReleaser configuration with `goreleaser check`
