# ====================
# Stage 1: Build
# ====================
FROM golang:1.25-alpine AS builder

# Change to China mirror for apk
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Set GOPROXY for China
ENV GOPROXY=https://goproxy.cn,direct

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" for smaller binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/mcp-server \
    ./cmd/server

# ====================
# Stage 2: Runtime
# ====================
FROM alpine:3.20

# Change to China mirror for apk
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 -S mcp && \
    adduser -u 1000 -S mcp -G mcp

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mcp-server /app/mcp-server

# Copy config directory if needed (optional, can be mounted)
# COPY --from=builder /app/config /app/config

# Change ownership
RUN chown -R mcp:mcp /app

# Switch to non-root user
USER mcp

# Expose port
EXPOSE 11611

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:11611/health || exit 1

# Run the application
ENTRYPOINT ["/app/mcp-server"]
