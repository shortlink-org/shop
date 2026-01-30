use std::fmt;

/// Package status representing the delivery lifecycle
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PackageStatus {
    /// Package accepted into the pool
    Accepted,
    /// Package in pool, waiting for assignment
    InPool,
    /// Package assigned to a courier
    Assigned,
    /// Package in transit
    InTransit,
    /// Package delivered successfully
    Delivered,
    /// Package not delivered
    NotDelivered,
    /// Package requires manual handling
    RequiresHandling,
}

impl fmt::Display for PackageStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PackageStatus::Accepted => write!(f, "Accepted"),
            PackageStatus::InPool => write!(f, "InPool"),
            PackageStatus::Assigned => write!(f, "Assigned"),
            PackageStatus::InTransit => write!(f, "InTransit"),
            PackageStatus::Delivered => write!(f, "Delivered"),
            PackageStatus::NotDelivered => write!(f, "NotDelivered"),
            PackageStatus::RequiresHandling => write!(f, "RequiresHandling"),
        }
    }
}

/// Error for invalid state transitions
#[derive(Debug, Clone, PartialEq)]
pub struct InvalidTransitionError {
    pub from: PackageStatus,
    pub to: PackageStatus,
}

impl fmt::Display for InvalidTransitionError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "Invalid state transition from {} to {}",
            self.from, self.to
        )
    }
}

impl std::error::Error for InvalidTransitionError {}

impl PackageStatus {
    /// Check if transition to another status is valid
    ///
    /// Valid transitions:
    /// ```text
    /// Accepted -> InPool
    /// InPool -> Assigned
    /// Assigned -> InTransit
    /// InTransit -> Delivered | NotDelivered
    /// NotDelivered -> RequiresHandling
    /// RequiresHandling -> InPool (return to pool)
    /// ```
    pub fn can_transition_to(&self, target: PackageStatus) -> bool {
        matches!(
            (self, target),
            (PackageStatus::Accepted, PackageStatus::InPool)
                | (PackageStatus::InPool, PackageStatus::Assigned)
                | (PackageStatus::Assigned, PackageStatus::InTransit)
                | (PackageStatus::InTransit, PackageStatus::Delivered)
                | (PackageStatus::InTransit, PackageStatus::NotDelivered)
                | (PackageStatus::NotDelivered, PackageStatus::RequiresHandling)
                | (PackageStatus::RequiresHandling, PackageStatus::InPool)
        )
    }

    /// Transition to a new status, returning error if invalid
    pub fn transition_to(self, target: PackageStatus) -> Result<PackageStatus, InvalidTransitionError> {
        if self.can_transition_to(target) {
            Ok(target)
        } else {
            Err(InvalidTransitionError {
                from: self,
                to: target,
            })
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_transitions() {
        assert!(PackageStatus::Accepted.can_transition_to(PackageStatus::InPool));
        assert!(PackageStatus::InPool.can_transition_to(PackageStatus::Assigned));
        assert!(PackageStatus::Assigned.can_transition_to(PackageStatus::InTransit));
        assert!(PackageStatus::InTransit.can_transition_to(PackageStatus::Delivered));
        assert!(PackageStatus::InTransit.can_transition_to(PackageStatus::NotDelivered));
        assert!(PackageStatus::NotDelivered.can_transition_to(PackageStatus::RequiresHandling));
        assert!(PackageStatus::RequiresHandling.can_transition_to(PackageStatus::InPool));
    }

    #[test]
    fn test_invalid_transitions() {
        assert!(!PackageStatus::Accepted.can_transition_to(PackageStatus::Delivered));
        assert!(!PackageStatus::InPool.can_transition_to(PackageStatus::Delivered));
        assert!(!PackageStatus::Delivered.can_transition_to(PackageStatus::InPool));
    }

    #[test]
    fn test_transition_to_success() {
        let status = PackageStatus::Accepted;
        let new_status = status.transition_to(PackageStatus::InPool);
        assert_eq!(new_status, Ok(PackageStatus::InPool));
    }

    #[test]
    fn test_transition_to_error() {
        let status = PackageStatus::Accepted;
        let result = status.transition_to(PackageStatus::Delivered);
        assert!(result.is_err());
    }
}
