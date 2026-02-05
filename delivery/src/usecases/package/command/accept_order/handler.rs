//! Accept Order Handler
//!
//! Handles accepting an order from OMS for delivery.
//!
//! ## Flow
//! 1. Validate command data
//! 2. Check for duplicate order
//! 3. Create Package aggregate
//! 4. Move to IN_POOL status
//! 5. Save to repository
//! 6. Return package ID

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::package::{Package, PackageStatus};
use crate::domain::ports::{CommandHandlerWithResult, PackageRepository, RepositoryError};

use super::Command;

/// Errors that can occur during order acceptance
#[derive(Debug, Error)]
pub enum AcceptOrderError {
    /// Order already exists
    #[error("Order already exists: {0}")]
    DuplicateOrder(Uuid),

    /// Invalid request data
    #[error("Invalid request: {0}")]
    InvalidRequest(String),

    /// Invalid address
    #[error("Invalid address: {0}")]
    InvalidAddress(String),

    /// Invalid delivery period
    #[error("Invalid delivery period: {0}")]
    InvalidDeliveryPeriod(String),

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from accepting an order
#[derive(Debug, Clone)]
pub struct Response {
    /// The created package ID
    pub package_id: Uuid,
    /// The order ID from OMS
    pub order_id: Uuid,
    /// Package status
    pub status: PackageStatus,
    /// Creation timestamp
    pub created_at: DateTime<Utc>,
}

/// Accept Order Handler
pub struct Handler<R>
where
    R: PackageRepository,
{
    package_repo: Arc<R>,
}

impl<R> Handler<R>
where
    R: PackageRepository,
{
    /// Create a new handler instance
    pub fn new(package_repo: Arc<R>) -> Self {
        Self { package_repo }
    }

    /// Derive work zone from delivery address (simplified)
    fn derive_zone(city: &str, postal_code: &str) -> String {
        // Simple zone derivation: City-PostalCodePrefix
        // In real implementation, this would use a proper zone mapping service
        let postal_prefix = if postal_code.len() >= 3 {
            &postal_code[..3]
        } else {
            postal_code
        };
        format!("{}-{}", city, postal_prefix)
    }
}

impl<R> CommandHandlerWithResult<Command, Response> for Handler<R>
where
    R: PackageRepository + Send + Sync,
{
    type Error = AcceptOrderError;

    /// Handle the AcceptOrder command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Validate command
        cmd.validate().map_err(AcceptOrderError::InvalidRequest)?;

        // 2. Check for duplicate order
        if let Some(_existing) = self
            .package_repo
            .find_by_order_id(cmd.order_id)
            .await?
        {
            return Err(AcceptOrderError::DuplicateOrder(cmd.order_id));
        }

        // 3. Convert inputs to domain objects
        let pickup_address = cmd
            .pickup_address
            .to_domain()
            .map_err(AcceptOrderError::InvalidAddress)?;

        let delivery_address = cmd
            .delivery_address
            .to_domain()
            .map_err(AcceptOrderError::InvalidAddress)?;

        let delivery_period = cmd
            .delivery_period
            .to_domain()
            .map_err(AcceptOrderError::InvalidDeliveryPeriod)?;

        // 4. Derive zone from delivery address
        let zone = Self::derive_zone(&cmd.delivery_address.city, &cmd.delivery_address.postal_code);

        // 5. Create Package aggregate with ACCEPTED status
        let mut package = Package::new(
            cmd.order_id,
            cmd.customer_id,
            cmd.customer_phone.clone(),
            cmd.recipient_name.clone(),
            cmd.recipient_phone.clone(),
            cmd.recipient_email.clone(),
            pickup_address,
            delivery_address,
            delivery_period,
            cmd.weight_kg,
            cmd.priority,
            zone,
        );

        // 6. Transition to IN_POOL status (ready for assignment)
        package.move_to_pool().map_err(|e| {
            AcceptOrderError::InvalidRequest(format!("Failed to move to pool: {}", e))
        })?;

        let package_id = *package.id();
        let created_at = package.created_at();
        let status = package.status();

        // 7. Save to repository
        self.package_repo.save(&package).await?;

        // 8. TODO: Publish PackageAccepted event
        // This should be done via event publisher when implemented

        Ok(Response {
            package_id: package_id.0,
            order_id: cmd.order_id,
            status,
            created_at,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::package::{PackageId, Priority};
    use crate::domain::ports::PackageFilter;
    use crate::usecases::package::command::accept_order::command::{
        AddressInput, DeliveryPeriodInput,
    };
    use async_trait::async_trait;
    use chrono::Utc;
    use std::collections::HashMap;
    use std::sync::Mutex;

    /// Mock PackageRepository for testing
    struct MockPackageRepository {
        packages: Mutex<HashMap<Uuid, Package>>,
        order_packages: Mutex<HashMap<Uuid, Uuid>>, // order_id -> package_id
    }

    impl MockPackageRepository {
        fn new() -> Self {
            Self {
                packages: Mutex::new(HashMap::new()),
                order_packages: Mutex::new(HashMap::new()),
            }
        }
    }

    #[async_trait]
    impl PackageRepository for MockPackageRepository {
        async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
            let mut packages = self.packages.lock().unwrap();
            let mut order_packages = self.order_packages.lock().unwrap();
            packages.insert(package.id().0, package.clone());
            order_packages.insert(package.order_id(), package.id().0);
            Ok(())
        }

        async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
            let packages = self.packages.lock().unwrap();
            Ok(packages.get(&id.0).cloned())
        }

        async fn find_by_order_id(&self, order_id: Uuid) -> Result<Option<Package>, RepositoryError> {
            let order_packages = self.order_packages.lock().unwrap();
            let packages = self.packages.lock().unwrap();
            if let Some(package_id) = order_packages.get(&order_id) {
                Ok(packages.get(package_id).cloned())
            } else {
                Ok(None)
            }
        }

        async fn find_by_filter(
            &self,
            _filter: PackageFilter,
            _limit: u64,
            _offset: u64,
        ) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn count_by_filter(&self, _filter: PackageFilter) -> Result<u64, RepositoryError> {
            Ok(0)
        }

        async fn find_by_courier(&self, _courier_id: Uuid) -> Result<Vec<Package>, RepositoryError> {
            Ok(vec![])
        }

        async fn delete(&self, _id: PackageId) -> Result<(), RepositoryError> {
            Ok(())
        }
    }

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

    fn create_valid_command() -> Command {
        Command::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_valid_address(),
            create_valid_address(),
            create_valid_period(),
            2.5,
            Priority::Normal,
        )
    }

    #[tokio::test]
    async fn test_accept_order_success() {
        let repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(repo.clone());

        let cmd = create_valid_command();
        let order_id = cmd.order_id;

        let result = handler.handle(cmd).await;
        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());

        let response = result.unwrap();
        assert_eq!(response.order_id, order_id);
        assert_eq!(response.status, PackageStatus::InPool);
    }

    #[tokio::test]
    async fn test_accept_order_duplicate() {
        let repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(repo.clone());

        let cmd = create_valid_command();
        let order_id = cmd.order_id;

        // First accept should succeed
        let result = handler.handle(cmd.clone()).await;
        assert!(result.is_ok());

        // Second accept with same order_id should fail
        let duplicate_cmd = Command::new(
            order_id, // Same order_id
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_valid_address(),
            create_valid_address(),
            create_valid_period(),
            2.5,
            Priority::Normal,
        );

        let result = handler.handle(duplicate_cmd).await;
        assert!(matches!(result, Err(AcceptOrderError::DuplicateOrder(_))));
    }

    #[tokio::test]
    async fn test_accept_order_invalid_weight() {
        let repo = Arc::new(MockPackageRepository::new());
        let handler = Handler::new(repo);

        let cmd = Command::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None, // customer_phone
            None,
            None,
            None,
            create_valid_address(),
            create_valid_address(),
            create_valid_period(),
            0.0, // Invalid weight
            Priority::Normal,
        );

        let result = handler.handle(cmd).await;
        assert!(matches!(result, Err(AcceptOrderError::InvalidRequest(_))));
    }
}
