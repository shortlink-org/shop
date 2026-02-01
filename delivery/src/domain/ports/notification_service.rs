//! Notification Service Port
//!
//! Defines the interface for sending push notifications to couriers.
//! This port is implemented by infrastructure adapters (e.g., FCM, APNs).

use async_trait::async_trait;
#[cfg(test)]
use mockall::automock;
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Order assigned notification payload
#[derive(Debug, Clone)]
pub struct OrderAssignedNotification {
    /// Package ID
    pub package_id: Uuid,
    /// Pickup address (street)
    pub pickup_address: String,
    /// Pickup location coordinates
    pub pickup_location: Location,
    /// Delivery address (street)
    pub delivery_address: String,
    /// Delivery location coordinates
    pub delivery_location: Location,
    /// Customer phone number
    pub customer_phone: String,
    /// Delivery window start (ISO 8601)
    pub delivery_start: String,
    /// Delivery window end (ISO 8601)
    pub delivery_end: String,
}

/// Delivery status notification payload
#[derive(Debug, Clone)]
pub struct DeliveryStatusNotification {
    /// Package ID
    pub package_id: Uuid,
    /// Status message
    pub message: String,
}

/// Errors that can occur during notification sending
#[derive(Debug, Error)]
pub enum NotificationError {
    /// Invalid push token
    #[error("Invalid push token: {0}")]
    InvalidToken(String),

    /// Connection error to notification service
    #[error("Connection error: {0}")]
    ConnectionError(String),

    /// Failed to send notification
    #[error("Failed to send notification: {0}")]
    SendError(String),

    /// Token expired or unregistered
    #[error("Token expired or unregistered: {0}")]
    TokenExpired(String),

    /// Rate limited
    #[error("Rate limited: {0}")]
    RateLimited(String),
}

/// Notification Service Port
///
/// Defines the contract for sending push notifications.
/// Implementations handle the actual push notification provider (FCM, APNs, etc.).
#[cfg_attr(test, automock)]
#[async_trait]
pub trait NotificationService: Send + Sync {
    /// Send order assigned notification to courier
    ///
    /// Notifies the courier that a new order has been assigned to them.
    async fn send_order_assigned(
        &self,
        push_token: &str,
        notification: OrderAssignedNotification,
    ) -> Result<(), NotificationError>;

    /// Send delivery status notification to courier
    ///
    /// Generic notification for status updates.
    async fn send_delivery_status(
        &self,
        push_token: &str,
        notification: DeliveryStatusNotification,
    ) -> Result<(), NotificationError>;
}
