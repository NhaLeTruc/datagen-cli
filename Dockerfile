# Multi-stage Dockerfile for datagen-cli
# Builds a minimal Alpine-based image (~10MB)

# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# - Static binary with no C dependencies (CGO_ENABLED=0)
# - Strip debug info (-ldflags="-s -w")
# - Trimpath for reproducible builds
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w \
        -X 'main.Version=${VERSION}' \
        -X 'main.Commit=${COMMIT}' \
        -X 'main.BuildDate=${BUILD_DATE}'" \
    -o /build/datagen \
    ./cmd/datagen

# Stage 2: Runtime stage
FROM alpine:3.19

# Install runtime dependencies
# - ca-certificates: for HTTPS connections
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 datagen && \
    adduser -D -u 1000 -G datagen datagen

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/datagen /usr/local/bin/datagen

# Copy templates and examples
COPY --from=builder /build/internal/templates /app/templates
COPY --from=builder /build/docs/examples /app/examples

# Set ownership
RUN chown -R datagen:datagen /app

# Switch to non-root user
USER datagen

# Set environment variables
ENV DATAGEN_HOME=/app

# Healthcheck (simple version check)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD datagen version || exit 1

# Default command (show help)
ENTRYPOINT ["datagen"]
CMD ["--help"]

# Labels (OCI annotations)
LABEL org.opencontainers.image.title="datagen-cli" \
      org.opencontainers.image.description="Generate PostgreSQL dump files from JSON schemas" \
      org.opencontainers.image.vendor="NhaLeTruc" \
      org.opencontainers.image.source="https://github.com/NhaLeTruc/datagen-cli" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}"
