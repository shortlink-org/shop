//! Domain Models
//!
//! This module contains all domain models including:
//! - Aggregates (Package, Courier)
//! - Value Objects (Location, Address, etc.)
//! - Proto-generated models (delivery/)

pub mod courier;
pub mod package;
pub mod vo;

// Proto generated code will be here after build
// pub mod delivery;
