# syntax=docker/dockerfile:1.21
# admin-graphql Dockerfile
# Uses Tailcall for REST → GraphQL

FROM ghcr.io/tailcallhq/tailcall/tc-server:latest

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-admin-graphql"
LABEL org.opencontainers.image.description="Admin GraphQL Subgraph - REST to GraphQL via Tailcall"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

# Set working directory
WORKDIR /app

# Copy Tailcall configuration
COPY config/admin.graphql /app/config/admin.graphql

# Environment variables
ENV ADMIN_API_URL="http://127.0.0.1:8000"

# Expose port
EXPOSE 8101

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8101/graphql || exit 1

# Run Tailcall (full path: PATH contains ~/.tailcall/bin which is not expanded in exec without shell)
CMD ["/root/.tailcall/bin/tailcall", "start", "/app/config/admin.graphql"]
