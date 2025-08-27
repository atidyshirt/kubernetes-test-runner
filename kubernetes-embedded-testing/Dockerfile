FROM ubuntu:24.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV NODE_VERSION=22

# Install system dependencies and tools
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    gnupg \
    ca-certificates \
    lsb-release \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Copy kubectl from official image
COPY --from=bitnami/kubectl:latest /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install mirrord using official installation script
RUN curl -fsSL https://raw.githubusercontent.com/metalbear-co/mirrord/main/scripts/install.sh | bash

# Create necessary directories
RUN mkdir -p /workspace /reports

# Set working directory
WORKDIR /workspace

# Verify installations
RUN kubectl version --client \
    && node --version \
    && npm --version \
    && mirrord --version

# Set default command
CMD ["/bin/bash"]
