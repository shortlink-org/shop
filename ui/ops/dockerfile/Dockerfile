# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM node:24.14.0-alpine AS builder

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy package files first (better layer caching)
COPY package.json pnpm-lock.yaml .npmrc* ./

# Fetch + offline install using cache mount for pnpm store (no node_modules copy between stages)
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm fetch --frozen-lockfile && pnpm install --offline --frozen-lockfile

# Copy application source
COPY . .

# Set build-time environment variables
ARG API_URI
ENV API_URI=${API_URI}

ARG NEXT_PUBLIC_API_URI
ENV NEXT_PUBLIC_API_URI=${NEXT_PUBLIC_API_URI}

ARG NEXT_PUBLIC_LOGIN_URL
ENV NEXT_PUBLIC_LOGIN_URL=${NEXT_PUBLIC_LOGIN_URL}

ENV NEXT_TELEMETRY_DISABLED=1

# Build the application
RUN pnpm build

# =============================================================================
# Stage 2: Runner
# =============================================================================
FROM node:24.14.0-alpine AS runner

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-ui"
LABEL org.opencontainers.image.description="Shop UI for ShortLink"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

WORKDIR /app

# Install tini (init for node process)
RUN apk add --no-cache tini

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Use base image's node user (UID/GID 1000). Do not run addgroup/adduser with 1000 — it's already in use in node:* Alpine.

# Copy built assets (this app has no public/ — static assets are in app/; create empty public for runtime)
RUN mkdir -p public
COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static

# Next.js prerender cache must be writable at runtime (non-root user); fetch-cache is created on first request
RUN mkdir -p .next/cache/fetch-cache .next/cache/fetch-cache-tmp && chown -R node:node .next public

USER node

# Next.js port
EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"
# Writable temp for Next/Node; if using readOnlyRootFilesystem, mount tmpfs on /tmp
ENV TMPDIR=/tmp

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["node", "server.js"]
