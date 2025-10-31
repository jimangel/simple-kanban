# Multi-stage build for minimal container size
# Optimized for security and small image size (<15MB)

# Build stage
FROM golang:1.25-alpine AS builder

# Add metadata
LABEL stage=builder

# Create non-root user in builder stage
RUN addgroup -g 1000 -S kanban && \
    adduser -u 1000 -S kanban -G kanban

WORKDIR /app

# Copy go mod files and download dependencies (cache layer)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with CGO disabled for static binary
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Install UPX for binary compression
RUN apk add --no-cache upx

# Build with optimizations to reduce binary size
# -s: strip symbol table
# -w: strip DWARF debug info
RUN go build -ldflags="-s -w" -o kanban-server ./cmd/server && \
    chmod +x kanban-server && \
    echo "Binary size before compression:" && \
    ls -lh kanban-server && \
    upx --best --lzma kanban-server && \
    echo "Binary size after UPX compression:" && \
    ls -lh kanban-server

# Create necessary directories with proper permissions
RUN mkdir -p /app/data && \
    chown -R kanban:kanban /app/data

# Runtime stage using scratch for minimal size
FROM scratch

# Add metadata
LABEL maintainer="kanban-simple"
LABEL version="1.0"
LABEL description="Lightweight Kanban board with SQLite backend"

# Copy user/group information from builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy CA certificates for HTTPS (if needed for external APIs)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory
WORKDIR /app

# Copy the binary from builder with proper ownership
COPY --from=builder --chown=kanban:kanban /app/kanban-server .

# Copy static files with proper ownership
COPY --from=builder --chown=kanban:kanban /app/web/static ./web/static

# Copy migrations with proper ownership
COPY --from=builder --chown=kanban:kanban /app/migrations ./migrations

# Create data directory with proper ownership
COPY --from=builder --chown=kanban:kanban /app/data ./data

# Switch to non-root user
USER kanban

# Create volume mount point for database
VOLUME ["/app/data"]

# Expose port (informational)
EXPOSE 8080

# Environment variables with secure defaults
ENV DATABASE_PATH=/app/data/kanban.db \
    MIGRATIONS_PATH=/app/migrations \
    PORT=8080 \
    GIN_MODE=release

# Health check endpoint (note: scratch doesn't have curl/wget)
# Health checks should be done externally or via docker-compose
# HEALTHCHECK NONE

# Run the server
ENTRYPOINT ["./kanban-server"]