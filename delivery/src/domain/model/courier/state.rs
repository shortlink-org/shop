use std::fmt;

/// Courier status representing availability
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CourierStatus {
    /// Courier is unavailable (initial status, off-duty)
    Unavailable,
    /// Courier is free and ready to accept orders
    Free,
    /// Courier is busy with deliveries
    Busy,
}

impl fmt::Display for CourierStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CourierStatus::Unavailable => write!(f, "Unavailable"),
            CourierStatus::Free => write!(f, "Free"),
            CourierStatus::Busy => write!(f, "Busy"),
        }
    }
}

/// Courier capacity management
#[derive(Debug, Clone, PartialEq)]
pub struct CourierCapacity {
    /// Current number of packages
    current_load: u32,
    /// Maximum number of packages
    max_load: u32,
}

#[derive(Debug, Clone, PartialEq)]
pub enum CapacityError {
    /// Courier is at full capacity
    AtFullCapacity,
    /// No packages to release
    NoPackagesToRelease,
}

impl fmt::Display for CapacityError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CapacityError::AtFullCapacity => write!(f, "Courier is at full capacity"),
            CapacityError::NoPackagesToRelease => write!(f, "No packages to release"),
        }
    }
}

impl std::error::Error for CapacityError {}

impl CourierCapacity {
    /// Create a new courier capacity
    pub fn new(max_load: u32) -> Self {
        Self {
            current_load: 0,
            max_load,
        }
    }

    /// Get current load
    pub fn current_load(&self) -> u32 {
        self.current_load
    }

    /// Get maximum load
    pub fn max_load(&self) -> u32 {
        self.max_load
    }

    /// Check if courier can accept more packages
    pub fn can_accept(&self) -> bool {
        self.current_load < self.max_load
    }

    /// Get available capacity
    pub fn available_capacity(&self) -> u32 {
        self.max_load - self.current_load
    }

    /// Add a package to courier's load
    pub fn add_package(&mut self) -> Result<(), CapacityError> {
        if !self.can_accept() {
            return Err(CapacityError::AtFullCapacity);
        }
        self.current_load += 1;
        Ok(())
    }

    /// Release a package from courier's load
    pub fn release_package(&mut self) -> Result<(), CapacityError> {
        if self.current_load == 0 {
            return Err(CapacityError::NoPackagesToRelease);
        }
        self.current_load -= 1;
        Ok(())
    }

    /// Check if courier is empty (no packages)
    pub fn is_empty(&self) -> bool {
        self.current_load == 0
    }
}

impl CourierStatus {
    /// Check if courier can accept new assignments
    pub fn can_accept_assignment(&self) -> bool {
        matches!(self, CourierStatus::Free)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_courier_status_can_accept() {
        assert!(CourierStatus::Free.can_accept_assignment());
        assert!(!CourierStatus::Unavailable.can_accept_assignment());
        assert!(!CourierStatus::Busy.can_accept_assignment());
    }

    #[test]
    fn test_capacity_new() {
        let capacity = CourierCapacity::new(5);
        assert_eq!(capacity.current_load(), 0);
        assert_eq!(capacity.max_load(), 5);
        assert!(capacity.can_accept());
        assert!(capacity.is_empty());
    }

    #[test]
    fn test_capacity_add_package() {
        let mut capacity = CourierCapacity::new(2);
        
        assert!(capacity.add_package().is_ok());
        assert_eq!(capacity.current_load(), 1);
        
        assert!(capacity.add_package().is_ok());
        assert_eq!(capacity.current_load(), 2);
        assert!(!capacity.can_accept());
        
        assert!(matches!(capacity.add_package(), Err(CapacityError::AtFullCapacity)));
    }

    #[test]
    fn test_capacity_release_package() {
        let mut capacity = CourierCapacity::new(2);
        capacity.add_package().unwrap();
        
        assert!(capacity.release_package().is_ok());
        assert_eq!(capacity.current_load(), 0);
        
        assert!(matches!(capacity.release_package(), Err(CapacityError::NoPackagesToRelease)));
    }

    #[test]
    fn test_available_capacity() {
        let mut capacity = CourierCapacity::new(5);
        assert_eq!(capacity.available_capacity(), 5);
        
        capacity.add_package().unwrap();
        assert_eq!(capacity.available_capacity(), 4);
    }
}
