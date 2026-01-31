//! Assign Order Command
//!
//! Assigns a package to a courier (automatically or manually).

mod command;
mod handler;

pub use command::{AssignmentMode, Command};
pub use handler::{AssignOrderError, Handler, Response};
