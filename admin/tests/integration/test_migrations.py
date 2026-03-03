"""Integration tests for Django migrations using Testcontainers.

Runs all migrations against a fresh Postgres+PostGIS container to ensure
migrations apply cleanly (e.g. auth, contenttypes, goods, offices).
"""

import os
import subprocess
import sys

import pytest


@pytest.fixture(scope="module")
def postgres_container():
    """Start a Postgres container with PostGIS for migration tests."""
    from testcontainers.postgres import PostgresContainer

    with PostgresContainer(
        image="postgis/postgis:16-3.4-alpine",
        username="postgres",
        password="shortlink",
        dbname="shortlink",
    ) as postgres:
        yield postgres


def test_migrations_apply(postgres_container):
    """Run Django migrations against a real Postgres+PostGIS; must succeed."""
    host = postgres_container.get_container_host_ip()
    port = postgres_container.get_exposed_port(5432)

    env = os.environ.copy()
    env["DJANGO_SETTINGS_MODULE"] = "admin.settings_migration"
    env["POSTGRES_HOST"] = host
    env["POSTGRES_PORT"] = str(port)
    env["POSTGRES_DB"] = "shortlink"
    env["POSTGRES_USER"] = "postgres"
    env["POSTGRES_PASSWORD"] = "shortlink"
    env["POSTGRES_SSLMODE"] = "disable"  # testcontainer does not use SSL

    # Run from admin directory so src/migration.py and Python path resolve
    admin_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    result = subprocess.run(
        [sys.executable, "src/migration.py", "migrate", "--no-input"],
        cwd=admin_dir,
        env=env,
        capture_output=True,
        text=True,
        timeout=120,
    )

    assert result.returncode == 0, (
        f"migrate failed (exit {result.returncode})\n"
        f"stdout:\n{result.stdout}\nstderr:\n{result.stderr}"
    )
