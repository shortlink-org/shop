# 3. OSRM as Routing Engine

Date: 2025-01-30

## Status

Accepted

## Context

For courier movement emulation, we need a mechanism for generating realistic routes:

- Routes must follow the real road network
- Accurate distances and travel times are required
- High performance is needed (generating thousands of routes)
- Self-hosted solution without external API dependencies is preferred

### Alternatives Considered

| Solution | Pros | Cons |
| -------- | ---- | ---- |
| **OSRM** | Fast, self-hosted, OSM data | Requires graph preprocessing |
| Google Maps API | Accurate data, traffic data | Paid, external dependency |
| GraphHopper | Flexible, Java-based | Slower than OSRM, harder to configure |
| Valhalla | Modern, isochrones | More complex deployment |
| Straight lines | Simple | Unrealistic routes |

## Decision

Use **OSRM (Open Source Routing Machine)** for route generation.

### Configuration

- **Region**: Berlin (test zone)
- **Profile**: `car.lua` (car transport)
- **Algorithm**: MLD (Multi-Level Dijkstra) — optimal for frequent queries
- **Deployment**: Docker container

### Data Preparation Process

```text
OSM PBF → osrm-extract → osrm-partition → osrm-customize → OSRM Server
```

1. **Extract** — parse OSM data, create road graph
2. **Partition** — split graph into cells for MLD
3. **Customize** — optimize weights for fast queries

### API Endpoints

```text
GET /route/v1/driving/{lon1},{lat1};{lon2},{lat2}?overview=full&steps=true
```

Response contains:

- `distance` — distance in meters
- `duration` — time in seconds
- `geometry` — encoded polyline
- `legs[].steps[]` — step-by-step instructions

### Berlin Bounding Box

```text
MinLat: 52.3383    MaxLat: 52.6755
MinLon: 13.0884    MaxLon: 13.7610
```

## Consequences

### Positive

- **Performance**: ~10-50ms per request, 10k routes in seconds
- **Realism**: routes on real roads with restrictions
- **Self-hosted**: no external dependencies, works offline
- **Free**: OSM data and OSRM are open source
- **Polyline**: compact format for route storage

### Negative

- **Data preparation**: ~5-10 minutes to build graph
- **Size**: Berlin graph ~500MB on disk
- **Freshness**: OSM data requires periodic updates
- **Roads only**: no pedestrian routes through parks (for `car.lua`)

### Migration to Another Region

To switch regions (e.g., Moscow):

```bash
# Download OSM data
wget https://download.geofabrik.de/russia/central-fed-district-latest.osm.pbf

# Rebuild graph
make osrm-build OSM_FILE=central-fed-district-latest.osm.pbf
```

## References

- [OSRM Documentation](http://project-osrm.org/)
- [OSRM Docker Hub](https://hub.docker.com/r/osrm/osrm-backend)
- [Geofabrik Downloads](https://download.geofabrik.de/)
- [Polyline Algorithm](https://developers.google.com/maps/documentation/utilities/polylinealgorithm)
