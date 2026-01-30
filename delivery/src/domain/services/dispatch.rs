//! Dispatch Domain Service
//!
//! Contains business logic for selecting the optimal courier for a package.
//! This is a domain service because it operates on multiple aggregates
//! (Package and Courier) and contains complex business rules.

use crate::domain::model::courier::{CourierCapacity, CourierStatus, TransportType};
use crate::domain::model::vo::location::Location;

/// Courier data needed for dispatch decision
#[derive(Debug, Clone)]
pub struct CourierForDispatch {
    pub id: String,
    pub status: CourierStatus,
    pub transport_type: TransportType,
    pub max_distance_km: f64,
    pub capacity: CourierCapacity,
    pub current_location: Option<Location>,
    pub rating: f64,
    pub work_zone: String,
}

/// Package data needed for dispatch decision
#[derive(Debug, Clone)]
pub struct PackageForDispatch {
    pub id: String,
    pub pickup_location: Location,
    pub delivery_zone: String,
    pub is_urgent: bool,
}

/// Result of dispatch selection
#[derive(Debug, Clone)]
pub struct DispatchResult {
    pub courier_id: String,
    pub distance_to_pickup_km: f64,
    pub estimated_arrival_minutes: f64,
}

/// Reasons why a courier was rejected
#[derive(Debug, Clone, PartialEq)]
pub enum RejectionReason {
    NotAvailable,
    AtFullCapacity,
    TooFarFromPickup,
    NoLocationData,
    WrongZone,
}

/// Dispatch service for selecting optimal courier
pub struct DispatchService;

impl DispatchService {
    /// Find the nearest available courier for a package
    ///
    /// Algorithm:
    /// 1. Filter by status (must be FREE)
    /// 2. Filter by capacity (must have available slots)
    /// 3. Filter by zone (must match delivery zone)
    /// 4. Filter by distance (must be within courier's max distance)
    /// 5. Calculate distance to pickup using Haversine
    /// 6. Sort by: distance (primary), rating (secondary)
    /// 7. Return the best match
    pub fn find_nearest_courier(
        couriers: &[CourierForDispatch],
        package: &PackageForDispatch,
    ) -> Option<DispatchResult> {
        let mut candidates: Vec<(usize, f64)> = couriers
            .iter()
            .enumerate()
            .filter_map(|(idx, courier)| {
                // Check availability
                if !courier.status.can_accept_assignment() {
                    return None;
                }

                // Check capacity
                if !courier.capacity.can_accept() {
                    return None;
                }

                // Check zone
                if courier.work_zone != package.delivery_zone {
                    return None;
                }

                // Check location
                let location = courier.current_location.as_ref()?;

                // Calculate distance
                let distance = location.distance_to(&package.pickup_location);

                // Check max distance
                if distance > courier.max_distance_km {
                    return None;
                }

                Some((idx, distance))
            })
            .collect();

        // Sort by distance (ascending), then by rating (descending)
        candidates.sort_by(|a, b| {
            let dist_cmp = a.1.partial_cmp(&b.1).unwrap_or(std::cmp::Ordering::Equal);
            if dist_cmp == std::cmp::Ordering::Equal {
                let rating_a = couriers[a.0].rating;
                let rating_b = couriers[b.0].rating;
                rating_b.partial_cmp(&rating_a).unwrap_or(std::cmp::Ordering::Equal)
            } else {
                dist_cmp
            }
        });

        // Return the best candidate
        candidates.first().map(|(idx, distance)| {
            let courier = &couriers[*idx];
            let estimated_minutes = Self::estimate_arrival_time(*distance, courier.transport_type);
            
            DispatchResult {
                courier_id: courier.id.clone(),
                distance_to_pickup_km: *distance,
                estimated_arrival_minutes: estimated_minutes,
            }
        })
    }

    /// Estimate arrival time based on distance and transport type
    fn estimate_arrival_time(distance_km: f64, transport: TransportType) -> f64 {
        let speed_kmh = transport.average_speed_kmh();
        (distance_km / speed_kmh) * 60.0 // Convert to minutes
    }

    /// Validate if a specific courier can be assigned to a package
    pub fn validate_assignment(
        courier: &CourierForDispatch,
        package: &PackageForDispatch,
    ) -> Result<(), RejectionReason> {
        // Check availability
        if !courier.status.can_accept_assignment() {
            return Err(RejectionReason::NotAvailable);
        }

        // Check capacity
        if !courier.capacity.can_accept() {
            return Err(RejectionReason::AtFullCapacity);
        }

        // Check zone
        if courier.work_zone != package.delivery_zone {
            return Err(RejectionReason::WrongZone);
        }

        // Check location and distance
        let location = courier
            .current_location
            .as_ref()
            .ok_or(RejectionReason::NoLocationData)?;

        let distance = location.distance_to(&package.pickup_location);
        if distance > courier.max_distance_km {
            return Err(RejectionReason::TooFarFromPickup);
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_courier(id: &str, lat: f64, lon: f64, distance: f64) -> CourierForDispatch {
        CourierForDispatch {
            id: id.to_string(),
            status: CourierStatus::Free,
            transport_type: TransportType::Car,
            max_distance_km: distance,
            capacity: CourierCapacity::new(5),
            current_location: Some(Location::new(lat, lon, 10.0).unwrap()),
            rating: 4.5,
            work_zone: "zone1".to_string(),
        }
    }

    fn create_test_package(lat: f64, lon: f64) -> PackageForDispatch {
        PackageForDispatch {
            id: "pkg-1".to_string(),
            pickup_location: Location::new(lat, lon, 10.0).unwrap(),
            delivery_zone: "zone1".to_string(),
            is_urgent: false,
        }
    }

    #[test]
    fn test_find_nearest_courier() {
        let couriers = vec![
            create_test_courier("c1", 55.7558, 37.6173, 50.0), // Moscow
            create_test_courier("c2", 55.7600, 37.6200, 50.0), // Slightly closer
        ];

        let package = create_test_package(55.7610, 37.6210); // Near c2

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_some());
        assert_eq!(result.unwrap().courier_id, "c2");
    }

    #[test]
    fn test_find_nearest_courier_no_available() {
        let mut courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        courier.status = CourierStatus::Busy;

        let couriers = vec![courier];
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_none());
    }

    #[test]
    fn test_find_nearest_courier_too_far() {
        let courier = create_test_courier("c1", 55.7558, 37.6173, 1.0); // Max 1km
        let couriers = vec![courier];

        // St. Petersburg - too far
        let package = create_test_package(59.9343, 30.3351);

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_none());
    }

    #[test]
    fn test_validate_assignment_success() {
        let courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(&courier, &package);
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_assignment_not_available() {
        let mut courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        courier.status = CourierStatus::Busy;
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(&courier, &package);
        assert_eq!(result, Err(RejectionReason::NotAvailable));
    }

    #[test]
    fn test_validate_assignment_wrong_zone() {
        let mut courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        courier.work_zone = "zone2".to_string();
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(&courier, &package);
        assert_eq!(result, Err(RejectionReason::WrongZone));
    }
}
