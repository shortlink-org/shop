//! Courier Aggregate Entity
//!
//! The Courier aggregate represents a delivery courier in the system.
//! It encapsulates all courier-related state and business rules.

use std::fmt;

use chrono::Datelike;

use crate::domain::model::vo::TransportType;

use super::{CapacityError, CourierCapacity, CourierStatus};

/// Unique identifier for a Courier
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct CourierId(pub uuid::Uuid);

impl CourierId {
    /// Create a new random courier ID
    pub fn new() -> Self {
        Self(uuid::Uuid::new_v4())
    }

    /// Create a courier ID from an existing UUID
    pub fn from_uuid(id: uuid::Uuid) -> Self {
        Self(id)
    }

    /// Get the underlying UUID
    pub fn as_uuid(&self) -> &uuid::Uuid {
        &self.0
    }
}

impl Default for CourierId {
    fn default() -> Self {
        Self::new()
    }
}

impl fmt::Display for CourierId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Work hours for a courier
#[derive(Debug, Clone, PartialEq)]
pub struct WorkHours {
    /// Start time (e.g., "09:00")
    pub start: chrono::NaiveTime,
    /// End time (e.g., "18:00")
    pub end: chrono::NaiveTime,
    /// Work days (1=Monday, 7=Sunday)
    pub days: Vec<u8>,
}

impl WorkHours {
    /// Create new work hours
    pub fn new(
        start: chrono::NaiveTime,
        end: chrono::NaiveTime,
        days: Vec<u8>,
    ) -> Result<Self, CourierError> {
        if start >= end {
            return Err(CourierError::InvalidWorkHours(
                "Start time must be before end time".to_string(),
            ));
        }
        if days.is_empty() {
            return Err(CourierError::InvalidWorkHours(
                "At least one work day required".to_string(),
            ));
        }
        for day in &days {
            if *day < 1 || *day > 7 {
                return Err(CourierError::InvalidWorkHours(format!(
                    "Invalid day: {}. Must be 1-7",
                    day
                )));
            }
        }
        Ok(Self { start, end, days })
    }

    /// Check if courier is working at the given time
    pub fn is_working_at(&self, time: chrono::NaiveDateTime) -> bool {
        let weekday = time.weekday().number_from_monday() as u8;
        let time_of_day = time.time();

        self.days.contains(&weekday) && time_of_day >= self.start && time_of_day <= self.end
    }
}

/// Courier aggregate errors
#[derive(Debug, Clone, PartialEq)]
pub enum CourierError {
    /// Invalid email format
    InvalidEmail(String),
    /// Invalid phone format
    InvalidPhone(String),
    /// Invalid work hours
    InvalidWorkHours(String),
    /// Invalid max distance
    InvalidMaxDistance(String),
    /// Capacity error
    CapacityError(CapacityError),
    /// Status transition not allowed
    InvalidStatusTransition { from: CourierStatus, to: CourierStatus },
    /// Courier not available for assignment
    NotAvailable,
}

impl fmt::Display for CourierError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CourierError::InvalidEmail(msg) => write!(f, "Invalid email: {}", msg),
            CourierError::InvalidPhone(msg) => write!(f, "Invalid phone: {}", msg),
            CourierError::InvalidWorkHours(msg) => write!(f, "Invalid work hours: {}", msg),
            CourierError::InvalidMaxDistance(msg) => write!(f, "Invalid max distance: {}", msg),
            CourierError::CapacityError(e) => write!(f, "Capacity error: {}", e),
            CourierError::InvalidStatusTransition { from, to } => {
                write!(f, "Invalid status transition from {} to {}", from, to)
            }
            CourierError::NotAvailable => write!(f, "Courier is not available"),
        }
    }
}

impl std::error::Error for CourierError {}

impl From<CapacityError> for CourierError {
    fn from(e: CapacityError) -> Self {
        CourierError::CapacityError(e)
    }
}

/// Courier aggregate - represents a delivery courier in the system
#[derive(Debug, Clone)]
pub struct Courier {
    /// Unique identifier
    id: CourierId,
    /// Courier's full name
    name: String,
    /// Phone number (international format)
    phone: String,
    /// Email address
    email: String,
    /// Type of transport
    transport_type: TransportType,
    /// Maximum delivery distance in km
    max_distance_km: f64,
    /// Work zone (region/district)
    work_zone: String,
    /// Work hours and days
    work_hours: WorkHours,
    /// Push notification token (FCM/APNS)
    push_token: Option<String>,
    /// Current status
    status: CourierStatus,
    /// Capacity management
    capacity: CourierCapacity,
    /// Performance rating (0.0 - 5.0)
    rating: f64,
    /// Number of successful deliveries
    successful_deliveries: u32,
    /// Number of failed deliveries
    failed_deliveries: u32,
    /// Creation timestamp
    created_at: chrono::DateTime<chrono::Utc>,
    /// Last update timestamp
    updated_at: chrono::DateTime<chrono::Utc>,
    /// Version for optimistic locking
    version: u32,
}

/// Builder for creating a new Courier
#[derive(Debug)]
pub struct CourierBuilder {
    name: String,
    phone: String,
    email: String,
    transport_type: TransportType,
    max_distance_km: f64,
    work_zone: String,
    work_hours: WorkHours,
    push_token: Option<String>,
}

impl CourierBuilder {
    /// Create a new courier builder with required fields
    pub fn new(
        name: String,
        phone: String,
        email: String,
        transport_type: TransportType,
        max_distance_km: f64,
        work_zone: String,
        work_hours: WorkHours,
    ) -> Self {
        Self {
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
            push_token: None,
        }
    }

    /// Set push notification token
    pub fn with_push_token(mut self, token: String) -> Self {
        self.push_token = Some(token);
        self
    }

    /// Build the Courier aggregate
    pub fn build(self) -> Result<Courier, CourierError> {
        // Validate email format (basic validation)
        if !self.email.contains('@') || !self.email.contains('.') {
            return Err(CourierError::InvalidEmail(
                "Invalid email format".to_string(),
            ));
        }

        // Validate phone (basic validation - should start with +)
        if !self.phone.starts_with('+') || self.phone.len() < 10 {
            return Err(CourierError::InvalidPhone(
                "Phone must be in international format starting with +".to_string(),
            ));
        }

        // Validate max distance
        if self.max_distance_km <= 0.0 {
            return Err(CourierError::InvalidMaxDistance(
                "Max distance must be positive".to_string(),
            ));
        }

        // Determine max load based on transport type
        let max_load = match self.transport_type {
            TransportType::Walking => 1,
            TransportType::Bicycle => 2,
            TransportType::Motorcycle => 3,
            TransportType::Car => 5,
        };

        let now = chrono::Utc::now();

        Ok(Courier {
            id: CourierId::new(),
            name: self.name,
            phone: self.phone,
            email: self.email,
            transport_type: self.transport_type,
            max_distance_km: self.max_distance_km,
            work_zone: self.work_zone,
            work_hours: self.work_hours,
            push_token: self.push_token,
            status: CourierStatus::Unavailable, // Initial status
            capacity: CourierCapacity::new(max_load),
            rating: 0.0,
            successful_deliveries: 0,
            failed_deliveries: 0,
            created_at: now,
            updated_at: now,
            version: 1,
        })
    }
}

impl Courier {
    /// Create a builder for a new Courier
    pub fn builder(
        name: String,
        phone: String,
        email: String,
        transport_type: TransportType,
        max_distance_km: f64,
        work_zone: String,
        work_hours: WorkHours,
    ) -> CourierBuilder {
        CourierBuilder::new(
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
        )
    }

    /// Reconstruct a Courier from persistence (used by repository)
    #[allow(clippy::too_many_arguments)]
    pub fn reconstitute(
        id: CourierId,
        name: String,
        phone: String,
        email: String,
        transport_type: TransportType,
        max_distance_km: f64,
        work_zone: String,
        work_hours: WorkHours,
        push_token: Option<String>,
        status: CourierStatus,
        capacity: CourierCapacity,
        rating: f64,
        successful_deliveries: u32,
        failed_deliveries: u32,
        created_at: chrono::DateTime<chrono::Utc>,
        updated_at: chrono::DateTime<chrono::Utc>,
        version: u32,
    ) -> Self {
        Self {
            id,
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
            push_token,
            status,
            capacity,
            rating,
            successful_deliveries,
            failed_deliveries,
            created_at,
            updated_at,
            version,
        }
    }

    // === Getters ===

    /// Get courier ID
    pub fn id(&self) -> &CourierId {
        &self.id
    }

    /// Get courier name
    pub fn name(&self) -> &str {
        &self.name
    }

    /// Get courier phone
    pub fn phone(&self) -> &str {
        &self.phone
    }

    /// Get courier email
    pub fn email(&self) -> &str {
        &self.email
    }

    /// Get transport type
    pub fn transport_type(&self) -> TransportType {
        self.transport_type
    }

    /// Get max distance in km
    pub fn max_distance_km(&self) -> f64 {
        self.max_distance_km
    }

    /// Get work zone
    pub fn work_zone(&self) -> &str {
        &self.work_zone
    }

    /// Get work hours
    pub fn work_hours(&self) -> &WorkHours {
        &self.work_hours
    }

    /// Get push token
    pub fn push_token(&self) -> Option<&str> {
        self.push_token.as_deref()
    }

    /// Get current status
    pub fn status(&self) -> CourierStatus {
        self.status
    }

    /// Get capacity
    pub fn capacity(&self) -> &CourierCapacity {
        &self.capacity
    }

    /// Get max load
    pub fn max_load(&self) -> u32 {
        self.capacity.max_load()
    }

    /// Get current load
    pub fn current_load(&self) -> u32 {
        self.capacity.current_load()
    }

    /// Get rating
    pub fn rating(&self) -> f64 {
        self.rating
    }

    /// Get successful deliveries count
    pub fn successful_deliveries(&self) -> u32 {
        self.successful_deliveries
    }

    /// Get failed deliveries count
    pub fn failed_deliveries(&self) -> u32 {
        self.failed_deliveries
    }

    /// Get creation timestamp
    pub fn created_at(&self) -> chrono::DateTime<chrono::Utc> {
        self.created_at
    }

    /// Get last update timestamp
    pub fn updated_at(&self) -> chrono::DateTime<chrono::Utc> {
        self.updated_at
    }

    /// Get version for optimistic locking
    pub fn version(&self) -> u32 {
        self.version
    }

    // === Business Methods ===

    /// Go online (become available for deliveries)
    pub fn go_online(&mut self) -> Result<(), CourierError> {
        match self.status {
            CourierStatus::Unavailable => {
                self.status = CourierStatus::Free;
                self.touch();
                Ok(())
            }
            _ => Err(CourierError::InvalidStatusTransition {
                from: self.status,
                to: CourierStatus::Free,
            }),
        }
    }

    /// Go offline (become unavailable)
    pub fn go_offline(&mut self) -> Result<(), CourierError> {
        match self.status {
            CourierStatus::Free => {
                self.status = CourierStatus::Unavailable;
                self.touch();
                Ok(())
            }
            CourierStatus::Busy => Err(CourierError::InvalidStatusTransition {
                from: self.status,
                to: CourierStatus::Unavailable,
            }),
            CourierStatus::Unavailable => Ok(()), // Already offline
            CourierStatus::Archived => Err(CourierError::InvalidStatusTransition {
                from: self.status,
                to: CourierStatus::Unavailable,
            }),
        }
    }

    /// Accept a package assignment
    pub fn accept_package(&mut self) -> Result<(), CourierError> {
        if self.status != CourierStatus::Free && self.status != CourierStatus::Busy {
            return Err(CourierError::NotAvailable);
        }

        self.capacity.add_package()?;

        if !self.capacity.can_accept() {
            self.status = CourierStatus::Busy;
        }

        self.touch();
        Ok(())
    }

    /// Complete a delivery (release a package)
    pub fn complete_delivery(&mut self, success: bool) -> Result<(), CourierError> {
        self.capacity.release_package()?;

        if success {
            self.successful_deliveries += 1;
        } else {
            self.failed_deliveries += 1;
        }

        // Update rating (simple moving average)
        self.update_rating();

        // If no more packages and was busy, become free
        if self.capacity.is_empty() && self.status == CourierStatus::Busy {
            self.status = CourierStatus::Free;
        }

        self.touch();
        Ok(())
    }

    /// Check if courier can accept new assignments
    pub fn can_accept_assignment(&self) -> bool {
        self.status.can_accept_assignment() && self.capacity.can_accept()
    }

    /// Update push token
    pub fn update_push_token(&mut self, token: Option<String>) {
        self.push_token = token;
        self.touch();
    }

    /// Update work hours
    pub fn update_work_hours(&mut self, work_hours: WorkHours) {
        self.work_hours = work_hours;
        self.touch();
    }

    /// Update phone number
    pub fn update_phone(&mut self, phone: String) {
        self.phone = phone;
        self.touch();
    }

    /// Update email address
    pub fn update_email(&mut self, email: String) {
        self.email = email;
        self.touch();
    }

    /// Update work zone
    pub fn update_work_zone(&mut self, work_zone: String) {
        self.work_zone = work_zone;
        self.touch();
    }

    /// Update max distance
    pub fn update_max_distance(&mut self, max_distance_km: f64) {
        self.max_distance_km = max_distance_km;
        self.touch();
    }

    /// Change transport type and recalculate max load
    pub fn change_transport_type(&mut self, transport_type: TransportType) {
        self.transport_type = transport_type;

        // Recalculate max load based on new transport type
        let new_max_load = match transport_type {
            TransportType::Walking => 1,
            TransportType::Bicycle => 2,
            TransportType::Motorcycle => 3,
            TransportType::Car => 5,
        };

        self.capacity = CourierCapacity::new(new_max_load);
        self.touch();
    }

    // === Private Methods ===

    /// Update the updated_at timestamp and increment version
    fn touch(&mut self) {
        self.updated_at = chrono::Utc::now();
        self.version += 1;
    }

    /// Update rating based on successful/failed deliveries
    fn update_rating(&mut self) {
        let total = self.successful_deliveries + self.failed_deliveries;
        if total > 0 {
            // Rating is success rate * 5 (0-5 scale)
            self.rating = (self.successful_deliveries as f64 / total as f64) * 5.0;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_work_hours() -> WorkHours {
        WorkHours::new(
            chrono::NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            chrono::NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5], // Monday to Friday
        )
        .unwrap()
    }

    fn create_test_courier() -> Courier {
        Courier::builder(
            "John Doe".to_string(),
            "+491234567890".to_string(),
            "john@example.com".to_string(),
            TransportType::Bicycle,
            10.0,
            "Berlin-Mitte".to_string(),
            create_work_hours(),
        )
        .build()
        .unwrap()
    }

    #[test]
    fn test_courier_creation() {
        let courier = create_test_courier();

        assert_eq!(courier.name(), "John Doe");
        assert_eq!(courier.phone(), "+491234567890");
        assert_eq!(courier.email(), "john@example.com");
        assert_eq!(courier.transport_type(), TransportType::Bicycle);
        assert_eq!(courier.max_distance_km(), 10.0);
        assert_eq!(courier.work_zone(), "Berlin-Mitte");
        assert_eq!(courier.status(), CourierStatus::Unavailable);
        assert_eq!(courier.max_load(), 2); // Bicycle = 2
        assert_eq!(courier.current_load(), 0);
        assert_eq!(courier.rating(), 0.0);
        assert_eq!(courier.version(), 1);
    }

    #[test]
    fn test_go_online_offline() {
        let mut courier = create_test_courier();

        // Initially unavailable
        assert_eq!(courier.status(), CourierStatus::Unavailable);

        // Go online
        assert!(courier.go_online().is_ok());
        assert_eq!(courier.status(), CourierStatus::Free);
        assert_eq!(courier.version(), 2);

        // Go offline
        assert!(courier.go_offline().is_ok());
        assert_eq!(courier.status(), CourierStatus::Unavailable);
        assert_eq!(courier.version(), 3);
    }

    #[test]
    fn test_accept_package() {
        let mut courier = create_test_courier();
        courier.go_online().unwrap();

        // Accept first package
        assert!(courier.accept_package().is_ok());
        assert_eq!(courier.current_load(), 1);
        assert_eq!(courier.status(), CourierStatus::Free); // Still free, capacity = 2

        // Accept second package (capacity full)
        assert!(courier.accept_package().is_ok());
        assert_eq!(courier.current_load(), 2);
        assert_eq!(courier.status(), CourierStatus::Busy); // Now busy

        // Cannot accept more
        assert!(courier.accept_package().is_err());
    }

    #[test]
    fn test_complete_delivery() {
        let mut courier = create_test_courier();
        courier.go_online().unwrap();
        courier.accept_package().unwrap();
        courier.accept_package().unwrap();

        assert_eq!(courier.status(), CourierStatus::Busy);

        // Complete first delivery successfully
        assert!(courier.complete_delivery(true).is_ok());
        assert_eq!(courier.current_load(), 1);
        assert_eq!(courier.successful_deliveries(), 1);
        assert_eq!(courier.status(), CourierStatus::Busy); // Still has packages

        // Complete second delivery with failure
        assert!(courier.complete_delivery(false).is_ok());
        assert_eq!(courier.current_load(), 0);
        assert_eq!(courier.failed_deliveries(), 1);
        assert_eq!(courier.status(), CourierStatus::Free); // No more packages

        // Rating should be 50% = 2.5
        assert_eq!(courier.rating(), 2.5);
    }

    #[test]
    fn test_invalid_email() {
        let result = Courier::builder(
            "John".to_string(),
            "+491234567890".to_string(),
            "invalid-email".to_string(),
            TransportType::Car,
            10.0,
            "Zone".to_string(),
            create_work_hours(),
        )
        .build();

        assert!(matches!(result, Err(CourierError::InvalidEmail(_))));
    }

    #[test]
    fn test_invalid_phone() {
        let result = Courier::builder(
            "John".to_string(),
            "12345".to_string(), // Missing +
            "john@example.com".to_string(),
            TransportType::Car,
            10.0,
            "Zone".to_string(),
            create_work_hours(),
        )
        .build();

        assert!(matches!(result, Err(CourierError::InvalidPhone(_))));
    }

    #[test]
    fn test_work_hours_validation() {
        // Invalid: start >= end
        let result = WorkHours::new(
            chrono::NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            chrono::NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            vec![1],
        );
        assert!(matches!(result, Err(CourierError::InvalidWorkHours(_))));

        // Invalid: empty days
        let result = WorkHours::new(
            chrono::NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            chrono::NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![],
        );
        assert!(matches!(result, Err(CourierError::InvalidWorkHours(_))));

        // Invalid: day out of range
        let result = WorkHours::new(
            chrono::NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            chrono::NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![8], // Invalid day
        );
        assert!(matches!(result, Err(CourierError::InvalidWorkHours(_))));
    }

    #[test]
    fn test_max_load_by_transport_type() {
        let walking = Courier::builder(
            "W".to_string(),
            "+491234567890".to_string(),
            "w@example.com".to_string(),
            TransportType::Walking,
            3.0,
            "Zone".to_string(),
            create_work_hours(),
        )
        .build()
        .unwrap();
        assert_eq!(walking.max_load(), 1);

        let car = Courier::builder(
            "C".to_string(),
            "+491234567891".to_string(),
            "c@example.com".to_string(),
            TransportType::Car,
            50.0,
            "Zone".to_string(),
            create_work_hours(),
        )
        .build()
        .unwrap();
        assert_eq!(car.max_load(), 5);
    }
}
