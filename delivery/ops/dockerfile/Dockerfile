# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

# ==============================================================================
# Chef stage - install cargo-chef for dependency caching
# ==============================================================================
FROM --platform=$BUILDPLATFORM rust:1.93-alpine AS chef

RUN apk add --no-cache musl-dev
RUN cargo install cargo-chef --locked
WORKDIR /app

# ==============================================================================
# Planner stage - analyze dependencies
# ==============================================================================
FROM chef AS planner

COPY Cargo.toml Cargo.lock* ./
COPY migration ./migration
COPY src ./src
COPY build.rs ./
COPY .cargo ./.cargo

RUN cargo chef prepare --recipe-path recipe.json

# ==============================================================================
# Builder stage - build dependencies first, then application
# ==============================================================================
FROM chef AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Install build dependencies
# Using system librdkafka for faster builds
RUN apk add --no-cache \
    musl-dev \
    protobuf-dev \
    bash \
    make \
    gcc \
    g++ \
    zlib-dev \
    zlib-static \
    librdkafka-dev

# Copy recipe and build dependencies (cached layer)
COPY --from=planner /app/recipe.json recipe.json
COPY .cargo ./.cargo
COPY migration ./migration
RUN cargo chef cook --release --recipe-path recipe.json

# Copy source and build application
COPY Cargo.toml Cargo.lock* ./
COPY migration ./migration
COPY src ./src
COPY build.rs ./

RUN cargo build --release

# ==============================================================================
# Final stage - minimal runtime image
# ==============================================================================
FROM alpine:3.23

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-delivery"
LABEL org.opencontainers.image.description="Delivery Service for ShortLink Shop"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

# Install runtime dependencies
# - ca-certificates: HTTPS connections
# - tini: proper signal handling
# - librdkafka: Kafka client library (dynamic linking)
RUN apk add --no-cache ca-certificates tini librdkafka

# gRPC port
EXPOSE 50051
# Metrics/health port
EXPOSE 9090

WORKDIR /app

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

# Copy the binary from builder
COPY --from=builder /app/target/release/delivery /app/delivery

# Change ownership
RUN chown -R app:app /app

USER app

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/delivery"]
