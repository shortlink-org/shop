# syntax=docker/dockerfile:1.21
# delivery-graphql Dockerfile
# Uses Tailcall for gRPC → GraphQL

FROM ghcr.io/tailcallhq/tailcall/tc-server:latest

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-delivery-graphql"
LABEL org.opencontainers.image.description="Delivery GraphQL Subgraph - gRPC to GraphQL via Tailcall"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

# Set working directory
WORKDIR /app

# Copy Tailcall configuration
COPY config/ /app/config/

# Environment variables
ENV DELIVERY_GRPC_URL="http://localhost:50051"

# Expose port
EXPOSE 8102

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8102/graphql || exit 1

# Run Tailcall (full path: PATH contains ~/.tailcall/bin which is not expanded in exec without shell)
CMD ["/root/.tailcall/bin/tailcall", "start", "/app/config/delivery.graphql"]
