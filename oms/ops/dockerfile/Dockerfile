# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -o /oms \
    ./cmd

# Final stage
FROM --platform=$TARGETPLATFORM alpine:3.23

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-oms"
LABEL org.opencontainers.image.description="Order Management System for ShortLink Shop"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

# Install ca-certificates for HTTPS and tini for proper signal handling
RUN apk add --no-cache ca-certificates tini

# gRPC port
EXPOSE 50051
# Metrics/health port
EXPOSE 9090

WORKDIR /app

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

# Copy the binary from builder
COPY --from=builder /oms /app/oms

# Change ownership
RUN chown -R app:app /app

USER app

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/oms"]
