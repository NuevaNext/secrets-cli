# Dockerfile for secrets-cli end-to-end tests
# Uses debian:bookworm with all required dependencies

FROM golang:1.22-bookworm AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o secrets-cli ./cmd/secrets-cli

# Runtime image
FROM debian:bookworm-slim

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies (yq no longer needed - handled natively by Go)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gnupg \
    pass \
    git \
    bash \
    coreutils \
    findutils \
    grep \
    sed \
    gawk \
    && rm -rf /var/lib/apt/lists/*

# Create test user (non-root for realistic testing)
RUN useradd -m -s /bin/bash testuser

# Set up working directory
WORKDIR /workspace

# Copy the Go binary from builder
COPY --from=builder /build/secrets-cli /workspace/secrets-cli

# Copy test files
COPY tests/test-utils.sh /workspace/tests/test-utils.sh
COPY tests/e2e-tests.sh /workspace/tests/e2e-tests.sh

# Make scripts executable
RUN chmod +x /workspace/secrets-cli \
    && chmod +x /workspace/tests/test-utils.sh \
    && chmod +x /workspace/tests/e2e-tests.sh

# Change ownership to testuser
RUN chown -R testuser:testuser /workspace

# Switch to test user
USER testuser

# Set GPG to use loopback for pinentry (non-interactive)
RUN mkdir -p ~/.gnupg && chmod 700 ~/.gnupg \
    && echo "pinentry-mode loopback" >> ~/.gnupg/gpg.conf \
    && echo "allow-loopback-pinentry" >> ~/.gnupg/gpg-agent.conf

# Default command runs tests
ENTRYPOINT ["/workspace/tests/e2e-tests.sh"]
