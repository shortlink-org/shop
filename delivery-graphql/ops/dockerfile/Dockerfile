# syntax=docker/dockerfile:1.21

ARG BUILDKIT_SBOM_SCAN_STAGE=true
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -o /delivery-graphql \
    ./cmd/service

FROM --platform=$TARGETPLATFORM alpine:3.23

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-delivery-graphql"
LABEL org.opencontainers.image.description="Delivery GraphQL Subgraph - Cosmo Connect gRPC adapter for Delivery"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

RUN apk add --no-cache ca-certificates tini

EXPOSE 4013

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /delivery-graphql /app/delivery-graphql

RUN chown -R app:app /app

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4013/healthz || exit 1

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/delivery-graphql"]
