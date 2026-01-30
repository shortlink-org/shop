# Offices Domain

This module manages pickup office locations for the shop.

## Overview

Offices are physical locations where customers can pick up their orders. Each office has:
- Geographic coordinates (using GeoDjango PointField)
- Address information
- Working hours
- Contact information

## GeoDjango Integration

This module uses GeoDjango for geospatial features:
- `PointField` for storing office locations
- Leaflet map widget for visual location selection in admin

### Requirements

- PostgreSQL with PostGIS extension
- GDAL/GEOS libraries installed on the system

### Database Setup

Ensure PostGIS extension is enabled in your PostgreSQL database:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
```

## Initial Data

The migration `0002_berlin_offices` creates two test offices in central Berlin:

1. **ShortLink Berlin Mitte** - Near Alexanderplatz
   - Address: Alexanderplatz 1, 10178 Berlin
   - Coordinates: 52.5219, 13.4132
   - Hours: Mon-Sat 08:00-20:00

2. **ShortLink Berlin Potsdamer Platz** - Near Potsdamer Platz
   - Address: Potsdamer Platz 5, 10785 Berlin
   - Coordinates: 52.5096, 13.3761
   - Hours: Mon-Sun 09:00-21:00

These locations are in central Berlin to match the courier emulator which uses Berlin OSRM data.

## Admin Interface

The admin interface provides:
- Interactive Leaflet map for location selection
- List view with working hours and status
- Filtering by active status and working days
- Search by name, address, phone, email

## Model Fields

| Field | Type | Description |
|-------|------|-------------|
| name | CharField | Display name of the office |
| address | TextField | Full street address |
| location | PointField | Geographic coordinates (SRID 4326) |
| opening_time | TimeField | Opening time |
| closing_time | TimeField | Closing time |
| working_days | CharField | Days of operation |
| phone | CharField | Contact phone (optional) |
| email | EmailField | Contact email (optional) |
| is_active | BooleanField | Whether office is operational |
