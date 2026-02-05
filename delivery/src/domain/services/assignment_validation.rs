//! Assignment Validation Domain Service
//!
//! Contains business rules for validating package assignments.
//! Validates that all business constraints are met before assignment.

use crate::domain::model::courier::CourierStatus;
use crate::domain::model::package::PackageStatus;

/// Validation errors for assignment
#[derive(Debug, Clone, PartialEq)]
pub enum AssignmentValidationError {
    /// Package is not in a valid status for assignment
    InvalidPackageStatus {
        current: PackageStatus,
        expected: PackageStatus,
    },
    /// Courier is not available
    CourierNotAvailable {
        current: CourierStatus,
    },
    /// Courier is outside working hours
    OutsideWorkingHours {
        current_hour: u8,
        start_hour: u8,
        end_hour: u8,
    },
    /// Courier has reached maximum capacity
    CourierAtCapacity {
        current: u32,
        max: u32,
    },
    /// Distance exceeds courier's maximum
    DistanceExceedsMax {
        distance_km: f64,
        max_km: f64,
    },
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
                write!(
                    f,
                    "Courier at capacity: {}/{} packages",
                    current, max
                )
            }
            AssignmentValidationError::DistanceExceedsMax { distance_km, max_km } => {
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

/// Courier availability data for validation
#[derive(Debug, Clone)]
pub struct CourierAvailability {
    pub status: CourierStatus,
    pub current_load: u32,
    pub max_load: u32,
    pub work_start_hour: u8,
    pub work_end_hour: u8,
    pub max_distance_km: f64,
}

/// Package data for validation
#[derive(Debug, Clone)]
pub struct PackageForValidation {
    pub status: PackageStatus,
    pub distance_to_courier_km: f64,
}

/// Assignment validation service
pub struct AssignmentValidationService;

impl AssignmentValidationService {
    /// Validate all business rules for package assignment
    ///
    /// Returns Ok(()) if assignment is valid, or a list of all validation errors.
    pub fn validate(
        courier: &CourierAvailability,
        package: &PackageForValidation,
        current_hour: u8,
    ) -> Result<(), Vec<AssignmentValidationError>> {
        let mut errors = Vec::new();

        // Validate package status
        if package.status != PackageStatus::InPool {
            errors.push(AssignmentValidationError::InvalidPackageStatus {
                current: package.status,
                expected: PackageStatus::InPool,
            });
        }

        // Validate courier status
        if courier.status != CourierStatus::Free {
            errors.push(AssignmentValidationError::CourierNotAvailable {
                current: courier.status,
            });
        }

        // Validate working hours
        if !Self::is_within_working_hours(current_hour, courier.work_start_hour, courier.work_end_hour) {
            errors.push(AssignmentValidationError::OutsideWorkingHours {
                current_hour,
                start_hour: courier.work_start_hour,
                end_hour: courier.work_end_hour,
            });
        }

        // Validate capacity
        if courier.current_load >= courier.max_load {
            errors.push(AssignmentValidationError::CourierAtCapacity {
                current: courier.current_load,
                max: courier.max_load,
            });
        }

        // Validate distance
        if package.distance_to_courier_km > courier.max_distance_km {
            errors.push(AssignmentValidationError::DistanceExceedsMax {
                distance_km: package.distance_to_courier_km,
                max_km: courier.max_distance_km,
            });
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(errors)
        }
    }

    /// Check if current hour is within working hours
    fn is_within_working_hours(current: u8, start: u8, end: u8) -> bool {
        if start <= end {
            // Normal case: e.g., 9-18
            current >= start && current < end
        } else {
            // Overnight case: e.g., 22-6
            current >= start || current < end
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_valid_courier() -> CourierAvailability {
        CourierAvailability {
            status: CourierStatus::Free,
            current_load: 2,
            max_load: 5,
            work_start_hour: 9,
            work_end_hour: 18,
            max_distance_km: 20.0,
        }
    }

    fn create_valid_package() -> PackageForValidation {
        PackageForValidation {
            status: PackageStatus::InPool,
            distance_to_courier_km: 5.0,
        }
    }

    #[test]
    fn test_valid_assignment() {
        let courier = create_valid_courier();
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_ok());
    }

    #[test]
    fn test_invalid_package_status() {
        let courier = create_valid_courier();
        let mut package = create_valid_package();
        package.status = PackageStatus::Assigned;

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| matches!(e, AssignmentValidationError::InvalidPackageStatus { .. })));
    }

    #[test]
    fn test_courier_not_available() {
        let mut courier = create_valid_courier();
        courier.status = CourierStatus::Busy;
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| matches!(e, AssignmentValidationError::CourierNotAvailable { .. })));
    }

    #[test]
    fn test_outside_working_hours() {
        let courier = create_valid_courier();
        let package = create_valid_package();

        // Before working hours
        let result = AssignmentValidationService::validate(&courier, &package, 7);
        assert!(result.is_err());

        // After working hours
        let result = AssignmentValidationService::validate(&courier, &package, 20);
        assert!(result.is_err());
    }

    #[test]
    fn test_courier_at_capacity() {
        let mut courier = create_valid_courier();
        courier.current_load = 5; // At max
        let package = create_valid_package();

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| matches!(e, AssignmentValidationError::CourierAtCapacity { .. })));
    }

    #[test]
    fn test_distance_exceeds_max() {
        let courier = create_valid_courier();
        let mut package = create_valid_package();
        package.distance_to_courier_km = 30.0; // Exceeds 20km max

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.iter().any(|e| matches!(e, AssignmentValidationError::DistanceExceedsMax { .. })));
    }

    #[test]
    fn test_multiple_errors() {
        let mut courier = create_valid_courier();
        courier.status = CourierStatus::Busy;
        courier.current_load = 5;

        let mut package = create_valid_package();
        package.status = PackageStatus::Delivered;

        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors.len() >= 3); // At least 3 errors
    }

    #[test]
    fn test_overnight_working_hours() {
        let mut courier = create_valid_courier();
        courier.work_start_hour = 22;
        courier.work_end_hour = 6;
        let package = create_valid_package();

        // Should be valid at 23:00
        let result = AssignmentValidationService::validate(&courier, &package, 23);
        assert!(result.is_ok());

        // Should be valid at 2:00
        let result = AssignmentValidationService::validate(&courier, &package, 2);
        assert!(result.is_ok());

        // Should be invalid at 12:00
        let result = AssignmentValidationService::validate(&courier, &package, 12);
        assert!(result.is_err());
    }

    #[test]
    fn test_working_hours_boundary_inclusive_start_exclusive_end() {
        let courier = create_valid_courier(); // 9-18
        let package = create_valid_package();

        // At start hour (9) is valid
        let result = AssignmentValidationService::validate(&courier, &package, 9);
        assert!(result.is_ok());

        // At 17 (before end 18) is valid
        let result = AssignmentValidationService::validate(&courier, &package, 17);
        assert!(result.is_ok());

        // At end hour (18) is invalid (exclusive end)
        let result = AssignmentValidationService::validate(&courier, &package, 18);
        assert!(result.is_err());
    }
}
