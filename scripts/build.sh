#!/usr/bin/env bash

# Build script for datagen-cli
# Supports cross-compilation for Linux, macOS, and Windows (amd64/arm64)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Build configuration
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
BUILD_DATE="${BUILD_DATE:-$(date -u '+%Y-%m-%d %H:%M:%S')}"
GO_VERSION="$(go version | awk '{print $3}')"

# Output directory
OUTPUT_DIR="${PROJECT_ROOT}/bin"
DIST_DIR="${PROJECT_ROOT}/dist"

# Binary name
BINARY_NAME="datagen"

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Ldflags for version injection
LDFLAGS="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.Commit=${COMMIT}' \
    -X 'main.BuildDate=${BUILD_DATE}' \
    -X 'main.GoVersion=${GO_VERSION}'"

# Help message
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Build datagen-cli for multiple platforms.

OPTIONS:
    -h, --help          Show this help message
    -v, --version VER   Set version (default: dev)
    -p, --platform PLATFORM
                        Build for specific platform (e.g., linux/amd64)
    -o, --output DIR    Output directory (default: bin/)
    -d, --dist          Build for distribution (output to dist/)
    --local             Build for local platform only
    --clean             Clean build artifacts before building

PLATFORMS:
    linux/amd64         Linux (64-bit)
    linux/arm64         Linux (ARM64)
    darwin/amd64        macOS (Intel)
    darwin/arm64        macOS (Apple Silicon)
    windows/amd64       Windows (64-bit)

EXAMPLES:
    # Build for local platform
    ./scripts/build.sh --local

    # Build for all platforms
    ./scripts/build.sh

    # Build for specific platform
    ./scripts/build.sh --platform linux/amd64

    # Build for distribution with version
    ./scripts/build.sh --dist --version 1.0.0

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

# Clean build artifacts
clean() {
    info "Cleaning build artifacts..."
    rm -rf "${OUTPUT_DIR}"
    rm -rf "${DIST_DIR}"
    info "Clean complete"
}

# Build for a specific platform
build_platform() {
    local platform=$1
    local os="${platform%/*}"
    local arch="${platform#*/}"

    local output_name="${BINARY_NAME}"
    local output_path="${OUTPUT_DIR}"

    # Use dist directory if specified
    if [[ "${DIST_MODE:-false}" == "true" ]]; then
        output_path="${DIST_DIR}"
    fi

    # Add .exe extension for Windows
    if [[ "$os" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi

    # Create platform-specific subdirectory for dist builds
    if [[ "${DIST_MODE:-false}" == "true" ]]; then
        output_path="${output_path}/${os}-${arch}"
        mkdir -p "$output_path"
    else
        mkdir -p "$output_path"
    fi

    local output_file="${output_path}/${output_name}"

    info "Building for ${os}/${arch}..."

    # Build command
    env GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags="${LDFLAGS}" \
        -o "$output_file" \
        "${PROJECT_ROOT}/cmd/datagen"

    if [[ $? -eq 0 ]]; then
        # Get file size
        local size
        if [[ "$os" == "darwin" ]]; then
            size=$(ls -lh "$output_file" | awk '{print $5}')
        else
            size=$(ls -lh "$output_file" | awk '{print $5}')
        fi

        info "✓ Built ${os}/${arch}: $output_file ($size)"

        # Copy additional files for dist builds
        if [[ "${DIST_MODE:-false}" == "true" ]]; then
            cp "${PROJECT_ROOT}/README.md" "$output_path/" 2>/dev/null || true
            cp "${PROJECT_ROOT}/LICENSE" "$output_path/" 2>/dev/null || true
            info "  Added README.md and LICENSE"
        fi

        return 0
    else
        error "✗ Failed to build ${os}/${arch}"
        return 1
    fi
}

# Build for local platform
build_local() {
    local os
    local arch

    case "$OSTYPE" in
        linux*)   os="linux" ;;
        darwin*)  os="darwin" ;;
        msys*|cygwin*) os="windows" ;;
        *)        error "Unsupported OS: $OSTYPE" ;;
    esac

    case "$(uname -m)" in
        x86_64)   arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)        error "Unsupported architecture: $(uname -m)" ;;
    esac

    build_platform "${os}/${arch}"
}

# Main build function
main() {
    local clean_first=false
    local local_only=false
    local specific_platform=""

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
            -p|--platform)
                specific_platform="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -d|--dist)
                DIST_MODE=true
                shift
                ;;
            --local)
                local_only=true
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

    # Navigate to project root
    cd "$PROJECT_ROOT"

    # Clean if requested
    if [[ "$clean_first" == "true" ]]; then
        clean
    fi

    # Print build configuration
    info "Build Configuration:"
    echo "  Version:    ${VERSION}"
    echo "  Commit:     ${COMMIT}"
    echo "  Build Date: ${BUILD_DATE}"
    echo "  Go Version: ${GO_VERSION}"
    echo "  Output:     ${OUTPUT_DIR}"
    echo ""

    # Build based on mode
    if [[ "$local_only" == "true" ]]; then
        info "Building for local platform..."
        build_local
    elif [[ -n "$specific_platform" ]]; then
        info "Building for specific platform: ${specific_platform}"
        build_platform "$specific_platform"
    else
        info "Building for all platforms..."
        local failed=0

        for platform in "${PLATFORMS[@]}"; do
            if ! build_platform "$platform"; then
                failed=$((failed + 1))
            fi
        done

        echo ""
        if [[ $failed -eq 0 ]]; then
            info "✓ All builds completed successfully"
        else
            warn "⚠ ${failed} build(s) failed"
            exit 1
        fi
    fi

    echo ""
    info "Build complete!"
    info "Binaries available in: ${OUTPUT_DIR}"

    # Show created files
    echo ""
    info "Created files:"
    if [[ "${DIST_MODE:-false}" == "true" ]]; then
        find "${DIST_DIR}" -type f -name "${BINARY_NAME}*" | sort
    else
        find "${OUTPUT_DIR}" -type f -name "${BINARY_NAME}*" | sort
    fi
}

# Run main function
main "$@"
