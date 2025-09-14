# Multi-stage Dockerfile for cross-platform kafy builds
FROM --platform=$BUILDPLATFORM golang:1.21 AS base

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    build-essential \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build stage for different architectures
FROM base AS builder
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Install cross-compilation tools and librdkafka for the target architecture
RUN case ${TARGETARCH} in \
    "amd64") \
        apt-get update && apt-get install -y \
            librdkafka-dev \
            gcc \
            && rm -rf /var/lib/apt/lists/* \
        ;; \
    "arm64") \
        apt-get update && apt-get install -y \
            librdkafka-dev \
            gcc-aarch64-linux-gnu \
            libc6-dev-arm64-cross \
            && rm -rf /var/lib/apt/lists/* \
        ;; \
    *) \
        echo "Unsupported architecture: ${TARGETARCH}" && exit 1 \
        ;; \
    esac

# Set up cross-compilation environment for ARM64
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        ln -s /usr/aarch64-linux-gnu/lib/librdkafka.so.1 /usr/lib/aarch64-linux-gnu/librdkafka.so.1 || true; \
        export CC=aarch64-linux-gnu-gcc; \
        export CXX=aarch64-linux-gnu-g++; \
        export AR=aarch64-linux-gnu-ar; \
        export STRIP=aarch64-linux-gnu-strip; \
        export PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig; \
    fi

# Build the binary
ENV CGO_ENABLED=1
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Set cross-compilation environment variables for ARM64
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        export CC=aarch64-linux-gnu-gcc && \
        export CXX=aarch64-linux-gnu-g++ && \
        export AR=aarch64-linux-gnu-ar && \
        export STRIP=aarch64-linux-gnu-strip && \
        export PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig && \
        go build -o kafy .; \
    else \
        go build -o kafy .; \
    fi

# Final minimal image
FROM ubuntu:22.04 AS final
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/kafy .

# Copy additional files
COPY README.md LICENSE ./

ENTRYPOINT ["./kafy"]