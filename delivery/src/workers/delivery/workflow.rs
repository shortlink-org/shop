//! Delivery Workflows
//!
//! Temporal workflows for order assignment and delivery.
//!
//! Workflows are deterministic and should only:
//! - Call activities for side effects
//! - Use Temporal primitives (signals, queries, timers)
//! - Manage workflow state
//!
//! The actual workflow implementations are registered in `runner.rs`.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Order assignment workflow state
#[derive(Debug, Clone)]
pub struct AssignOrderWorkflowState {
    /// Order ID being assigned
    pub order_id: Uuid,
    /// Pickup location
    pub pickup_location: Location,
    /// Delivery zone
    pub delivery_zone: String,
    /// Assigned courier ID (if assigned)
    pub assigned_courier_id: Option<Uuid>,
    /// Assignment attempts
    pub attempts: u32,
}

impl AssignOrderWorkflowState {
    /// Create initial workflow state
    pub fn new(order_id: Uuid, pickup_location: Location, delivery_zone: String) -> Self {
        Self {
            order_id,
            pickup_location,
            delivery_zone,
            assigned_courier_id: None,
            attempts: 0,
        }
    }
}

/// Delivery workflow state
#[derive(Debug, Clone)]
pub struct DeliverOrderWorkflowState {
    /// Order ID
    pub order_id: Uuid,
    /// Courier ID
    pub courier_id: Uuid,
    /// Current status
    pub status: DeliveryStatus,
}

impl DeliverOrderWorkflowState {
    /// Create initial workflow state
    pub fn new(order_id: Uuid, courier_id: Uuid) -> Self {
        Self {
            order_id,
            courier_id,
            status: DeliveryStatus::HeadingToPickup,
        }
    }
}

/// Delivery status
///
/// Represents the current stage of the delivery process.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum DeliveryStatus {
    /// Courier heading to pickup location
    HeadingToPickup,
    /// At pickup location, collecting package
    Collecting,
    /// En route to delivery location
    InTransit,
    /// At delivery location
    Delivering,
    /// Delivery completed successfully
    Completed,
    /// Delivery failed
    Failed,
}

/// Signal types for delivery workflow
///
/// These signals can be sent to a running delivery workflow
/// to trigger state transitions.
#[derive(Debug, Clone)]
pub enum DeliverySignal {
    /// Courier arrived at pickup location
    ArrivedAtPickup,
    /// Package collected from pickup location
    PackageCollected,
    /// Courier arrived at delivery location
    ArrivedAtDelivery,
    /// Delivery completed (success or failure)
    DeliveryCompleted { success: bool },
    /// Cancel delivery with reason
    Cancel { reason: String },
}

/// Workflow type name for order assignment
pub const ASSIGN_ORDER_WORKFLOW: &str = "assign_order";

/// Workflow type name for order delivery
pub const DELIVER_ORDER_WORKFLOW: &str = "deliver_order";

/// Signal channel name for delivery events
pub const DELIVERY_SIGNAL_CHANNEL: &str = "delivery_signal";

/// Maximum number of assignment attempts before giving up
pub const MAX_ASSIGNMENT_ATTEMPTS: u32 = 3;

/// Delay between assignment retry attempts (in seconds)
pub const ASSIGNMENT_RETRY_DELAY_SECS: u64 = 30;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_delivery_status_serialization() {
        let status = DeliveryStatus::HeadingToPickup;
        let json = serde_json::to_string(&status).unwrap();
        assert!(json.contains("HeadingToPickup"));
    }

    #[test]
    fn test_delivery_signal_serialization() {
        let signal = DeliverySignal::DeliveryCompleted { success: true };
        let json = serde_json::to_string(&signal).unwrap();
        assert!(json.contains("DeliveryCompleted"));
        assert!(json.contains("true"));
    }

    #[test]
    fn test_workflow_state_creation() {
        let order_id = Uuid::new_v4();
        let courier_id = Uuid::new_v4();
        let state = DeliverOrderWorkflowState::new(order_id, courier_id);

        assert_eq!(state.order_id, order_id);
        assert_eq!(state.courier_id, courier_id);
        assert_eq!(state.status, DeliveryStatus::HeadingToPickup);
    }
}
