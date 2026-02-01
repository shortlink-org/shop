//! Deliver Order Command
//!
//! Confirms delivery by courier (successful or failed).

mod command;
mod handler;

pub use command::{Command, DeliveryResult, NotDeliveredReason};
pub use handler::{DeliverOrderError, Handler, Response};
