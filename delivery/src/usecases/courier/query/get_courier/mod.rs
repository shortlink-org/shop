//! Get Courier Query
//!
//! Retrieves a single courier by ID.

mod handler;
mod query;

pub use handler::{GetCourierError, Handler, Response};
pub use query::Query;
