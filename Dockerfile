# Multi-stage Dockerfile for cross-platform kaf builds
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

# Install librdkafka for the target architecture
RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

# Build the binary
ENV CGO_ENABLED=1
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN go build -o kaf .

# Final minimal image
FROM ubuntu:22.04 AS final
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/kaf .

# Copy additional files
COPY README.md LICENSE ./

ENTRYPOINT ["./kaf"]