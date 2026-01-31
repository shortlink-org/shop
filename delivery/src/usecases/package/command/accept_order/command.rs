//! Accept Order Command
//!
//! Data structure representing the command to accept an order for delivery.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Command to accept an order from OMS
#[derive(Debug, Clone)]
pub struct Command {
    /// The order ID from OMS
    pub order_id: Uuid,
    /// Customer ID
    pub customer_id: Uuid,
    /// Pickup location
    pub pickup_location: Location,
    /// Delivery location
    pub delivery_location: Location,
    /// Delivery zone identifier
    pub delivery_zone: String,
    /// Package weight in kg
    pub weight_kg: f64,
    /// Priority level (1-5, where 5 is highest)
    pub priority: u8,
    /// Special instructions for delivery
    pub instructions: Option<String>,
}

impl Command {
    /// Create a new AcceptOrder command
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        order_id: Uuid,
        customer_id: Uuid,
        pickup_location: Location,
        delivery_location: Location,
        delivery_zone: String,
        weight_kg: f64,
        priority: u8,
        instructions: Option<String>,
    ) -> Self {
        Self {
            order_id,
            customer_id,
            pickup_location,
            delivery_location,
            delivery_zone,
            weight_kg,
            priority,
            instructions,
        }
    }
}
