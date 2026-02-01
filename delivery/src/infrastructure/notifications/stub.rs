//! Stub Notification Service
//!
//! Placeholder implementation that logs notifications instead of sending them.
//! TODO: Replace with real FCM/APNs implementation.

use async_trait::async_trait;
use tracing::info;

use crate::domain::ports::{
    DeliveryStatusNotification, NotificationError, NotificationService, OrderAssignedNotification,
};

/// Stub notification service for development/testing
///
/// This implementation logs notifications instead of sending them.
/// Replace with real FCM/APNs implementation when ready.
pub struct StubNotificationService;

impl StubNotificationService {
    /// Create a new stub notification service
    pub fn new() -> Self {
        Self
    }
}

impl Default for StubNotificationService {
    fn default() -> Self {
        Self::new()
    }
}

#[async_trait]
impl NotificationService for StubNotificationService {
    async fn send_order_assigned(
        &self,
        push_token: &str,
        notification: OrderAssignedNotification,
    ) -> Result<(), NotificationError> {
        info!(
            push_token = push_token,
            package_id = %notification.package_id,
            pickup_address = %notification.pickup_address,
            delivery_address = %notification.delivery_address,
            delivery_start = %notification.delivery_start,
            delivery_end = %notification.delivery_end,
            "TODO: Send push notification - Order Assigned"
        );

        // TODO: Implement real FCM/APNs notification
        // Example FCM payload:
        // {
        //     "message": {
        //         "token": push_token,
        //         "notification": {
        //             "title": "New order assigned",
        //             "body": "Order #{package_id} is ready for pickup"
        //         },
        //         "data": {
        //             "package_id": notification.package_id,
        //             "pickup_address": notification.pickup_address,
        //             "pickup_lat": notification.pickup_location.latitude(),
        //             "pickup_lon": notification.pickup_location.longitude(),
        //             "delivery_address": notification.delivery_address,
        //             "delivery_lat": notification.delivery_location.latitude(),
        //             "delivery_lon": notification.delivery_location.longitude(),
        //             "customer_phone": notification.customer_phone,
        //             "delivery_start": notification.delivery_start,
        //             "delivery_end": notification.delivery_end
        //         }
        //     }
        // }

        Ok(())
    }

    async fn send_delivery_status(
        &self,
        push_token: &str,
        notification: DeliveryStatusNotification,
    ) -> Result<(), NotificationError> {
        info!(
            push_token = push_token,
            package_id = %notification.package_id,
            message = %notification.message,
            "TODO: Send push notification - Delivery Status"
        );

        // TODO: Implement real FCM/APNs notification

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::vo::location::Location;
    use uuid::Uuid;

    #[tokio::test]
    async fn test_send_order_assigned() {
        let service = StubNotificationService::new();
        
        let notification = OrderAssignedNotification {
            package_id: Uuid::new_v4(),
            pickup_address: "123 Pickup St".to_string(),
            pickup_location: Location::new(52.52, 13.405, 10.0).unwrap(),
            delivery_address: "456 Delivery Ave".to_string(),
            delivery_location: Location::new(52.53, 13.41, 10.0).unwrap(),
            customer_phone: "+1234567890".to_string(),
            delivery_start: "2024-01-15T10:00:00Z".to_string(),
            delivery_end: "2024-01-15T12:00:00Z".to_string(),
        };

        let result = service.send_order_assigned("test_token", notification).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_send_delivery_status() {
        let service = StubNotificationService::new();
        
        let notification = DeliveryStatusNotification {
            package_id: Uuid::new_v4(),
            message: "Package delivered successfully".to_string(),
        };

        let result = service.send_delivery_status("test_token", notification).await;
        assert!(result.is_ok());
    }
}
