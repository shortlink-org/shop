//! Deliver Order Command
//!
//! Data structure representing the command to confirm delivery.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Reason for failed delivery
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NotDeliveredReason {
    /// Customer not available
    CustomerUnavailable,
    /// Wrong address
    WrongAddress,
    /// Package refused
    Refused,
    /// Access denied to building/area
    AccessDenied,
    /// Other reason
    Other(String),
}

/// Command to confirm a successful delivery.
#[derive(Debug, Clone)]
pub struct ConfirmDelivered {
    /// The package ID
    package_id: Uuid,
    /// The courier ID
    courier_id: Uuid,
    /// Location where delivery was confirmed
    confirmation_location: Location,
    /// Optional photo proof (URL or base64)
    photo_proof: Option<String>,
    /// Optional signature (base64 encoded)
    signature: Option<String>,
}

impl ConfirmDelivered {
    /// Create a successful delivery command.
    pub fn new(
        package_id: Uuid,
        courier_id: Uuid,
        confirmation_location: Location,
        photo_proof: Option<String>,
        signature: Option<String>,
    ) -> Self {
        Self {
            package_id,
            courier_id,
            confirmation_location,
            photo_proof,
            signature,
        }
    }

    pub fn package_id(&self) -> Uuid {
        self.package_id
    }

    pub fn courier_id(&self) -> Uuid {
        self.courier_id
    }

    pub fn confirmation_location(&self) -> Location {
        self.confirmation_location
    }

    pub fn photo_proof(&self) -> Option<&str> {
        self.photo_proof.as_deref()
    }

    pub fn signature(&self) -> Option<&str> {
        self.signature.as_deref()
    }
}

/// Command to confirm a failed delivery.
#[derive(Debug, Clone)]
pub struct ConfirmNotDelivered {
    /// The package ID
    package_id: Uuid,
    /// The courier ID
    courier_id: Uuid,
    /// Location where delivery was confirmed
    confirmation_location: Location,
    /// Reason for failed delivery
    reason: NotDeliveredReason,
}

impl ConfirmNotDelivered {
    /// Create a failed delivery command.
    pub fn new(
        package_id: Uuid,
        courier_id: Uuid,
        confirmation_location: Location,
        reason: NotDeliveredReason,
    ) -> Self {
        Self {
            package_id,
            courier_id,
            confirmation_location,
            reason,
        }
    }

    pub fn package_id(&self) -> Uuid {
        self.package_id
    }

    pub fn courier_id(&self) -> Uuid {
        self.courier_id
    }

    pub fn confirmation_location(&self) -> Location {
        self.confirmation_location
    }

    pub fn reason(&self) -> &NotDeliveredReason {
        &self.reason
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn confirm_delivered_carries_delivery_fields() {
        let cmd = ConfirmDelivered::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
            Some("photo".to_string()),
            Some("signature".to_string()),
        );

        assert_eq!(cmd.photo_proof(), Some("photo"));
        assert_eq!(cmd.signature(), Some("signature"));
    }

    #[test]
    fn confirm_not_delivered_always_has_reason() {
        let cmd = ConfirmNotDelivered::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
            NotDeliveredReason::CustomerUnavailable,
        );

        assert!(matches!(
            cmd.reason(),
            NotDeliveredReason::CustomerUnavailable
        ));
    }
}
