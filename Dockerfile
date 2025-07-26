# Build stage
FROM golang:1.23 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o snapshot-cosmos .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S snapshot && \
    adduser -u 1001 -S snapshot -G snapshot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/snapshot-cosmos .

# Copy config directory
COPY --from=builder /app/config ./config

# Change ownership to non-root user
RUN chown -R snapshot:snapshot /app

# Switch to non-root user
USER snapshot

# Set entrypoint
ENTRYPOINT ["./snapshot-cosmos"] 
