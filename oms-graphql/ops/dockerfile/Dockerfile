# syntax=docker/dockerfile:1.21
# oms-graphql Dockerfile
# Uses Tailcall for gRPC → GraphQL

FROM ghcr.io/tailcallhq/tailcall/tc-server:latest

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-oms-graphql"
LABEL org.opencontainers.image.description="OMS GraphQL Subgraph - gRPC to GraphQL via Tailcall"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

# Set working directory
WORKDIR /app

# Copy tailcall binary to /app so it works when container runs as non-root (e.g. Kubernetes runAsNonRoot)
RUN mkdir -p /app/bin && cp /root/.tailcall/bin/tailcall /app/bin/tailcall && chmod 755 /app/bin/tailcall

# Copy Tailcall configuration
COPY config/ /app/config/

# Environment variables
ENV OMS_GRPC_URL="http://localhost:50051"

# Expose port
EXPOSE 8100

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8100/graphql || exit 1

CMD ["/app/bin/tailcall", "start", "/app/config/grpc.graphql"]
