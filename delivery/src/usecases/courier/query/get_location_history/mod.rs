//! Get Location History Query
//!
//! Retrieves a courier's location history for a given time range.

mod handler;
mod query;

pub use handler::{GetLocationHistoryError, Handler, Response};
pub use query::Query;
