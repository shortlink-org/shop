//! Deliver Order Command
//!
//! Data structure representing the command to confirm delivery.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Result of the delivery attempt
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DeliveryResult {
    /// Package delivered successfully
    Delivered,
    /// Package not delivered
    NotDelivered,
}

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

/// Command to confirm delivery
#[derive(Debug, Clone)]
pub struct Command {
    /// The package ID
    pub package_id: Uuid,
    /// The courier ID
    pub courier_id: Uuid,
    /// Delivery result
    pub result: DeliveryResult,
    /// Reason for failed delivery (required if result is NotDelivered)
    pub not_delivered_reason: Option<NotDeliveredReason>,
    /// Location where delivery was confirmed
    pub confirmation_location: Location,
    /// Optional photo proof (URL or base64)
    pub photo_proof: Option<String>,
    /// Optional signature (base64 encoded)
    pub signature: Option<String>,
    /// Optional notes from courier
    pub notes: Option<String>,
}

impl Command {
    /// Create a successful delivery command
    pub fn delivered(
        package_id: Uuid,
        courier_id: Uuid,
        confirmation_location: Location,
        photo_proof: Option<String>,
        signature: Option<String>,
    ) -> Self {
        Self {
            package_id,
            courier_id,
            result: DeliveryResult::Delivered,
            not_delivered_reason: None,
            confirmation_location,
            photo_proof,
            signature,
            notes: None,
        }
    }

    /// Create a failed delivery command
    pub fn not_delivered(
        package_id: Uuid,
        courier_id: Uuid,
        confirmation_location: Location,
        reason: NotDeliveredReason,
        notes: Option<String>,
    ) -> Self {
        Self {
            package_id,
            courier_id,
            result: DeliveryResult::NotDelivered,
            not_delivered_reason: Some(reason),
            confirmation_location,
            photo_proof: None,
            signature: None,
            notes,
        }
    }
}
