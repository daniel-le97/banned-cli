#!/bin/bash
set -e

# banned CLI publish script
# Usage: ./publish.sh [patch|minor|major|<version>]

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

step() {
    echo -e "${CYAN}🚀${NC} $1"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
    fi
}

# Check if working directory is clean
check_clean_working_dir() {
    local force_flag="$1"
    
    if ! git diff-index --quiet HEAD --; then
        if [ "$force_flag" = "true" ]; then
            warn "Working directory is not clean (--force specified, continuing anyway)"
            git status --porcelain
            return
        fi
        
        warn "Working directory is not clean. Uncommitted changes:"
        git status --porcelain
        echo ""
        read -p "Do you want to continue anyway? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            error "Aborted by user"
        fi
    fi
}

# Check if required tools are installed
check_dependencies() {
    local missing_deps=()
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if ! command -v goreleaser &> /dev/null; then
        missing_deps+=("goreleaser")
    fi
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        error "Missing dependencies: ${missing_deps[*]}"
    fi
}

# Get current version from git tags
get_current_version() {
    local current_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "$current_tag"
}

# Parse semantic version
parse_version() {
    local version="$1"
    # Remove 'v' prefix if present
    version=${version#v}
    
    # Split version into parts
    IFS='.' read -ra PARTS <<< "$version"
    
    MAJOR=${PARTS[0]:-0}
    MINOR=${PARTS[1]:-0}
    PATCH=${PARTS[2]:-0}
    
    # Remove any pre-release or build metadata
    PATCH=${PATCH%%-*}
    PATCH=${PATCH%%+*}
}

# Increment version based on type
increment_version() {
    local increment_type="$1"
    local current_version="$2"
    
    parse_version "$current_version"
    
    case "$increment_type" in
        "major")
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        "minor")
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        "patch")
            PATCH=$((PATCH + 1))
            ;;
        *)
            error "Invalid increment type: $increment_type"
            ;;
    esac
    
    echo "v$MAJOR.$MINOR.$PATCH"
}

# Validate version format
validate_version() {
    local version="$1"
    
    # Remove 'v' prefix for validation
    local version_no_v=${version#v}
    
    if [[ ! $version_no_v =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: $version (expected: vX.Y.Z)"
    fi
}

# Update version in files (if you have version files)
update_version_files() {
    local new_version="$1"
    
    # Update version in main.go or cmd/version.go if they exist
    if [ -f "cmd/version.go" ]; then
        info "Updating version in cmd/version.go"
        sed -i "s/Version = \".*\"/Version = \"$new_version\"/" cmd/version.go
    fi
    
    # You can add more files here if needed
}

# Create git tag and push
create_and_push_tag() {
    local new_version="$1"
    
    step "Creating git tag: $new_version"
    
    # Create annotated tag
    git tag -a "$new_version" -m "Release $new_version"
    
    # Push tag to remote
    git push origin "$new_version"
    
    success "Tag $new_version created and pushed"
}

# Run tests before release
run_tests() {
    step "Running tests..."
    
    # Run go tests if test files exist
    if ls *_test.go 1> /dev/null 2>&1 || find . -name "*_test.go" | grep -q .; then
        go test ./... || error "Tests failed"
        success "All tests passed"
    else
        info "No tests found, skipping test phase"
    fi
}

# Build and validate
build_and_validate() {
    step "Building and validating with GoReleaser..."
    
    # Check GoReleaser configuration
    goreleaser check || error "GoReleaser configuration is invalid"
    
    # Build without releasing to validate
    goreleaser build --snapshot --clean || error "Build validation failed"
    
    success "Build validation passed"
}

# Create release
create_release() {
    local dry_run="$1"
    
    if [ "$dry_run" = "true" ]; then
        step "Creating dry-run release (no actual publish)..."
        goreleaser release --snapshot --clean --skip=publish
    else
        step "Creating and publishing release..."
        goreleaser release --clean
    fi
    
    success "Release process completed"
}

# Show release summary
show_release_summary() {
    local new_version="$1"
    local repo_url="https://github.com/daniel-le97/banned-cli"
    
    echo ""
    echo "🎉 Release Summary"
    echo "=================="
    echo "Version: $new_version"
    echo "Repository: $repo_url"
    echo "Release URL: $repo_url/releases/tag/$new_version"
    echo ""
    echo "Install commands:"
    echo "  curl -sSL $repo_url/releases/latest/download/install.sh | bash"
    echo ""
    echo "Direct downloads:"
    echo "  $repo_url/releases/tag/$new_version"
    echo ""
}

# Show usage
show_usage() {
    echo "Usage: $0 [patch|minor|major|<version>] [options]"
    echo ""
    echo "Version increment (defaults to 'patch' if not specified):"
    echo "  (no args)     Auto-increment patch version (X.Y.Z -> X.Y.Z+1)"
    echo "  patch         Increment patch version (X.Y.Z -> X.Y.Z+1)"
    echo "  minor         Increment minor version (X.Y.Z -> X.Y+1.0)"
    echo "  major         Increment major version (X.Y.Z -> X+1.0.0)"
    echo "  vX.Y.Z        Use specific version"
    echo ""
    echo "Options:"
    echo "  --dry-run     Build and validate but don't publish"
    echo "  --skip-tests  Skip running tests"
    echo "  --force       Skip working directory clean check"
    echo "  --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Auto patch: v1.2.3 -> v1.2.4"
    echo "  $0 patch                    # Same as above"
    echo "  $0 minor                    # v1.2.3 -> v1.3.0" 
    echo "  $0 major                    # v1.2.3 -> v2.0.0"
    echo "  $0 v1.5.0                   # Use specific version"
    echo "  $0 --dry-run                # Test patch release process"
    echo "  $0 minor --dry-run          # Test minor release process"
}

# Main function
main() {
    local increment_type=""
    local dry_run="false"
    local skip_tests="false"
    local force_clean="false"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                dry_run="true"
                shift
                ;;
            --skip-tests)
                skip_tests="true"
                shift
                ;;
            --force)
                force_clean="true"
                shift
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            -*)
                error "Unknown option: $1"
                ;;
            *)
                if [ -z "$increment_type" ]; then
                    increment_type="$1"
                fi
                shift
                ;;
        esac
    done
    
    # Default to patch increment if no version type specified
    if [ -z "$increment_type" ]; then
        increment_type="patch"
        info "No version specified, defaulting to patch increment"
    fi
    
    echo "🚀 banned CLI Release Publisher"
    echo "==============================="
    
    # Pre-flight checks
    step "Running pre-flight checks..."
    check_dependencies
    check_git_repo
    check_clean_working_dir "$force_clean"
    
    # Get current version
    local current_version=$(get_current_version)
    info "Current version: $current_version"
    
    # Determine new version
    local new_version
    case "$increment_type" in
        "patch"|"minor"|"major")
            new_version=$(increment_version "$increment_type" "$current_version")
            ;;
        v*)
            new_version="$increment_type"
            validate_version "$new_version"
            ;;
        *)
            error "Invalid version specification: $increment_type"
            ;;
    esac
    
    info "New version: $new_version"
    
    # Show what will happen
    echo ""
    echo "📋 Release Plan"
    echo "==============="
    echo "Current version: $current_version"
    echo "New version:     $new_version" 
    echo "Increment type:  $increment_type"
    echo "Dry run:         $dry_run"
    echo "Skip tests:      $skip_tests"
    echo ""
    
    # Confirm with user (unless dry run)
    if [ "$dry_run" = "false" ]; then
        read -p "Proceed with release $new_version? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            error "Aborted by user"
        fi
    else
        info "Running in dry-run mode (no actual release will be created)"
    fi
    
    # Run tests
    if [ "$skip_tests" = "false" ]; then
        run_tests
    fi
    
    # Build and validate
    build_and_validate
    
    # Update version files and commit (only for real releases)
    if [ "$dry_run" = "false" ]; then
        update_version_files "$new_version"
        
        # Commit version changes if any files were modified
        if ! git diff-index --quiet HEAD --; then
            step "Committing version updates..."
            git add -A
            git commit -m "chore: bump version to $new_version"
            git push origin main
        fi
        
        # Create and push tag
        create_and_push_tag "$new_version"
    fi
    
    # Create release
    create_release "$dry_run"
    
    # Show summary
    if [ "$dry_run" = "false" ]; then
        show_release_summary "$new_version"
    else
        success "Dry run completed successfully"
        info "Use without --dry-run to actually publish"
    fi
}

# Run main function with all arguments
main "$@"