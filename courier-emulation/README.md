# Courier Emulation Service

> [!NOTE]
> Courier emulation service for testing the Delivery Boundary. Simulates courier behavior: route navigation, location updates, order acceptance, and delivery confirmation.

## Features

- Automatic courier location updates via Kafka
- Route-based movement simulation using OSRM
- Automatic order assignment handling
- Delivery flow emulation
- Configurable simulation speed

## Quick Start

```bash
# Setup OSRM (see ADR-0003 for details)
make osrm-build
make osrm-run

# Run the service
make run
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OSRM_URL` | `http://localhost:5000` | OSRM routing server URL |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses |
| `SIMULATION_UPDATE_INTERVAL` | `5s` | Location update frequency |
| `SIMULATION_SPEED_KMH` | `30.0` | Courier speed in km/h |
| `SIMULATION_TIME_MULTIPLIER` | `1.0` | Time acceleration (2.0 = 2x speed) |

## Makefile Commands

```bash
make help       # Show all commands
make run        # Run locally
make test       # Run tests
make osrm-run   # Start OSRM server
```

## Architecture

- [ADR-0001: Init](docs/ADR/decisions/0001-init.md) — Service overview
- [ADR-0002: C4 System](docs/ADR/decisions/0002-c4-system.md) — Architecture diagrams
- [ADR-0003: OSRM Routing](docs/ADR/decisions/0003-osrm-routing.md) — Routing engine setup
