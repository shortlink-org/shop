//! Activate Courier Command
//!
//! Activates a courier by setting their status to FREE.

mod command;
mod handler;

pub use command::Command;
pub use handler::{ActivateCourierError, Handler, Response};
