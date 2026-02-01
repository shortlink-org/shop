//! Courier Queries
//!
//! Read-only queries for courier data.

pub mod get_courier;
pub mod get_location;
pub mod get_location_history;
pub mod get_pool;

// Re-export main types
pub use get_courier::{
    GetCourierError, Handler as GetCourierHandler, Query as GetCourierQuery,
    Response as GetCourierResponse,
};
pub use get_location::{
    BatchQuery as GetLocationBatchQuery, BatchResponse as GetLocationBatchResponse,
    GetLocationError, Handler as GetLocationHandler, Query as GetLocationQuery,
    Response as GetLocationResponse,
};
pub use get_location_history::{
    GetLocationHistoryError, Handler as GetLocationHistoryHandler,
    Query as GetLocationHistoryQuery, Response as GetLocationHistoryResponse,
};
pub use get_pool::{
    CourierFilter, CourierWithState, GetCourierPoolError, Handler as GetPoolHandler,
    Query as GetPoolQuery, Response as GetPoolResponse,
};
