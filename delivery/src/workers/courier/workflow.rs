//! Courier Workflows
//!
//! Temporal workflows for courier lifecycle management.
//!
//! Workflows are deterministic and should only:
//! - Call activities for side effects
//! - Use Temporal primitives (signals, queries, timers)
//! - Manage workflow state
//!
//! The actual workflow implementation is registered in `runner.rs`.

use uuid::Uuid;

use crate::domain::model::courier::CourierStatus;

/// Courier lifecycle workflow state
#[derive(Debug, Clone)]
pub struct CourierWorkflowState {
    /// Courier ID
    pub courier_id: Uuid,
    /// Current status
    pub status: CourierStatus,
    /// Whether the workflow is running
    pub is_running: bool,
}

impl CourierWorkflowState {
    /// Create initial workflow state
    pub fn new(courier_id: Uuid) -> Self {
        Self {
            courier_id,
            status: CourierStatus::Unavailable,
            is_running: true,
        }
    }
}

/// Signal types for courier workflow
///
/// These signals can be sent to a running courier lifecycle workflow
/// to trigger state transitions and activity executions.
#[derive(Debug, Clone)]
pub enum CourierSignal {
    /// Courier goes online - triggers status update to Free
    GoOnline,
    /// Courier goes offline - triggers status update to Unavailable
    GoOffline,
    /// Package assigned to courier - triggers accept_package activity
    PackageAssigned { package_id: Uuid },
    /// Delivery completed - triggers complete_delivery activity
    DeliveryCompleted { package_id: Uuid, success: bool },
    /// Stop the workflow gracefully
    Stop,
}

/// Query types for courier workflow
///
/// These queries can be used to inspect the current state of
/// a running courier workflow without affecting its execution.
#[derive(Debug, Clone)]
pub enum CourierQuery {
    /// Get current courier status
    GetStatus,
    /// Get current package load
    GetLoad,
    /// Check if courier is available for new assignments
    IsAvailable,
}

/// Workflow type name for registration
pub const COURIER_LIFECYCLE_WORKFLOW: &str = "courier_lifecycle";

/// Signal channel name for courier events
pub const COURIER_SIGNAL_CHANNEL: &str = "courier_signal";

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_workflow_state_creation() {
        let id = Uuid::new_v4();
        let state = CourierWorkflowState::new(id);

        assert_eq!(state.courier_id, id);
        assert_eq!(state.status, CourierStatus::Unavailable);
        assert!(state.is_running);
    }
}
