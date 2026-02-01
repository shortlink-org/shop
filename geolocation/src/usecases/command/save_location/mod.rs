//! Save Location Command
//!
//! Saves courier's current location to database and cache.

mod command;
mod handler;

pub use command::Command;
pub use handler::{Handler, Response, SaveLocationError};
