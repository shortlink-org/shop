//! Update Courier Location Command
//!
//! Updates courier's GPS location in real-time.

mod command;
mod handler;

pub use command::Command;
pub use handler::{Handler, UpdateCourierLocationError};
