//! Assign Order Command
//!
//! Data structure representing the command to assign a package to a courier.

use uuid::Uuid;

/// Assignment mode for the package
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AssignmentMode {
    /// Automatically find the best courier using DispatchService
    Auto,
    /// Manually assign to a specific courier
    Manual,
}

/// Command to assign an order to a courier
#[derive(Debug, Clone)]
pub struct Command {
    /// The package ID to assign
    pub package_id: Uuid,
    /// Assignment mode
    pub mode: AssignmentMode,
    /// Specific courier ID (required for Manual mode)
    pub courier_id: Option<Uuid>,
}

impl Command {
    /// Create a new auto-assignment command
    pub fn auto_assign(package_id: Uuid) -> Self {
        Self {
            package_id,
            mode: AssignmentMode::Auto,
            courier_id: None,
        }
    }

    /// Create a new manual assignment command
    pub fn manual_assign(package_id: Uuid, courier_id: Uuid) -> Self {
        Self {
            package_id,
            mode: AssignmentMode::Manual,
            courier_id: Some(courier_id),
        }
    }
}
