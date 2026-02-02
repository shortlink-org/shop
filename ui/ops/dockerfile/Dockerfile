# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

# =============================================================================
# Stage 1: Dependencies
# =============================================================================
FROM node:24.13.0-alpine AS deps

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy package files
COPY package.json pnpm-lock.yaml .npmrc* ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# =============================================================================
# Stage 2: Builder
# =============================================================================
FROM node:24.13.0-alpine AS builder

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Set build-time environment variables
ARG API_URI
ENV API_URI=${API_URI}

ARG NEXT_PUBLIC_API_URI
ENV NEXT_PUBLIC_API_URI=${NEXT_PUBLIC_API_URI}

ENV NEXT_TELEMETRY_DISABLED=1

# Build the application
RUN pnpm build

# Ensure public directory exists (Next.js optional directory)
RUN mkdir -p public

# =============================================================================
# Stage 3: Runner
# =============================================================================
FROM node:24.13.0-alpine AS runner

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-ui"
LABEL org.opencontainers.image.description="Shop UI for ShortLink"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

WORKDIR /app

# Install pnpm and tini
RUN apk add --no-cache tini && \
    corepack enable && corepack prepare pnpm@latest --activate

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy built assets
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Next.js prerender cache must be writable at runtime (non-root user)
RUN mkdir -p .next/cache && chown -R nextjs:nodejs .next

USER nextjs

# Next.js port
EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["node", "server.js"]
