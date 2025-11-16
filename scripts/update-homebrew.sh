#!/usr/bin/env bash

# Update Homebrew formula with release checksums
# This script updates the datagen.rb formula with SHA256 checksums
# from the release artifacts

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Paths
FORMULA_FILE="${PROJECT_ROOT}/homebrew/datagen.rb"
RELEASE_DIR="${PROJECT_ROOT}/release"
VERSION="${VERSION:-}"

# Help message
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Update Homebrew formula with release checksums.

OPTIONS:
    -h, --help              Show this help message
    -v, --version VERSION   Release version (required)
    -r, --release-dir DIR   Release directory (default: release/)
    -f, --formula FILE      Formula file (default: homebrew/datagen.rb)

EXAMPLES:
    # Update formula for v1.0.0 release
    ./scripts/update-homebrew.sh --version 1.0.0

    # Use custom release directory
    ./scripts/update-homebrew.sh --version 1.0.0 --release-dir /path/to/release

EOF
}

# Logger functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Get SHA256 checksum for a file
get_sha256() {
    local file=$1

    if [[ ! -f "$file" ]]; then
        error "File not found: $file"
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        error "Neither sha256sum nor shasum found"
    fi
}

# Update formula with checksums
update_formula() {
    local version=$1

    info "Updating Homebrew formula for v${version}..."

    # Extract checksums from checksums.txt if it exists
    local checksums_file="${RELEASE_DIR}/checksums.txt"
    if [[ ! -f "$checksums_file" ]]; then
        error "Checksums file not found: $checksums_file"
    fi

    # Read checksums
    declare -A shas
    while IFS= read -r line; do
        local sha=$(echo "$line" | awk '{print $1}')
        local file=$(echo "$line" | awk '{print $2}')

        # Extract platform from filename
        if [[ "$file" =~ darwin-amd64 ]]; then
            shas[darwin_amd64]=$sha
        elif [[ "$file" =~ darwin-arm64 ]]; then
            shas[darwin_arm64]=$sha
        elif [[ "$file" =~ linux-amd64 ]]; then
            shas[linux_amd64]=$sha
        elif [[ "$file" =~ linux-arm64 ]]; then
            shas[linux_arm64]=$sha
        fi
    done < "$checksums_file"

    # Validate all checksums are present
    for platform in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64; do
        if [[ -z "${shas[$platform]:-}" ]]; then
            error "Missing checksum for $platform"
        fi
    done

    # Update formula file
    info "Updating formula with checksums..."

    # Create temporary file
    local temp_file=$(mktemp)

    # Update version and checksums
    sed -e "s/version \".*\"/version \"${version}\"/" \
        -e "s|/v[0-9.]\+/|/v${version}/|g" \
        -e "s/PLACEHOLDER_SHA256_DARWIN_AMD64/${shas[darwin_amd64]}/" \
        -e "s/PLACEHOLDER_SHA256_DARWIN_ARM64/${shas[darwin_arm64]}/" \
        -e "s/PLACEHOLDER_SHA256_LINUX_AMD64/${shas[linux_amd64]}/" \
        -e "s/PLACEHOLDER_SHA256_LINUX_ARM64/${shas[linux_arm64]}/" \
        "$FORMULA_FILE" > "$temp_file"

    # Replace original file
    mv "$temp_file" "$FORMULA_FILE"

    info "✓ Formula updated: $FORMULA_FILE"

    # Show diff summary
    echo ""
    info "Updated checksums:"
    echo "  darwin-amd64: ${shas[darwin_amd64]}"
    echo "  darwin-arm64: ${shas[darwin_arm64]}"
    echo "  linux-amd64:  ${shas[linux_amd64]}"
    echo "  linux-arm64:  ${shas[linux_arm64]}"
}

# Main function
main() {
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
            -r|--release-dir)
                RELEASE_DIR="$2"
                shift 2
                ;;
            -f|--formula)
                FORMULA_FILE="$2"
                shift 2
                ;;
            *)
                error "Unknown option: $1\nUse --help for usage information"
                ;;
        esac
    done

    # Validate version
    if [[ -z "$VERSION" ]]; then
        error "Version is required. Use --version flag."
    fi

    # Remove 'v' prefix if present
    VERSION=${VERSION#v}

    # Navigate to project root
    cd "$PROJECT_ROOT"

    # Check if formula file exists
    if [[ ! -f "$FORMULA_FILE" ]]; then
        error "Formula file not found: $FORMULA_FILE"
    fi

    # Check if release directory exists
    if [[ ! -d "$RELEASE_DIR" ]]; then
        error "Release directory not found: $RELEASE_DIR"
    fi

    # Update formula
    update_formula "$VERSION"

    echo ""
    info "Homebrew formula updated successfully!"
    info "Next steps:"
    echo "  1. Review changes: git diff homebrew/datagen.rb"
    echo "  2. Commit formula: git add homebrew/datagen.rb && git commit -m 'Update Homebrew formula for v${VERSION}'"
    echo "  3. Push to tap repo (if separate)"
}

# Run main function
main "$@"
