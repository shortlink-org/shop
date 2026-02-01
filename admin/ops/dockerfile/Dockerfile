# syntax=docker/dockerfile:1.21

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

FROM ghcr.io/astral-sh/uv:python3.14-bookworm

LABEL maintainer=batazor111@gmail.com
LABEL org.opencontainers.image.title="shortlink-shop-admin"
LABEL org.opencontainers.image.description="shortlink-shop-admin"
LABEL org.opencontainers.image.authors="Login Viktor @batazor"
LABEL org.opencontainers.image.vendor="Login Viktor @batazor"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="http://shortlink.best/"
LABEL org.opencontainers.image.source="https://github.com/shortlink-org/shortlink"

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV VIRTUAL_ENV=/app/.venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"
ENV UV_CACHE_DIR=/home/shop/.cache/uv
ENV PYTHONPATH=/app/src

# HTTP API && Prometheus metrics
EXPOSE 8000

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      curl sqlite3 git \
      # GeoDjango dependencies (GDAL, GEOS, PROJ)
      gdal-bin libgdal-dev libgeos-dev libproj-dev && \
    rm -rf /var/lib/apt/lists/*

# GDAL/GEOS library paths for Django GeoDjango (Debian x86_64)
# These must match the symlinks created by libgdal-dev
ENV GDAL_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/libgdal.so
ENV GEOS_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/libgeos_c.so

RUN addgroup --system --gid 1000 shop && adduser --system --uid 1000 --gid 1000 shop

WORKDIR /app

RUN chown shop:shop /app && \
    mkdir -p /home/shop/.cache/uv && \
    chown -R shop:shop /home/shop/.cache

COPY --chown=shop:shop pyproject.toml uv.lock ./

USER shop

RUN uv sync --frozen --no-install-project

COPY --chown=shop:shop . .

# Collect static files for WhiteNoise
RUN python src/manage.py collectstatic --noinput

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=20s \
  CMD curl -f http://localhost:8000/healthz || exit 1

CMD ["/app/.venv/bin/gunicorn", "admin.wsgi:application", "-b", "0.0.0.0:8000", "--workers", "2", "--threads", "4", "--timeout", "60", "--graceful-timeout", "30", "--worker-tmp-dir", "/dev/shm"]
