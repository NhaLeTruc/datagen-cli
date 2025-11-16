#!/usr/bin/env bash

# Release script for datagen-cli
# Creates release builds, archives, checksums, and optionally GitHub releases

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Release configuration
VERSION="${VERSION:-}"
RELEASE_DIR="${PROJECT_ROOT}/release"
DIST_DIR="${PROJECT_ROOT}/dist"

# Binary name
BINARY_NAME="datagen"

# GitHub release settings
GITHUB_REPO="${GITHUB_REPO:-NhaLeTruc/datagen-cli}"
CREATE_GITHUB_RELEASE="${CREATE_GITHUB_RELEASE:-false}"

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Help message
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Create a release for datagen-cli.

OPTIONS:
    -h, --help              Show this help message
    -v, --version VERSION   Set release version (required)
    -g, --github            Create GitHub release (requires gh CLI)
    --draft                 Create draft GitHub release
    --prerelease            Mark as pre-release
    --dry-run               Show what would be done without creating release
    --clean                 Clean release artifacts before building

PROCESS:
    1. Validate version format
    2. Build for all platforms
    3. Create archives (tar.gz, zip)
    4. Generate checksums (SHA256)
    5. Optionally create GitHub release

EXAMPLES:
    # Create release v1.0.0
    ./scripts/release.sh --version 1.0.0

    # Create GitHub release
    ./scripts/release.sh --version 1.0.0 --github

    # Create draft GitHub release
    ./scripts/release.sh --version 1.0.0 --github --draft

    # Dry run (preview)
    ./scripts/release.sh --version 1.0.0 --dry-run

EOF
}

# Logger functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Validate version format (semantic versioning)
validate_version() {
    local version=$1

    if [[ -z "$version" ]]; then
        error "Version is required. Use --version flag."
    fi

    # Remove 'v' prefix if present
    version=${version#v}

    # Validate semantic versioning format (X.Y.Z or X.Y.Z-suffix)
    if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        error "Invalid version format: $version. Expected format: X.Y.Z or X.Y.Z-suffix (e.g., 1.0.0, 1.0.0-beta.1)"
    fi

    VERSION="$version"
    info "Validated version: v${VERSION}"
}

# Check prerequisites
check_prerequisites() {
    section "Checking Prerequisites"

    # Check for Go
    if ! command -v go >/dev/null 2>&1; then
        error "Go is not installed. Please install Go 1.21 or later."
    fi
    info "✓ Go found: $(go version | awk '{print $3}')"

    # Check for git
    if ! command -v git >/dev/null 2>&1; then
        error "Git is not installed."
    fi
    info "✓ Git found: $(git --version | awk '{print $3}')"

    # Check for tar
    if ! command -v tar >/dev/null 2>&1; then
        error "tar is not installed."
    fi
    info "✓ tar found"

    # Check for zip
    if ! command -v zip >/dev/null 2>&1; then
        warn "zip is not installed. Windows archives will be skipped."
    else
        info "✓ zip found"
    fi

    # Check for sha256sum or shasum
    if command -v sha256sum >/dev/null 2>&1; then
        info "✓ sha256sum found"
    elif command -v shasum >/dev/null 2>&1; then
        info "✓ shasum found"
    else
        error "Neither sha256sum nor shasum found. Please install coreutils."
    fi

    # Check for gh CLI if GitHub release is requested
    if [[ "$CREATE_GITHUB_RELEASE" == "true" ]]; then
        if ! command -v gh >/dev/null 2>&1; then
            error "GitHub CLI (gh) is not installed. Install from https://cli.github.com/"
        fi
        info "✓ GitHub CLI found: $(gh --version | head -n 1)"

        # Check gh authentication
        if ! gh auth status >/dev/null 2>&1; then
            error "GitHub CLI is not authenticated. Run 'gh auth login' first."
        fi
        info "✓ GitHub CLI authenticated"
    fi

    # Check working directory is clean (warn only)
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        warn "Working directory has uncommitted changes"
        warn "Consider committing or stashing changes before release"
    else
        info "✓ Working directory is clean"
    fi
}

# Clean release artifacts
clean() {
    info "Cleaning release artifacts..."
    rm -rf "${RELEASE_DIR}"
    rm -rf "${DIST_DIR}"
    info "Clean complete"
}

# Build for all platforms
build_all_platforms() {
    section "Building for All Platforms"

    # Call build script with dist mode
    DIST_MODE=true VERSION="v${VERSION}" "${SCRIPT_DIR}/build.sh" --dist

    info "✓ All platform builds complete"
}

# Create archive for a platform
create_archive() {
    local os=$1
    local arch=$2
    local archive_name="${BINARY_NAME}-v${VERSION}-${os}-${arch}"
    local source_dir="${DIST_DIR}/${os}-${arch}"

    if [[ ! -d "$source_dir" ]]; then
        warn "Source directory not found: $source_dir"
        return 1
    fi

    # Create release directory
    mkdir -p "${RELEASE_DIR}"

    # Create tar.gz for all platforms
    info "Creating archive: ${archive_name}.tar.gz"
    tar -czf "${RELEASE_DIR}/${archive_name}.tar.gz" -C "${DIST_DIR}" "${os}-${arch}"

    # Create zip for Windows
    if [[ "$os" == "windows" ]] && command -v zip >/dev/null 2>&1; then
        info "Creating archive: ${archive_name}.zip"
        (cd "${DIST_DIR}" && zip -q -r "${RELEASE_DIR}/${archive_name}.zip" "${os}-${arch}")
    fi

    info "✓ Created archive for ${os}/${arch}"
}

# Create all archives
create_archives() {
    section "Creating Release Archives"

    for platform in "${PLATFORMS[@]}"; do
        local os="${platform%/*}"
        local arch="${platform#*/}"
        create_archive "$os" "$arch"
    done

    info "✓ All archives created"
}

# Generate checksums
generate_checksums() {
    section "Generating Checksums"

    local checksum_file="${RELEASE_DIR}/checksums.txt"

    # Remove old checksum file
    rm -f "$checksum_file"

    # Generate checksums for all archives
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "${RELEASE_DIR}" && sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt)
    elif command -v shasum >/dev/null 2>&1; then
        (cd "${RELEASE_DIR}" && shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt)
    fi

    info "✓ Checksums generated: ${checksum_file}"
    echo ""
    cat "$checksum_file"
}

# Create changelog section
generate_changelog() {
    local version=$1
    local changelog_file="${RELEASE_DIR}/CHANGELOG.md"

    # Get commits since last tag
    local last_tag
    last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

    if [[ -n "$last_tag" ]]; then
        info "Generating changelog since ${last_tag}..."
        git log "${last_tag}..HEAD" --pretty=format:"- %s (%h)" > "$changelog_file"
    else
        info "No previous tag found. Generating changelog from first commit..."
        git log --pretty=format:"- %s (%h)" > "$changelog_file"
    fi

    # Add header
    {
        echo "# Release v${version}"
        echo ""
        echo "## Changes"
        echo ""
        cat "$changelog_file"
    } > "${changelog_file}.tmp"
    mv "${changelog_file}.tmp" "$changelog_file"

    info "✓ Changelog generated: ${changelog_file}"
}

# Create GitHub release
create_github_release() {
    section "Creating GitHub Release"

    local version="v${VERSION}"
    local release_flags=""

    if [[ "${DRAFT_RELEASE:-false}" == "true" ]]; then
        release_flags="--draft"
        info "Creating draft release..."
    fi

    if [[ "${PRERELEASE:-false}" == "true" ]]; then
        release_flags="$release_flags --prerelease"
        info "Marking as pre-release..."
    fi

    # Generate release notes from changelog
    local changelog_file="${RELEASE_DIR}/CHANGELOG.md"
    if [[ ! -f "$changelog_file" ]]; then
        generate_changelog "$VERSION"
    fi

    # Create release
    info "Creating GitHub release ${version}..."

    # Create release with changelog
    gh release create "$version" \
        --repo "$GITHUB_REPO" \
        --title "Release ${version}" \
        --notes-file "$changelog_file" \
        $release_flags \
        "${RELEASE_DIR}"/*.tar.gz \
        "${RELEASE_DIR}"/*.zip \
        "${RELEASE_DIR}"/checksums.txt

    if [[ $? -eq 0 ]]; then
        info "✓ GitHub release created: https://github.com/${GITHUB_REPO}/releases/tag/${version}"
    else
        error "Failed to create GitHub release"
    fi
}

# Show release summary
show_summary() {
    section "Release Summary"

    echo "  Version:        v${VERSION}"
    echo "  Release Dir:    ${RELEASE_DIR}"
    echo "  GitHub Release: $(if [[ "$CREATE_GITHUB_RELEASE" == "true" ]]; then echo "yes"; else echo "no"; fi)"
    echo ""

    info "Release artifacts:"
    find "${RELEASE_DIR}" -type f | sort | while read -r file; do
        local size
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "  - $(basename "$file") ($size)"
    done

    echo ""
    info "Release complete! 🎉"

    if [[ "$CREATE_GITHUB_RELEASE" == "true" ]]; then
        info "GitHub release: https://github.com/${GITHUB_REPO}/releases/tag/v${VERSION}"
    else
        info "To create GitHub release, run:"
        echo "    gh release create v${VERSION} --repo ${GITHUB_REPO} release/*"
    fi
}

# Main release function
main() {
    local clean_first=false
    local dry_run=false

    # Parse command-line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -g|--github)
                CREATE_GITHUB_RELEASE=true
                shift
                ;;
            --draft)
                DRAFT_RELEASE=true
                shift
                ;;
            --prerelease)
                PRERELEASE=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --clean)
                clean_first=true
                shift
                ;;
            *)
                error "Unknown option: $1\nUse --help for usage information"
                ;;
        esac
    done

    # Validate version
    validate_version "$VERSION"

    # Navigate to project root
    cd "$PROJECT_ROOT"

    # Check prerequisites
    check_prerequisites

    if [[ "$dry_run" == "true" ]]; then
        info "Dry run mode - no changes will be made"
        info "Would create release v${VERSION}"
        info "Would build for: ${PLATFORMS[*]}"
        if [[ "$CREATE_GITHUB_RELEASE" == "true" ]]; then
            info "Would create GitHub release"
        fi
        exit 0
    fi

    # Clean if requested
    if [[ "$clean_first" == "true" ]]; then
        clean
    fi

    # Create release
    build_all_platforms
    create_archives
    generate_checksums
    generate_changelog "$VERSION"

    # Create GitHub release if requested
    if [[ "$CREATE_GITHUB_RELEASE" == "true" ]]; then
        create_github_release
    fi

    # Show summary
    show_summary
}

# Run main function
main "$@"
