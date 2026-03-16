//! Assignment Validation Domain Service
//!
//! Contains business rules for validating package assignments.
//! Validates that all business constraints are met before assignment.

use chrono::{DateTime, Timelike, Utc};

use crate::domain::model::courier::{Courier, CourierStatus};
use crate::domain::model::package::{Package, PackageStatus};

/// Validation errors for assignment
#[derive(Debug, Clone, PartialEq)]
pub enum AssignmentValidationError {
    /// Package is not in a valid status for assignment
    InvalidPackageStatus {
        current: PackageStatus,
        expected: PackageStatus,
    },
    /// Courier is not available
    CourierNotAvailable { current: CourierStatus },
    /// Courier is outside working hours
    OutsideWorkingHours {
        current_hour: u8,
        start_hour: u8,
        end_hour: u8,
    },
    /// Courier has reached maximum capacity
    CourierAtCapacity { current: u32, max: u32 },
    /// Distance exceeds courier's maximum
    DistanceExceedsMax { distance_km: f64, max_km: f64 },
}

impl std::fmt::Display for AssignmentValidationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            AssignmentValidationError::InvalidPackageStatus { current, expected } => {
                write!(
                    f,
                    "Invalid package status: {} (expected {})",
                    current, expected
                )
            }
            AssignmentValidationError::CourierNotAvailable { current } => {
                write!(f, "Courier not available: current status is {}", current)
            }
            AssignmentValidationError::OutsideWorkingHours {
                current_hour,
                start_hour,
                end_hour,
            } => {
                write!(
                    f,
                    "Outside working hours: current hour {} is not between {} and {}",
                    current_hour, start_hour, end_hour
                )
            }
            AssignmentValidationError::CourierAtCapacity { current, max } => {
                write!(f, "Courier at capacity: {}/{} packages", current, max)
            }
            AssignmentValidationError::DistanceExceedsMax {
                distance_km,
                max_km,
            } => {
                write!(
                    f,
                    "Distance {:.2} km exceeds courier's maximum {:.2} km",
                    distance_km, max_km
                )
            }
        }
    }
}

impl std::error::Error for AssignmentValidationError {}

/// Small domain context for validating manual assignments.
#[derive(Debug, Clone)]
pub struct AssignmentContext<'a> {
    pub courier: &'a Courier,
    pub package: &'a Package,
    pub current_time: DateTime<Utc>,
    pub distance_to_pickup_km: Option<f64>,
}

/// Assignment validation service
pub struct AssignmentValidationService;

impl AssignmentValidationService {
    /// Validate all business rules for package assignment.
    ///
    /// Returns `Ok(())` if assignment is valid, or a list of all validation errors.
    pub fn validate(context: AssignmentContext<'_>) -> Result<(), Vec<AssignmentValidationError>> {
        let mut errors = Vec::new();

        if context.package.status() != PackageStatus::InPool {
            errors.push(AssignmentValidationError::InvalidPackageStatus {
                current: context.package.status(),
                expected: PackageStatus::InPool,
            });
        }

        if context.courier.status() != CourierStatus::Free {
            errors.push(AssignmentValidationError::CourierNotAvailable {
                current: context.courier.status(),
            });
        }

        let current_hour = context.current_time.time().hour() as u8;
        let work_hours = context.courier.work_hours();
        if !Self::is_within_working_hours(
            current_hour,
            work_hours.start.hour() as u8,
            work_hours.end.hour() as u8,
        ) {
            errors.push(AssignmentValidationError::OutsideWorkingHours {
                current_hour,
                start_hour: work_hours.start.hour() as u8,
                end_hour: work_hours.end.hour() as u8,
            });
        }

        if context.courier.current_load() >= context.courier.max_load() {
            errors.push(AssignmentValidationError::CourierAtCapacity {
                current: context.courier.current_load(),
                max: context.courier.max_load(),
            });
        }

        if let Some(distance_to_pickup_km) = context.distance_to_pickup_km {
            if distance_to_pickup_km > context.courier.max_distance_km() {
                errors.push(AssignmentValidationError::DistanceExceedsMax {
                    distance_km: distance_to_pickup_km,
                    max_km: context.courier.max_distance_km(),
                });
            }
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(errors)
        }
    }

    fn is_within_working_hours(current: u8, start: u8, end: u8) -> bool {
        if start <= end {
            current >= start && current < end
        } else {
            current >= start || current < end
        }
    }
}

#[cfg(test)]
mod tests {
    use chrono::NaiveTime;

    use super::*;
    use crate::domain::model::courier::WorkHours;
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::domain::model::vo::TransportType;

    fn create_valid_courier() -> Courier {
        let mut courier = Courier::builder(
            "Courier".to_string(),
            "+491234567890".to_string(),
            "courier@test.com".to_string(),
            TransportType::Car,
            20.0,
            "zone-1".to_string(),
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

    fn create_valid_package() -> Package {
        let now = Utc::now();
        let mut package = Package::new(
            uuid::Uuid::new_v4(),
            uuid::Uuid::new_v4(),
            None,
            None,
            None,
            None,
            Address::new(
                "Pickup".to_string(),
                "Berlin".to_string(),
                Location::new(52.52, 13.405, 10.0).unwrap(),
            ),
            Address::new(
                "Drop".to_string(),
                "Berlin".to_string(),
                Location::new(52.53, 13.41, 10.0).unwrap(),
            ),
            DeliveryPeriod::new(
                now + chrono::Duration::hours(1),
                now + chrono::Duration::hours(2),
            )
            .unwrap(),
            1.0,
            Priority::Normal,
            "zone-1".to_string(),
        );
        package.move_to_pool().unwrap();
        package
    }

    fn midday() -> DateTime<Utc> {
        Utc::now()
            .date_naive()
            .and_hms_opt(12, 0, 0)
            .unwrap()
            .and_utc()
    }

    #[test]
    fn test_valid_assignment() {
        let courier = create_valid_courier();
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(AssignmentContext {
            courier: &courier,
            package: &package,
            current_time: midday(),
            distance_to_pickup_km: Some(5.0),
        });

        assert!(result.is_ok());
    }

    #[test]
    fn test_invalid_package_status() {
        let courier = create_valid_courier();
        let package = Package::new(
            uuid::Uuid::new_v4(),
            uuid::Uuid::new_v4(),
            None,
            None,
            None,
            None,
            Address::new(
                "Pickup".to_string(),
                "Berlin".to_string(),
                Location::new(52.52, 13.405, 10.0).unwrap(),
            ),
            Address::new(
                "Drop".to_string(),
                "Berlin".to_string(),
                Location::new(52.53, 13.41, 10.0).unwrap(),
            ),
            DeliveryPeriod::new(
                Utc::now() + chrono::Duration::hours(1),
                Utc::now() + chrono::Duration::hours(2),
            )
            .unwrap(),
            1.0,
            Priority::Normal,
            "zone-1".to_string(),
        );

        let result = AssignmentValidationService::validate(AssignmentContext {
            courier: &courier,
            package: &package,
            current_time: midday(),
            distance_to_pickup_km: Some(5.0),
        });

        assert!(result.is_err());
    }

    #[test]
    fn test_courier_not_available() {
        let courier = Courier::builder(
            "Courier".to_string(),
            "+491234567890".to_string(),
            "courier@test.com".to_string(),
            TransportType::Car,
            20.0,
            "zone-1".to_string(),
            WorkHours::new(
                NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
                NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
                vec![1, 2, 3, 4, 5],
            )
            .unwrap(),
        )
        .build()
        .unwrap();
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(AssignmentContext {
            courier: &courier,
            package: &package,
            current_time: midday(),
            distance_to_pickup_km: Some(5.0),
        });

        assert!(result.is_err());
    }

    #[test]
    fn test_distance_exceeds_max() {
        let courier = create_valid_courier();
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(AssignmentContext {
            courier: &courier,
            package: &package,
            current_time: midday(),
            distance_to_pickup_km: Some(25.0),
        });

        assert!(result.is_err());
    }
}
