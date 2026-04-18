# syntax=docker/dockerfile:1.23

# Link: https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md
# enable scanning for the intermediate build stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true
# scan the build context only if the build is run to completion
ARG BUILDKIT_SBOM_SCAN_CONTEXT=true

FROM ghcr.io/astral-sh/uv:python3.14-bookworm AS builder

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV VIRTUAL_ENV=/app/.venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"
ENV UV_CACHE_DIR=/home/shop/.cache/uv
ENV PYTHONPATH=/app/src

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      curl sqlite3 git \
      build-essential pkg-config ca-certificates \
      # GeoDjango dependencies (GDAL, GEOS, PROJ)
      gdal-bin libgdal-dev libgeos-dev libproj-dev && \
    rm -rf /var/lib/apt/lists/*

# rpds-py ships manylinux wheels, but uv may fall back to the sdist (maturin) when no wheel
# matches the interpreter ABI (e.g. free-threaded Python). Install Rust so PEP 517 builds succeed.
ENV RUSTUP_HOME=/usr/local/rustup
ENV CARGO_HOME=/usr/local/cargo
ENV PATH="/usr/local/cargo/bin:${PATH}"
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain stable

RUN addgroup --system --gid 1000 shop && adduser --system --uid 1000 --gid 1000 shop

WORKDIR /app

RUN chown shop:shop /app && \
    mkdir -p /home/shop/.cache/uv /home/shop/.local/share/uv && \
    chown -R shop:shop /home/shop

COPY --chown=shop:shop pyproject.toml uv.lock ./

USER shop
ENV HOME=/home/shop

RUN uv sync --frozen --no-install-project

COPY --chown=shop:shop . .

# Collect static files for WhiteNoise
# Use uv run to ensure correct virtual environment
RUN uv run python src/manage.py collectstatic --noinput

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
ENV PATH="/app/.venv/bin:$PATH"
ENV PYTHONPATH=/app/src

EXPOSE 8000

# Runtime-only OS deps (no Rust/gcc/-dev headers).
# gdal-bin pulls libgdal; libgeos-c1v5 / libproj25 are explicit so GeoDjango/GDAL always finds GEOS & PROJ.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      curl sqlite3 ca-certificates \
      gdal-bin \
      libgeos-c1v5 \
      libproj25 \
    && rm -rf /var/lib/apt/lists/*

RUN addgroup --system --gid 1000 shop && adduser --system --uid 1000 --gid 1000 shop

WORKDIR /app

RUN mkdir -p /home/shop && chown -R shop:shop /home/shop

COPY --from=builder --chown=shop:shop /app /app

USER shop
ENV HOME=/home/shop

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=20s \
  CMD curl -f http://localhost:8000/healthz || exit 1

CMD ["/app/.venv/bin/gunicorn", "admin.wsgi:application", "-b", "0.0.0.0:8000", "--workers", "2", "--threads", "4", "--timeout", "60", "--graceful-timeout", "30", "--worker-tmp-dir", "/dev/shm"]
