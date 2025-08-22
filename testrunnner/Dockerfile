# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o testrunner ./cmd/testrunner

# Runtime stage
FROM node:18-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    curl \
    bash \
    git \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# Create workspace directory
WORKDIR /workspace

# Copy the binary from builder stage
COPY --from=builder /app/testrunner /usr/local/bin/testrunner

# Create necessary directories
RUN mkdir -p /reports /tools

# Set the binary as executable
RUN chmod +x /usr/local/bin/testrunner

# Set default command
ENTRYPOINT ["/usr/local/bin/testrunner"]
CMD ["--help"]
