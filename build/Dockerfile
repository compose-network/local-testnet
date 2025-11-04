# Multi-stage Dockerfile for local-testnet
# Stage 1: Build the Go binary

FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o localnet ./cmd/localnet

# Stage 2: Runtime image with all dependencies

FROM ubuntu:22.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    # Docker CLI for running docker compose
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    wget \
    # Git for cloning repositories
    git \
    # jq for JSON processing in contract scripts
    jq \
    # Just - command runner for contract setup
    && rm -rf /var/lib/apt/lists/*

# Install just (command runner)
RUN wget -qO - 'https://proget.makedeb.org/debian-feeds/prebuilt-mpr.pub' | gpg --dearmor | tee /usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg 1> /dev/null && \
    echo "deb [arch=all,$(dpkg --print-architecture) signed-by=/usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg] https://proget.makedeb.org prebuilt-mpr $(lsb_release -cs)" | tee /etc/apt/sources.list.d/prebuilt-mpr.list && \
    apt-get update && \
    apt-get install -y just && \
    rm -rf /var/lib/apt/lists/*

# Install Docker CLI
RUN install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && \
    apt-get install -y docker-ce-cli docker-compose-plugin && \
    rm -rf /var/lib/apt/lists/*

# Install Foundry (for Solidity compilation)
RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="/root/.foundry/bin:${PATH}"
RUN foundryup

# Copy the binary from builder stage
COPY --from=builder /build/localnet /usr/local/bin/localnet

# Create working directory for runtime data
WORKDIR /workspace

# Default command
ENTRYPOINT ["localnet"]
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.source="https://github.com/compose-network/local-testnet"
LABEL org.opencontainers.image.description="Local testnet control plane for L1 and L2 Ethereum networks"
LABEL org.opencontainers.image.licenses="Apache-2.0"
