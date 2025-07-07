# Multi-stage build with Python base image for reliability
FROM --platform=$BUILDPLATFORM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code (including embedded assets)
COPY . .

# Build the binary with proper architecture flags
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Disable CGO for cross-compilation (using pure Go SQLite driver)
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Build the binary with architecture-specific optimizations
RUN go build -ldflags="-s -w" -o zendo ./main.go

# Make the binary executable
RUN chmod +x /app/zendo

# Final stage - use minimal alpine image
FROM --platform=$TARGETPLATFORM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    tzdata

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy the binary
COPY --from=builder /app/zendo /app/zendo

# Set working directory
WORKDIR /app

# Create storage directory for database and set permissions
RUN mkdir -p /app/storage && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set environment variables with defaults
ENV TZ="America/Los_Angeles"

# Set the entrypoint to run the binary
ENTRYPOINT ["/app/zendo"] 

