# Multi-stage build for Cosmo Router
# https://cosmo-docs.wundergraph.com/deployments-and-hosting/docker

# =============================================================================
# Stage 1: Generate router-config.json (offline composition)
# =============================================================================
FROM node:22-alpine AS builder

WORKDIR /build

# Install Cosmo CLI
RUN npm install -g wgc@latest

# Copy schema files and configuration
COPY schemas/ schemas/
COPY graph-static.yaml .

# Compose the federated schema from static SDL files
RUN wgc router compose -i graph-static.yaml -o router-config.json

# =============================================================================
# Stage 2: Final router image
# =============================================================================
FROM ghcr.io/wundergraph/cosmo/router:latest

WORKDIR /app

# Copy router configuration
COPY router.yaml /app/router.yaml
COPY --from=builder /build/router-config.json /app/router-config.json

# Environment variables
ENV LISTEN_ADDR="0.0.0.0:9991"
ENV LOG_LEVEL="info"

# Expose ports
# 9991 - GraphQL API
# 8088 - Prometheus metrics
EXPOSE 9991 8088

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:9991/health || exit 1

# Run the router
CMD ["--config", "/app/router.yaml"]
