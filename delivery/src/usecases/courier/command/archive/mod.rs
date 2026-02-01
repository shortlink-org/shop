//! Archive Courier Command
//!
//! Archives a courier by setting their status to ARCHIVED (soft delete).

mod command;
mod handler;

pub use command::Command;
pub use handler::{ArchiveCourierError, Handler, Response};
