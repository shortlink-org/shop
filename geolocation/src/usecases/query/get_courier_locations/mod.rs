//! Get Courier Locations Query
//!
//! Retrieves current locations for one or more couriers.

mod handler;
mod query;

pub use handler::{GetCourierLocationsError, Handler, LocationInfo, Response};
pub use query::Query;
