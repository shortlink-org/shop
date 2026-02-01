//! Get Location History Query
//!
//! Retrieves location history for a courier within a time range.

mod handler;
mod query;

pub use handler::{GetLocationHistoryError, Handler, Response};
pub use query::Query;
