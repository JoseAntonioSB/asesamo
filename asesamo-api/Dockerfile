# ==============================================================================
# ASESAMO API — Production Multi-Stage Dockerfile
# Service: asesamo-api (Go Clean Architecture REST API)
# ==============================================================================

# ------------------------------------------------------------------------------
# Stage 1: Build & Dependency Resolution
# ------------------------------------------------------------------------------
FROM golang:1.27.1-alpine AS builder

# Disable cgo for fully static compilation and ensure target OS is Linux
ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

# Step 1: Cache Go modules layer
# Copy only dependency definitions first to maximize Docker layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Step 2: Copy application source code
COPY . .

# Step 3: Build statically linked binary
# -trimpath strips file system paths from compiled binary for reproducibility and security
# -ldflags="-w -s" strips DWARF debugging information and symbol tables for minimal footprint
RUN go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /bin/api \
    ./cmd/api

# ------------------------------------------------------------------------------
# Stage 2: Minimal Hardened Runtime
# ------------------------------------------------------------------------------
FROM alpine:3.19 AS runner

# Install essential runtime utilities:
# - ca-certificates: TLS/HTTPS communication with external services (R2, etc.)
# - tzdata: accurate timezone processing
# - curl: container healthchecks
RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl && \
    rm -rf /var/cache/apk/*

# Create unprivileged system group and user
RUN addgroup -g 10001 -S asesamo && \
    adduser -u 10001 -S -G asesamo -s /sbin/nologin asesamo

WORKDIR /app

# Copy binary from builder stage with strict ownership
COPY --from=builder --chown=asesamo:asesamo /bin/api /bin/api

# Switch to unprivileged user
USER asesamo:asesamo

# Expose HTTP port
EXPOSE 8080

# Configure container healthcheck
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Launch application
ENTRYPOINT ["/bin/api"]
