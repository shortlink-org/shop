# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

# Build stage
FROM --platform=$BUILDPLATFORM rust:1.93-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /app

# Install build dependencies
# bash, make, gcc, g++ are required for rdkafka-sys to build librdkafka from source
RUN apk add --no-cache musl-dev protobuf-dev bash make gcc g++

# Create a new empty shell project for caching dependencies
RUN cargo new --bin delivery
WORKDIR /app/delivery

# Copy manifests
COPY Cargo.toml Cargo.lock* ./
COPY .cargo ./.cargo

# Copy migration crate (dev-dependency, but Cargo needs to resolve it)
COPY migration ./migration

# Build dependencies only (for caching)
RUN cargo build --release || true
RUN rm -rf src

# Copy source code
COPY src ./src
COPY build.rs ./

# Build the application
RUN cargo build --release

# Final stage
FROM alpine:3.21

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-delivery"
LABEL org.opencontainers.image.description="Delivery Service for ShortLink Shop"
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
COPY --from=builder /app/delivery/target/release/delivery /app/delivery

# Change ownership
RUN chown -R app:app /app

USER app

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/delivery"]
