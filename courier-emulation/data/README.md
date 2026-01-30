# Data Directory

This directory contains OSM data and generated OSRM routing files.

## Files

- `berlin-latest.osm.pbf` — Berlin OpenStreetMap data (download from [Geofabrik](https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf))
- `berlin-latest.osrm*` — Generated OSRM routing graph files (created by `make osrm-build`)

## Download OSM Data

```bash
wget https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf
```

## Build Routing Graph

```bash
make osrm-build
```
