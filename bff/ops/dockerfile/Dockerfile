# Cosmo Router Dockerfile
# https://cosmo-docs.wundergraph.com/deployments-and-hosting/docker

FROM ghcr.io/wundergraph/cosmo/router:latest

# Set working directory
WORKDIR /app

# Copy router configuration
COPY router.yaml /app/router.yaml
COPY router-config.json /app/router-config.json

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
