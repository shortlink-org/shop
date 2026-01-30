//! Accept Order Use Case
//!
//! Accepts an order from OMS for delivery.
//! Creates a new package in the pool with ACCEPTED status.
//!
//! ## Flow
//! 1. Validate order data
//! 2. Create Package aggregate
//! 3. Transition to IN_POOL status
//! 4. Save to repository
//! 5. Generate PackageAcceptedEvent

// TODO: Implement use case
// pub struct AcceptOrderUseCase { ... }
