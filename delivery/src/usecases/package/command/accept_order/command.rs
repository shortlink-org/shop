//! Accept Order Command
//!
//! Data structure representing the command to accept an order for delivery.

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::domain::model::package::{Address, DeliveryPeriod, Priority};
use crate::domain::model::vo::location::Location;

/// Command to accept an order from OMS
#[derive(Debug, Clone)]
pub struct Command {
    /// The order ID from OMS
    pub order_id: Uuid,
    /// Customer ID
    pub customer_id: Uuid,
    /// Pickup address
    pub pickup_address: AddressInput,
    /// Delivery address
    pub delivery_address: AddressInput,
    /// Delivery time window
    pub delivery_period: DeliveryPeriodInput,
    /// Package weight in kg
    pub weight_kg: f64,
    /// Package dimensions (LxWxH format)
    pub dimensions: String,
    /// Priority level
    pub priority: Priority,
}

/// Address input for the command
#[derive(Debug, Clone)]
pub struct AddressInput {
    /// Street address
    pub street: String,
    /// City
    pub city: String,
    /// Postal code
    pub postal_code: String,
    /// Country
    pub country: String,
    /// Latitude
    pub latitude: f64,
    /// Longitude
    pub longitude: f64,
}

/// Default GPS accuracy in meters when not provided
const DEFAULT_GPS_ACCURACY_METERS: f64 = 10.0;

impl AddressInput {
    /// Convert to domain Address
    pub fn to_domain(&self) -> Result<Address, String> {
        let location = Location::new(self.latitude, self.longitude, DEFAULT_GPS_ACCURACY_METERS)
            .map_err(|e| format!("Invalid location: {}", e))?;

        Ok(Address::new(
            self.street.clone(),
            self.city.clone(),
            self.postal_code.clone(),
            location,
        ))
    }

    /// Validate address fields
    pub fn validate(&self) -> Result<(), String> {
        if self.street.trim().is_empty() {
            return Err("Street is required".to_string());
        }
        if self.city.trim().is_empty() {
            return Err("City is required".to_string());
        }
        if self.country.trim().is_empty() {
            return Err("Country is required".to_string());
        }
        // Validate coordinates
        if self.latitude < -90.0 || self.latitude > 90.0 {
            return Err(format!("Invalid latitude: {}", self.latitude));
        }
        if self.longitude < -180.0 || self.longitude > 180.0 {
            return Err(format!("Invalid longitude: {}", self.longitude));
        }
        Ok(())
    }
}

/// Delivery period input for the command
#[derive(Debug, Clone)]
pub struct DeliveryPeriodInput {
    /// Start of delivery window
    pub start_time: DateTime<Utc>,
    /// End of delivery window
    pub end_time: DateTime<Utc>,
}

impl DeliveryPeriodInput {
    /// Convert to domain DeliveryPeriod
    pub fn to_domain(&self) -> Result<DeliveryPeriod, String> {
        DeliveryPeriod::new(self.start_time, self.end_time)
            .map_err(|e| format!("Invalid delivery period: {}", e))
    }

    /// Validate delivery period according to business rules
    pub fn validate(&self) -> Result<(), String> {
        let now = Utc::now();

        // Must be in the future
        if self.start_time <= now {
            return Err("Delivery period must be in the future".to_string());
        }

        // Start must be before end
        if self.start_time >= self.end_time {
            return Err("Start time must be before end time".to_string());
        }

        let duration = self.end_time - self.start_time;

        // Minimum duration: 1 hour
        if duration < chrono::Duration::hours(1) {
            return Err("Delivery period must be at least 1 hour".to_string());
        }

        // Maximum duration: 24 hours
        if duration > chrono::Duration::hours(24) {
            return Err("Delivery period cannot exceed 24 hours".to_string());
        }

        Ok(())
    }
}

impl Command {
    /// Create a new AcceptOrder command
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        order_id: Uuid,
        customer_id: Uuid,
        pickup_address: AddressInput,
        delivery_address: AddressInput,
        delivery_period: DeliveryPeriodInput,
        weight_kg: f64,
        dimensions: String,
        priority: Priority,
    ) -> Self {
        Self {
            order_id,
            customer_id,
            pickup_address,
            delivery_address,
            delivery_period,
            weight_kg,
            dimensions,
            priority,
        }
    }

    /// Validate the command according to business rules
    pub fn validate(&self) -> Result<(), String> {
        // Validate addresses
        self.pickup_address
            .validate()
            .map_err(|e| format!("Pickup address: {}", e))?;
        self.delivery_address
            .validate()
            .map_err(|e| format!("Delivery address: {}", e))?;

        // Validate delivery period
        self.delivery_period.validate()?;

        // Validate weight (must be positive)
        if self.weight_kg <= 0.0 {
            return Err("Package weight must be positive".to_string());
        }

        // Validate dimensions (optional but if provided, should not be empty)
        // Dimensions format: "LxWxH" e.g., "30x20x15"
        if !self.dimensions.is_empty() && !self.dimensions.contains('x') {
            return Err("Dimensions must be in LxWxH format".to_string());
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_valid_address() -> AddressInput {
        AddressInput {
            street: "123 Main St".to_string(),
            city: "Berlin".to_string(),
            postal_code: "10115".to_string(),
            country: "Germany".to_string(),
            latitude: 52.52,
            longitude: 13.405,
        }
    }

    fn create_valid_period() -> DeliveryPeriodInput {
        let now = Utc::now();
        DeliveryPeriodInput {
            start_time: now + chrono::Duration::hours(2),
            end_time: now + chrono::Duration::hours(4),
        }
    }

    #[test]
    fn test_address_validation_empty_street() {
        let mut addr = create_valid_address();
        addr.street = "".to_string();
        assert!(addr.validate().is_err());
    }

    #[test]
    fn test_address_validation_invalid_latitude() {
        let mut addr = create_valid_address();
        addr.latitude = 100.0;
        assert!(addr.validate().is_err());
    }

    #[test]
    fn test_delivery_period_in_past() {
        let now = Utc::now();
        let period = DeliveryPeriodInput {
            start_time: now - chrono::Duration::hours(1),
            end_time: now + chrono::Duration::hours(1),
        };
        assert!(period.validate().is_err());
    }

    #[test]
    fn test_delivery_period_too_short() {
        let now = Utc::now();
        let period = DeliveryPeriodInput {
            start_time: now + chrono::Duration::hours(1),
            end_time: now + chrono::Duration::minutes(90), // Only 30 min
        };
        assert!(period.validate().is_err());
    }

    #[test]
    fn test_delivery_period_too_long() {
        let now = Utc::now();
        let period = DeliveryPeriodInput {
            start_time: now + chrono::Duration::hours(1),
            end_time: now + chrono::Duration::hours(30), // 29 hours
        };
        assert!(period.validate().is_err());
    }

    #[test]
    fn test_valid_command() {
        let cmd = Command::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            create_valid_address(),
            create_valid_address(),
            create_valid_period(),
            2.5,
            "30x20x15".to_string(),
            Priority::Normal,
        );
        assert!(cmd.validate().is_ok());
    }

    #[test]
    fn test_invalid_weight() {
        let cmd = Command::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            create_valid_address(),
            create_valid_address(),
            create_valid_period(),
            0.0, // Invalid
            "30x20x15".to_string(),
            Priority::Normal,
        );
        assert!(cmd.validate().is_err());
    }
}
