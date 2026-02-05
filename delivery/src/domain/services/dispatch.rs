//! Dispatch Domain Service
//!
//! Contains business logic for selecting the optimal courier for a package.
//! This is a domain service because it operates on multiple aggregates
//! (Package and Courier) and contains complex business rules.

use crate::domain::model::courier::{CourierCapacity, CourierStatus};
use crate::domain::model::vo::location::Location;
use crate::domain::model::vo::TransportType;

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

/// Reason why a single courier was rejected (for logging/metrics/debugging).
#[derive(Debug, Clone, PartialEq)]
pub struct CandidateRejection {
    pub courier_id: String,
    pub reason: RejectionReason,
}

/// Aggregate of all rejections when no courier could be selected.
#[derive(Debug, Clone)]
pub struct DispatchFailure {
    pub rejections: Vec<CandidateRejection>,
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
    /// Find the nearest available courier for a package.
    ///
    /// Returns `Ok(DispatchResult)` with the best candidate, or `Err(DispatchFailure)` with
    /// per-courier rejection reasons (useful for logs, metrics, debugging).
    ///
    /// Algorithm:
    /// 1. For each courier: validate via `validate_assignment` (gets distance once).
    /// 2. Sort by distance (asc), then by rating (desc).
    /// 3. Return best candidate or all rejections.
    pub fn find_nearest_courier(
        couriers: &[CourierForDispatch],
        package: &PackageForDispatch,
    ) -> Result<DispatchResult, DispatchFailure> {
        let mut rejections = Vec::with_capacity(couriers.len());
        let mut candidates: Vec<(usize, f64)> = Vec::new();

        for (idx, courier) in couriers.iter().enumerate() {
            match Self::validate_assignment(courier, package) {
                Ok(distance) => candidates.push((idx, distance)),
                Err(reason) => rejections.push(CandidateRejection {
                    courier_id: courier.id.clone(),
                    reason,
                }),
            }
        }

        if candidates.is_empty() {
            return Err(DispatchFailure { rejections });
        }

        // Sort by distance (asc), then by rating (desc). Use total_cmp for stable ordering (handles f64 without NaN).
        candidates.sort_by(|a, b| {
            let dist_cmp = a.1.total_cmp(&b.1);
            if dist_cmp == std::cmp::Ordering::Equal {
                let rating_a = couriers[a.0].rating;
                let rating_b = couriers[b.0].rating;
                rating_b.total_cmp(&rating_a)
            } else {
                dist_cmp
            }
        });

        let (idx, distance) = candidates[0];
        let courier = &couriers[idx];
        let estimated_minutes = courier.transport_type.calculate_travel_time_minutes(distance);

        Ok(DispatchResult {
            courier_id: courier.id.clone(),
            distance_to_pickup_km: distance,
            estimated_arrival_minutes: estimated_minutes,
        })
    }

    /// Validate if a specific courier can be assigned to a package.
    ///
    /// Returns the distance to pickup in km on success, so callers do not need to compute Haversine twice.
    pub fn validate_assignment(
        courier: &CourierForDispatch,
        package: &PackageForDispatch,
    ) -> Result<f64, RejectionReason> {
        if !courier.status.can_accept_assignment() {
            return Err(RejectionReason::NotAvailable);
        }

        if !courier.capacity.can_accept() {
            return Err(RejectionReason::AtFullCapacity);
        }

        if courier.work_zone != package.delivery_zone {
            return Err(RejectionReason::WrongZone);
        }

        let location = courier
            .current_location
            .as_ref()
            .ok_or(RejectionReason::NoLocationData)?;

        let distance = location.distance_to(&package.pickup_location);
        if distance > courier.max_distance_km {
            return Err(RejectionReason::TooFarFromPickup);
        }

        Ok(distance)
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
        assert!(result.is_ok());
        assert_eq!(result.unwrap().courier_id, "c2");
    }

    #[test]
    fn test_find_nearest_courier_equal_distance_chooses_by_rating() {
        let mut c1 = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        c1.rating = 3.0;
        let mut c2 = create_test_courier("c2", 55.7558, 37.6173, 50.0);
        c2.rating = 5.0;
        let couriers = vec![c1, c2];
        let package = create_test_package(55.7558, 37.6173); // Same point → same distance

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_ok());
        assert_eq!(result.unwrap().courier_id, "c2");
    }

    #[test]
    fn test_find_nearest_courier_empty_list_returns_failure() {
        let couriers: Vec<CourierForDispatch> = vec![];
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_err());
        let failure = result.unwrap_err();
        assert!(failure.rejections.is_empty());
    }

    #[test]
    fn test_find_nearest_courier_no_available() {
        let mut courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        courier.status = CourierStatus::Busy;

        let couriers = vec![courier];
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_err());
        let failure = result.unwrap_err();
        assert_eq!(failure.rejections.len(), 1);
        assert_eq!(failure.rejections[0].courier_id, "c1");
        assert_eq!(failure.rejections[0].reason, RejectionReason::NotAvailable);
    }

    #[test]
    fn test_find_nearest_courier_too_far() {
        let courier = create_test_courier("c1", 55.7558, 37.6173, 1.0); // Max 1km
        let couriers = vec![courier];

        let package = create_test_package(59.9343, 30.3351); // St. Petersburg - too far

        let result = DispatchService::find_nearest_courier(&couriers, &package);
        assert!(result.is_err());
        let failure = result.unwrap_err();
        assert_eq!(failure.rejections.len(), 1);
        assert_eq!(failure.rejections[0].reason, RejectionReason::TooFarFromPickup);
    }

    #[test]
    fn test_validate_assignment_success() {
        let courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(&courier, &package);
        assert!(result.is_ok());
        let distance = result.unwrap();
        assert!(distance > 0.0 && distance < 1.0);
    }

    #[test]
    fn test_validate_assignment_no_location_data() {
        let mut courier = create_test_courier("c1", 55.7558, 37.6173, 50.0);
        courier.current_location = None;
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(&courier, &package);
        assert_eq!(result, Err(RejectionReason::NoLocationData));
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
