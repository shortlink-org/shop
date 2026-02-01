//! Update Work Schedule Command
//!
//! Updates courier work schedule (work hours, work zone, max distance).

mod command;
mod handler;

pub use command::Command;
pub use handler::{Handler, Response, UpdateWorkScheduleError};
