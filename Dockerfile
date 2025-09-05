# Multi-stage Dockerfile for cross-platform kaf builds
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS base

# Install build dependencies
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    pkgconfig

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
RUN case ${TARGETARCH} in \
    "amd64") \
        apk add --no-cache librdkafka-dev \
        ;; \
    "arm64") \
        apk add --no-cache librdkafka-dev \
        ;; \
    *) \
        echo "Unsupported architecture: ${TARGETARCH}" && exit 1 \
        ;; \
    esac

# Build the binary
ENV CGO_ENABLED=1
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o kaf .

# Final minimal image
FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/kaf .

# Copy additional files
COPY README.md LICENSE ./

ENTRYPOINT ["./kaf"]