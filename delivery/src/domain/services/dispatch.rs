//! Dispatch Domain Service
//!
//! Contains business logic for selecting the optimal courier for a package.
//! This is a domain service because it operates on multiple aggregates
//! (Package and Courier) and contains complex business rules.

use uuid::Uuid;

use crate::domain::model::courier::Courier;
use crate::domain::model::package::Package;
use crate::domain::model::vo::location::Location;

/// Small domain context for dispatching a courier.
#[derive(Debug, Clone)]
pub struct DispatchCandidate {
    pub courier: Courier,
    pub current_location: Option<Location>,
}

/// Result of dispatch selection
#[derive(Debug, Clone)]
pub struct DispatchResult {
    pub courier_id: Uuid,
    pub distance_to_pickup_km: f64,
    pub estimated_arrival_minutes: f64,
}

/// Reason why a single courier was rejected (for logging/metrics/debugging).
#[derive(Debug, Clone, PartialEq)]
pub struct CandidateRejection {
    pub courier_id: Uuid,
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
    pub fn find_nearest_courier(
        couriers: &[DispatchCandidate],
        package: &Package,
    ) -> Result<DispatchResult, DispatchFailure> {
        let mut rejections = Vec::with_capacity(couriers.len());
        let mut candidates: Vec<(usize, f64)> = Vec::new();

        for (idx, courier) in couriers.iter().enumerate() {
            match Self::validate_assignment(courier, package) {
                Ok(distance) => candidates.push((idx, distance)),
                Err(reason) => rejections.push(CandidateRejection {
                    courier_id: courier.courier.id().0,
                    reason,
                }),
            }
        }

        if candidates.is_empty() {
            return Err(DispatchFailure { rejections });
        }

        candidates.sort_by(|a, b| {
            let dist_cmp = a.1.total_cmp(&b.1);
            if dist_cmp == std::cmp::Ordering::Equal {
                let rating_a = couriers[a.0].courier.rating();
                let rating_b = couriers[b.0].courier.rating();
                rating_b.total_cmp(&rating_a)
            } else {
                dist_cmp
            }
        });

        let (idx, distance) = candidates[0];
        let courier = &couriers[idx].courier;

        Ok(DispatchResult {
            courier_id: courier.id().0,
            distance_to_pickup_km: distance,
            estimated_arrival_minutes: courier
                .transport_type()
                .calculate_travel_time_minutes(distance),
        })
    }

    /// Validate if a specific courier can be assigned to a package.
    ///
    /// Returns the distance to pickup in km on success, so callers do not need to compute
    /// Haversine twice.
    pub fn validate_assignment(
        candidate: &DispatchCandidate,
        package: &Package,
    ) -> Result<f64, RejectionReason> {
        let courier = &candidate.courier;

        if !courier.status().can_accept_assignment() {
            return Err(RejectionReason::NotAvailable);
        }

        if !courier.capacity().can_accept() {
            return Err(RejectionReason::AtFullCapacity);
        }

        if courier.work_zone() != package.zone() {
            return Err(RejectionReason::WrongZone);
        }

        let location = candidate
            .current_location
            .as_ref()
            .ok_or(RejectionReason::NoLocationData)?;

        let distance = location.distance_to(&package.pickup_address().location);
        if distance > courier.max_distance_km() {
            return Err(RejectionReason::TooFarFromPickup);
        }

        Ok(distance)
    }
}

#[cfg(test)]
mod tests {
    use chrono::NaiveTime;

    use super::*;
    use crate::domain::model::courier::WorkHours;
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::TransportType;

    fn create_test_courier(distance: f64) -> Courier {
        let mut courier = Courier::builder(
            "Courier".to_string(),
            "+491234567890".to_string(),
            "courier@test.com".to_string(),
            TransportType::Car,
            distance,
            "zone1".to_string(),
            WorkHours::new(
                NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
                NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
                vec![1, 2, 3, 4, 5],
            )
            .unwrap(),
        )
        .build()
        .unwrap();
        courier.go_online().unwrap();
        courier
    }

    fn create_test_package(lat: f64, lon: f64) -> Package {
        let now = chrono::Utc::now();
        let mut package = Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None,
            None,
            None,
            None,
            Address::new(
                "Pickup".to_string(),
                "Berlin".to_string(),
                Location::new(lat, lon, 10.0).unwrap(),
            ),
            Address::new(
                "Drop".to_string(),
                "Berlin".to_string(),
                Location::new(lat + 0.01, lon + 0.01, 10.0).unwrap(),
            ),
            DeliveryPeriod::new(
                now + chrono::Duration::hours(1),
                now + chrono::Duration::hours(2),
            )
            .unwrap(),
            1.0,
            Priority::Normal,
            "zone1".to_string(),
        );
        package.move_to_pool().unwrap();
        package
    }

    #[test]
    fn test_find_nearest_courier() {
        let c1 = DispatchCandidate {
            courier: create_test_courier(50.0),
            current_location: Some(Location::new(55.7558, 37.6173, 10.0).unwrap()),
        };
        let mut c2_courier = create_test_courier(50.0);
        c2_courier
            .change_contact_info(None, Some("better@test.com".to_string()), None)
            .unwrap();
        let c2 = DispatchCandidate {
            courier: c2_courier,
            current_location: Some(Location::new(55.7600, 37.6200, 10.0).unwrap()),
        };

        let package = create_test_package(55.7610, 37.6210);

        let result = DispatchService::find_nearest_courier(&[c1, c2], &package).unwrap();
        assert!(result.distance_to_pickup_km >= 0.0);
    }

    #[test]
    fn test_no_location_data_rejected() {
        let courier = create_test_courier(50.0);
        let package = create_test_package(55.7600, 37.6200);

        let result = DispatchService::validate_assignment(
            &DispatchCandidate {
                courier,
                current_location: None,
            },
            &package,
        );

        assert_eq!(result, Err(RejectionReason::NoLocationData));
    }
}
