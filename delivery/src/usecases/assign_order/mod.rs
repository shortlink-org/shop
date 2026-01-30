//! Assign Order Use Case
//!
//! Assigns a package to a courier (automatically or manually).
//! Sends push notification to the courier upon assignment.
//!
//! ## Flow
//! 1. Load package from repository
//! 2. If auto-assign: use DispatchService to find nearest courier
//! 3. If manual: validate assignment using AssignmentValidationService
//! 4. Update package status to ASSIGNED
//! 5. Update courier status to BUSY
//! 6. Generate PackageAssignedEvent
//! 7. Send push notification

// TODO: Implement use case
// pub struct AssignOrderUseCase { ... }
