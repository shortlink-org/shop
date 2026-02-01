//! Get Courier Location Query
//!
//! Retrieves a courier's current location from cache or repository.

mod handler;
mod query;

pub use handler::{BatchResponse, GetLocationError, Handler, Response};
pub use query::{BatchQuery, Query};
