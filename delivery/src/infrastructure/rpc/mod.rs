//! gRPC Infrastructure
//!
//! Contains gRPC service implementations and generated code.

pub mod converters;
pub(crate) mod handlers;
pub mod server;

// Include generated gRPC code
pub mod delivery_proto {
    tonic::include_proto!("infrastructure.rpc.delivery.v1");
}

pub use delivery_proto::delivery_service_server::{DeliveryService, DeliveryServiceServer};
pub use delivery_proto::*;
