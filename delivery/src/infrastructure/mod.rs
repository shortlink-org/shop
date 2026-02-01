//! Infrastructure Layer
//!
//! Contains implementations of port interfaces (adapters).
//! This layer handles external concerns like databases, caches, and external services.

pub mod cache;
pub mod messaging;
pub mod notifications;
pub mod repository;
pub mod rpc;
