# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ket ./cmd/testrunner

# Runtime stage
FROM node:18-alpine

# Install runtime dependencies and mirrord
RUN apk add --no-cache \
    curl \
    bash \
    git \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# Install mirrord for multiple architectures
RUN mkdir -p /tools && \
    # Download x86_64 version
    curl -L -s "https://github.com/metalbear-co/mirrord/releases/latest/download/mirrord-linux-x86_64" -o /tools/mirrord-x86_64 && \
    # Download ARM64 version  
    curl -L -s "https://github.com/metalbear-co/mirrord/releases/latest/download/mirrord-linux-aarch64" -o /tools/mirrord-aarch64 || echo "ARM64 version not available" && \
    # Set up architecture detection and symlink
    echo '#!/bin/sh' > /tools/mirrord && \
    echo 'ARCH=$(uname -m)' >> /tools/mirrord && \
    echo 'if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then' >> /tools/mirrord && \
    echo '    exec /tools/mirrord-x86_64 "$@"' >> /tools/mirrord && \
    echo 'elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then' >> /tools/mirrord && \
    echo '    if [ -f "/tools/mirrord-aarch64" ]; then' >> /tools/mirrord && \
    echo '        exec /tools/mirrord-aarch64 "$@"' >> /tools/mirrord && \
    echo '    else' >> /tools/mirrord && \
    echo '        echo "Warning: mirrord not available for ARM64, tests will run without traffic interception"' >> /tools/mirrord && \
    echo '        exit 0' >> /tools/mirrord && \
    echo '    fi' >> /tools/mirrord && \
    echo 'else' >> /tools/mirrord && \
    echo '    echo "Warning: Unknown architecture $ARCH, mirrord may not work"' >> /tools/mirrord && \
    echo '    exec /tools/mirrord-x86_64 "$@"' >> /tools/mirrord && \
    echo 'fi' >> /tools/mirrord && \
    chmod +x /tools/mirrord && \
    chmod +x /tools/mirrord-x86_64 && \
    chmod +x /tools/mirrord-aarch64 2>/dev/null || true

# Create workspace directory
WORKDIR /workspace

# Copy the binary from builder stage
COPY --from=builder /app/ket /usr/local/bin/ket

# Create necessary directories
RUN mkdir -p /reports

# Set the binary as executable
RUN chmod +x /usr/local/bin/ket

# Set default command
ENTRYPOINT ["/usr/local/bin/ket"]
CMD ["--help"]
