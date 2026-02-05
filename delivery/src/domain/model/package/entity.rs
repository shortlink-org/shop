//! Package Aggregate Entity
//!
//! The Package aggregate represents a delivery package in the system.
//! It encapsulates package-related state and business rules.

use std::fmt;

use chrono::{DateTime, Utc};
use uuid::Uuid;

use super::state::{InvalidTransitionError, PackageStatus};
use crate::domain::model::vo::location::Location;

/// Unique identifier for a Package
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct PackageId(pub Uuid);

impl PackageId {
    /// Create a new random package ID
    pub fn new() -> Self {
        Self(Uuid::new_v4())
    }

    /// Create a package ID from an existing UUID
    pub fn from_uuid(id: Uuid) -> Self {
        Self(id)
    }
}

impl Default for PackageId {
    fn default() -> Self {
        Self::new()
    }
}

impl fmt::Display for PackageId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Priority level for packages
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum Priority {
    #[default]
    Normal,
    Urgent,
}

impl fmt::Display for Priority {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Priority::Normal => write!(f, "Normal"),
            Priority::Urgent => write!(f, "Urgent"),
        }
    }
}

/// Address for pickup/delivery
#[derive(Debug, Clone, PartialEq)]
pub struct Address {
    /// Street address
    pub street: String,
    /// City
    pub city: String,
    /// Postal code
    pub postal_code: String,
    /// Location coordinates
    pub location: Location,
}

impl Address {
    /// Create a new address
    pub fn new(street: String, city: String, postal_code: String, location: Location) -> Self {
        Self {
            street,
            city,
            postal_code,
            location,
        }
    }
}

/// Delivery time window
#[derive(Debug, Clone, PartialEq)]
pub struct DeliveryPeriod {
    /// Start of delivery window
    pub start: DateTime<Utc>,
    /// End of delivery window
    pub end: DateTime<Utc>,
}

impl DeliveryPeriod {
    /// Create a new delivery period
    pub fn new(start: DateTime<Utc>, end: DateTime<Utc>) -> Result<Self, PackageError> {
        if start >= end {
            return Err(PackageError::InvalidDeliveryPeriod(
                "Start time must be before end time".to_string(),
            ));
        }
        Ok(Self { start, end })
    }

    /// Check if current time is within the delivery window
    pub fn is_within(&self, time: DateTime<Utc>) -> bool {
        time >= self.start && time <= self.end
    }

    /// Get the start of the delivery window
    pub fn start(&self) -> &DateTime<Utc> {
        &self.start
    }

    /// Get the end of the delivery window
    pub fn end(&self) -> &DateTime<Utc> {
        &self.end
    }
}

/// Package errors
#[derive(Debug, Clone, PartialEq)]
pub enum PackageError {
    /// Invalid delivery period
    InvalidDeliveryPeriod(String),
    /// Invalid state transition
    InvalidTransition(InvalidTransitionError),
    /// Already assigned
    AlreadyAssigned,
    /// Not assigned
    NotAssigned,
}

impl fmt::Display for PackageError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PackageError::InvalidDeliveryPeriod(msg) => {
                write!(f, "Invalid delivery period: {}", msg)
            }
            PackageError::InvalidTransition(e) => write!(f, "{}", e),
            PackageError::AlreadyAssigned => write!(f, "Package is already assigned"),
            PackageError::NotAssigned => write!(f, "Package is not assigned"),
        }
    }
}

impl std::error::Error for PackageError {}

impl From<InvalidTransitionError> for PackageError {
    fn from(e: InvalidTransitionError) -> Self {
        PackageError::InvalidTransition(e)
    }
}

/// Package aggregate - represents a delivery package in the system
#[derive(Debug, Clone)]
pub struct Package {
    /// Unique identifier
    id: PackageId,
    /// Order ID from OMS
    order_id: Uuid,
    /// Customer ID
    customer_id: Uuid,
    /// Customer phone number for delivery
    customer_phone: Option<String>,
    /// Recipient name (from OMS)
    recipient_name: Option<String>,
    /// Recipient phone (from OMS)
    recipient_phone: Option<String>,
    /// Recipient email (from OMS)
    recipient_email: Option<String>,
    /// Pickup address
    pickup_address: Address,
    /// Delivery address
    delivery_address: Address,
    /// Delivery time window
    delivery_period: DeliveryPeriod,
    /// Package weight in kg
    weight_kg: f64,
    /// Priority level
    priority: Priority,
    /// Current status
    status: PackageStatus,
    /// Assigned courier ID
    courier_id: Option<Uuid>,
    /// Work zone (derived from delivery address)
    zone: String,
    /// Creation timestamp
    created_at: DateTime<Utc>,
    /// Last update timestamp
    updated_at: DateTime<Utc>,
    /// Assigned at timestamp
    assigned_at: Option<DateTime<Utc>>,
    /// Delivered at timestamp
    delivered_at: Option<DateTime<Utc>>,
    /// Not delivered reason
    not_delivered_reason: Option<String>,
    /// Version for optimistic locking
    version: u32,
}

impl Package {
    /// Create a new package
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        order_id: Uuid,
        customer_id: Uuid,
        customer_phone: Option<String>,
        recipient_name: Option<String>,
        recipient_phone: Option<String>,
        recipient_email: Option<String>,
        pickup_address: Address,
        delivery_address: Address,
        delivery_period: DeliveryPeriod,
        weight_kg: f64,
        priority: Priority,
        zone: String,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: PackageId::new(),
            order_id,
            customer_id,
            customer_phone,
            recipient_name,
            recipient_phone,
            recipient_email,
            pickup_address,
            delivery_address,
            delivery_period,
            weight_kg,
            priority,
            status: PackageStatus::Accepted,
            courier_id: None,
            zone,
            created_at: now,
            updated_at: now,
            assigned_at: None,
            delivered_at: None,
            not_delivered_reason: None,
            version: 1,
        }
    }

    /// Reconstruct from persistence
    #[allow(clippy::too_many_arguments)]
    pub fn reconstitute(
        id: PackageId,
        order_id: Uuid,
        customer_id: Uuid,
        customer_phone: Option<String>,
        recipient_name: Option<String>,
        recipient_phone: Option<String>,
        recipient_email: Option<String>,
        pickup_address: Address,
        delivery_address: Address,
        delivery_period: DeliveryPeriod,
        weight_kg: f64,
        priority: Priority,
        status: PackageStatus,
        courier_id: Option<Uuid>,
        zone: String,
        created_at: DateTime<Utc>,
        updated_at: DateTime<Utc>,
        assigned_at: Option<DateTime<Utc>>,
        delivered_at: Option<DateTime<Utc>>,
        not_delivered_reason: Option<String>,
        version: u32,
    ) -> Self {
        Self {
            id,
            order_id,
            customer_id,
            customer_phone,
            recipient_name,
            recipient_phone,
            recipient_email,
            pickup_address,
            delivery_address,
            delivery_period,
            weight_kg,
            priority,
            status,
            courier_id,
            zone,
            created_at,
            updated_at,
            assigned_at,
            delivered_at,
            not_delivered_reason,
            version,
        }
    }

    // === Getters ===

    pub fn id(&self) -> &PackageId {
        &self.id
    }

    pub fn order_id(&self) -> Uuid {
        self.order_id
    }

    pub fn customer_id(&self) -> Uuid {
        self.customer_id
    }

    pub fn customer_phone(&self) -> Option<&str> {
        self.customer_phone.as_deref()
    }

    pub fn recipient_name(&self) -> Option<&str> {
        self.recipient_name.as_deref()
    }

    pub fn recipient_phone(&self) -> Option<&str> {
        self.recipient_phone.as_deref()
    }

    pub fn recipient_email(&self) -> Option<&str> {
        self.recipient_email.as_deref()
    }

    pub fn pickup_address(&self) -> &Address {
        &self.pickup_address
    }

    pub fn delivery_address(&self) -> &Address {
        &self.delivery_address
    }

    pub fn delivery_period(&self) -> &DeliveryPeriod {
        &self.delivery_period
    }

    pub fn weight_kg(&self) -> f64 {
        self.weight_kg
    }

    pub fn priority(&self) -> Priority {
        self.priority
    }

    pub fn status(&self) -> PackageStatus {
        self.status
    }

    pub fn courier_id(&self) -> Option<Uuid> {
        self.courier_id
    }

    pub fn zone(&self) -> &str {
        &self.zone
    }

    pub fn created_at(&self) -> DateTime<Utc> {
        self.created_at
    }

    pub fn updated_at(&self) -> DateTime<Utc> {
        self.updated_at
    }

    pub fn assigned_at(&self) -> Option<DateTime<Utc>> {
        self.assigned_at
    }

    pub fn delivered_at(&self) -> Option<DateTime<Utc>> {
        self.delivered_at
    }

    pub fn not_delivered_reason(&self) -> Option<&str> {
        self.not_delivered_reason.as_deref()
    }

    pub fn version(&self) -> u32 {
        self.version
    }

    // === Business Methods ===

    /// Move package to pool (ready for assignment)
    pub fn move_to_pool(&mut self) -> Result<(), PackageError> {
        self.status = self.status.transition_to(PackageStatus::InPool)?;
        self.touch();
        Ok(())
    }

    /// Assign package to a courier
    pub fn assign_to(&mut self, courier_id: Uuid) -> Result<(), PackageError> {
        if self.courier_id.is_some() {
            return Err(PackageError::AlreadyAssigned);
        }
        self.status = self.status.transition_to(PackageStatus::Assigned)?;
        self.courier_id = Some(courier_id);
        self.assigned_at = Some(Utc::now());
        self.touch();
        Ok(())
    }

    /// Start transit
    pub fn start_transit(&mut self) -> Result<(), PackageError> {
        self.status = self.status.transition_to(PackageStatus::InTransit)?;
        self.touch();
        Ok(())
    }

    /// Mark as delivered
    pub fn mark_delivered(&mut self) -> Result<(), PackageError> {
        self.status = self.status.transition_to(PackageStatus::Delivered)?;
        self.delivered_at = Some(Utc::now());
        self.touch();
        Ok(())
    }

    /// Mark as not delivered
    pub fn mark_not_delivered(&mut self, reason: String) -> Result<(), PackageError> {
        self.status = self.status.transition_to(PackageStatus::NotDelivered)?;
        self.not_delivered_reason = Some(reason);
        self.touch();
        Ok(())
    }

    /// Return to pool (after failed delivery)
    pub fn return_to_pool(&mut self) -> Result<(), PackageError> {
        self.status = self.status.transition_to(PackageStatus::RequiresHandling)?;
        self.touch();
        // Then back to pool
        self.status = self.status.transition_to(PackageStatus::InPool)?;
        self.courier_id = None;
        self.assigned_at = None;
        self.not_delivered_reason = None;
        self.touch();
        Ok(())
    }

    /// Check if package can be assigned
    pub fn can_be_assigned(&self) -> bool {
        self.status == PackageStatus::InPool && self.courier_id.is_none()
    }

    // === Private Methods ===

    fn touch(&mut self) {
        self.updated_at = Utc::now();
        self.version += 1;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_address() -> Address {
        Address::new(
            "123 Main St".to_string(),
            "Berlin".to_string(),
            "10115".to_string(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
        )
    }

    fn create_test_package() -> Package {
        let now = Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(1),
            now + chrono::Duration::hours(3),
        )
        .unwrap();

        Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_test_address(),
            create_test_address(),
            period,
            2.5,
            Priority::Normal,
            "Berlin-Mitte".to_string(),
        )
    }

    #[test]
    fn test_package_creation() {
        let package = create_test_package();
        assert_eq!(package.status(), PackageStatus::Accepted);
        assert!(package.courier_id().is_none());
        assert_eq!(package.version(), 1);
    }

    #[test]
    fn test_package_lifecycle() {
        let mut package = create_test_package();
        let courier_id = Uuid::new_v4();

        // Move to pool
        assert!(package.move_to_pool().is_ok());
        assert_eq!(package.status(), PackageStatus::InPool);

        // Assign to courier
        assert!(package.assign_to(courier_id).is_ok());
        assert_eq!(package.status(), PackageStatus::Assigned);
        assert_eq!(package.courier_id(), Some(courier_id));

        // Start transit
        assert!(package.start_transit().is_ok());
        assert_eq!(package.status(), PackageStatus::InTransit);

        // Deliver
        assert!(package.mark_delivered().is_ok());
        assert_eq!(package.status(), PackageStatus::Delivered);
        assert!(package.delivered_at().is_some());
    }

    #[test]
    fn test_cannot_assign_twice() {
        let mut package = create_test_package();
        package.move_to_pool().unwrap();
        package.assign_to(Uuid::new_v4()).unwrap();

        assert!(matches!(
            package.assign_to(Uuid::new_v4()),
            Err(PackageError::AlreadyAssigned)
        ));
    }
}
