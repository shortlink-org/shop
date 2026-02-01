//! Pick Up Order Command
//!
//! Confirms package pickup by courier (ASSIGNED -> IN_TRANSIT).

mod command;
mod handler;

pub use command::Command;
pub use handler::{Handler, PickUpOrderError, Response};
