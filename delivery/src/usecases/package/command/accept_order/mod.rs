//! Accept Order Command
//!
//! Accepts an order from OMS for delivery.

mod command;
mod handler;

pub use command::Command;
pub use handler::{AcceptOrderError, Handler, Response};
