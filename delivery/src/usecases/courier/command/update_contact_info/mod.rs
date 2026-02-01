//! Update Contact Info Command
//!
//! Updates courier contact information (phone, email, push token).

mod command;
mod handler;

pub use command::Command;
pub use handler::{Handler, Response, UpdateContactInfoError};
