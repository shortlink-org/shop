//! Domain Services
//!
//! Domain services contain business logic that:
//! - Operates on multiple aggregates
//! - Doesn't belong to a single entity
//! - Has no infrastructure dependencies
//!
//! Unlike Use Cases, domain services:
//! - Don't know about repositories or external services
//! - Work only with domain objects passed as parameters
//! - Contain pure business logic

pub mod assignment_validation;
pub mod dispatch;
