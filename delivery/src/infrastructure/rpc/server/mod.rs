//! gRPC Server Implementation
//!
//! Implements the DeliveryService gRPC interface.
//! Handler logic is delegated to specialized modules in `handlers/`.

mod service;

use std::sync::Arc;

use crate::di::AppState;

/// gRPC service implementation
pub struct DeliveryServiceImpl {
    pub(crate) state: Arc<AppState>,
}

impl DeliveryServiceImpl {
    /// Create a new service instance
    pub fn new(state: Arc<AppState>) -> Self {
        Self { state }
    }
}
