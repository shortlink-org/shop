# syntax=docker/dockerfile:1.21

# =============================================================================
# Stage 1: Download OSM data using modern image
# =============================================================================
FROM alpine:3.21 AS downloader

ARG REGION_URL=https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf
ARG REGION_NAME=berlin-latest

RUN apk add --no-cache wget
RUN mkdir -p /data && wget -q ${REGION_URL} -O /data/${REGION_NAME}.osm.pbf

# =============================================================================
# Stage 2: Build OSRM graph
# =============================================================================
FROM osrm/osrm-backend:latest AS builder

ARG REGION_NAME=berlin-latest

WORKDIR /data

# Copy downloaded OSM data from downloader stage
COPY --from=downloader /data/${REGION_NAME}.osm.pbf /data/

# Extract: parse OSM data, create road graph
RUN osrm-extract -p /opt/car.lua /data/${REGION_NAME}.osm.pbf

# Partition: split graph into cells for MLD algorithm
RUN osrm-partition /data/${REGION_NAME}.osrm

# Customize: optimize weights for fast queries
RUN osrm-customize /data/${REGION_NAME}.osrm

# Remove source PBF to save space
RUN rm -f /data/${REGION_NAME}.osm.pbf

# =============================================================================
# Stage 3: Runtime image with pre-built graph
# =============================================================================
FROM osrm/osrm-backend:latest

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-osrm"
LABEL org.opencontainers.image.description="OSRM Routing Server with Berlin data"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

ARG REGION_NAME=berlin-latest

WORKDIR /data

# Copy only the processed OSRM files (not the source PBF)
COPY --from=builder /data/${REGION_NAME}.osrm* /data/

EXPOSE 5000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:5000/health || exit 1

# Run OSRM routing server with MLD algorithm
CMD ["osrm-routed", "--algorithm=MLD", "--max-table-size=1000", "/data/berlin-latest.osrm"]
