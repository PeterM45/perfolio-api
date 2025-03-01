# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty)" \
    -o perfolio-api cmd/api/main.go

# Runtime stage
FROM alpine:3.19

# Add CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -H -h /app appuser

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/perfolio-api .

# Copy config files
COPY --from=builder /app/configs/config.yaml ./configs/

# Set ownership
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Set environment variables
ENV GIN_MODE=release

# Expose API port
EXPOSE 8080

# Command to run
ENTRYPOINT ["/app/perfolio-api"]

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1