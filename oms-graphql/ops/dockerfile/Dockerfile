# syntax=docker/dockerfile:1.23

ARG BUILDKIT_SBOM_SCAN_STAGE=true
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -o /oms-graphql \
    ./cmd/service

FROM --platform=$TARGETPLATFORM alpine:3.23

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-oms-graphql"
LABEL org.opencontainers.image.description="OMS GraphQL Subgraph - Cosmo Connect gRPC adapter for OMS"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

RUN apk add --no-cache ca-certificates tini

EXPOSE 4011

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /oms-graphql /app/oms-graphql

RUN chown -R app:app /app

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4011/healthz || exit 1

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/oms-graphql"]
