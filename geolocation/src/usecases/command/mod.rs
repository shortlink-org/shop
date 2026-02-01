//! Commands
//!
//! Commands that modify state in the Geolocation service.

pub mod create_geofence;
pub mod save_location;

pub use create_geofence::{
    Command as CreateGeofenceCommand, CreateGeofenceError, Handler as CreateGeofenceHandler,
    Response as CreateGeofenceResponse, ShapeInput,
};
pub use save_location::{
    Command as SaveLocationCommand, Handler as SaveLocationHandler,
    Response as SaveLocationResponse, SaveLocationError,
};
