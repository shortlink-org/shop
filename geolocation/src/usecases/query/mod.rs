//! Queries
//!
//! Read-only queries in the Geolocation service.

pub mod get_courier_locations;
pub mod get_location_history;

pub use get_courier_locations::{
    GetCourierLocationsError, Handler as GetCourierLocationsHandler, LocationInfo,
    Query as GetCourierLocationsQuery, Response as GetCourierLocationsResponse,
};
pub use get_location_history::{
    GetLocationHistoryError, Handler as GetLocationHistoryHandler,
    Query as GetLocationHistoryQuery, Response as GetLocationHistoryResponse,
};
