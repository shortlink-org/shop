//! PostgreSQL Implementation of PackageRepository
//!
//! Uses Sea-ORM for database operations.

use async_trait::async_trait;
use sea_orm::{
    ActiveModelTrait, ColumnTrait, ConnectionTrait, DatabaseConnection, EntityTrait,
    PaginatorTrait, QueryFilter, QueryOrder, QuerySelect, TransactionTrait,
};
use uuid::Uuid;

use crate::domain::model::{
    courier::Courier,
    package::{Package, PackageId, PackageStatus},
};
use crate::domain::ports::{DomainEvent, PackageFilter, PackageRepository, RepositoryError};
use crate::infrastructure::repository::entities::package::{
    self, ActiveModel, Entity as PackageEntity,
};
use crate::infrastructure::repository::{CourierPostgresRepository, OutboxPostgresRepository};

/// PostgreSQL implementation of PackageRepository using Sea-ORM
pub struct PackagePostgresRepository {
    db: DatabaseConnection,
}

impl PackagePostgresRepository {
    /// Create a new repository instance
    pub fn new(db: DatabaseConnection) -> Self {
        Self { db }
    }

    /// Convert PackageStatus to database string
    fn status_to_string(status: PackageStatus) -> String {
        match status {
            PackageStatus::Accepted => "accepted".to_string(),
            PackageStatus::InPool => "in_pool".to_string(),
            PackageStatus::Assigned => "assigned".to_string(),
            PackageStatus::InTransit => "in_transit".to_string(),
            PackageStatus::Delivered => "delivered".to_string(),
            PackageStatus::NotDelivered => "not_delivered".to_string(),
            PackageStatus::RequiresHandling => "requires_handling".to_string(),
        }
    }

    async fn save_with_conn<C>(&self, conn: &C, package: &Package) -> Result<(), RepositoryError>
    where
        C: ConnectionTrait,
    {
        let model = ActiveModel::from(package);

        let existing = PackageEntity::find_by_id(package.id().0)
            .one(conn)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        if let Some(ref existing_model) = existing {
            let existing_version = existing_model.version;
            if existing_version != package.version() as i32 - 1 {
                return Err(RepositoryError::VersionConflict {
                    expected: package.version() - 1,
                    actual: existing_version as u32,
                });
            }

            model
                .update(conn)
                .await
                .map_err(|e| RepositoryError::QueryError(e.to_string()))?;
        } else {
            model
                .insert(conn)
                .await
                .map_err(|e| RepositoryError::QueryError(e.to_string()))?;
        }

        Ok(())
    }
}

#[async_trait]
impl PackageRepository for PackagePostgresRepository {
    async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
        self.save_with_conn(&self.db, package).await
    }

    async fn save_with_events(
        &self,
        package: &Package,
        events: &[DomainEvent],
    ) -> Result<(), RepositoryError> {
        let tx = self
            .db
            .begin()
            .await
            .map_err(|e| RepositoryError::ConnectionError(e.to_string()))?;

        self.save_with_conn(&tx, package).await?;
        OutboxPostgresRepository::insert_events_in_tx(&tx, events).await?;

        tx.commit()
            .await
            .map_err(|e| RepositoryError::ConnectionError(e.to_string()))?;

        Ok(())
    }

    async fn save_courier_with_package_and_events(
        &self,
        courier: &Courier,
        package: &Package,
        events: &[DomainEvent],
    ) -> Result<(), RepositoryError> {
        let tx = self
            .db
            .begin()
            .await
            .map_err(|e| RepositoryError::ConnectionError(e.to_string()))?;

        self.save_with_conn(&tx, package).await?;
        CourierPostgresRepository::save_with_conn(&tx, courier).await?;
        OutboxPostgresRepository::insert_events_in_tx(&tx, events).await?;

        tx.commit()
            .await
            .map_err(|e| RepositoryError::ConnectionError(e.to_string()))?;

        Ok(())
    }

    async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
        let result = PackageEntity::find_by_id(id.0)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let package =
                    Package::try_from(model).map_err(RepositoryError::SerializationError)?;
                Ok(Some(package))
            }
            None => Ok(None),
        }
    }

    async fn find_by_order_id(&self, order_id: Uuid) -> Result<Option<Package>, RepositoryError> {
        let result = PackageEntity::find()
            .filter(package::Column::OrderId.eq(order_id))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let package =
                    Package::try_from(model).map_err(RepositoryError::SerializationError)?;
                Ok(Some(package))
            }
            None => Ok(None),
        }
    }

    async fn find_by_filter(
        &self,
        filter: PackageFilter,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<Package>, RepositoryError> {
        let mut query = PackageEntity::find();

        // Apply filters
        if let Some(status) = filter.status {
            query = query.filter(package::Column::Status.eq(Self::status_to_string(status)));
        }

        if let Some(statuses) = filter.statuses {
            let status_strings: Vec<String> = statuses
                .iter()
                .map(|s| Self::status_to_string(*s))
                .collect();
            query = query.filter(package::Column::Status.is_in(status_strings));
        }

        if let Some(zone) = filter.zone {
            query = query.filter(package::Column::Zone.eq(zone));
        }

        if let Some(courier_id) = filter.courier_id {
            query = query.filter(package::Column::CourierId.eq(courier_id));
        }

        if filter.unassigned_only {
            query = query.filter(package::Column::CourierId.is_null());
        }

        // Apply ordering and pagination
        let results = query
            .order_by_asc(package::Column::Priority)
            .order_by_asc(package::Column::CreatedAt)
            .limit(limit)
            .offset(offset)
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        let mut packages = Vec::with_capacity(results.len());
        for model in results {
            let package = Package::try_from(model).map_err(RepositoryError::SerializationError)?;
            packages.push(package);
        }

        Ok(packages)
    }

    async fn count_by_filter(&self, filter: PackageFilter) -> Result<u64, RepositoryError> {
        let mut query = PackageEntity::find();

        if let Some(status) = filter.status {
            query = query.filter(package::Column::Status.eq(Self::status_to_string(status)));
        }

        if let Some(zone) = filter.zone {
            query = query.filter(package::Column::Zone.eq(zone));
        }

        if let Some(courier_id) = filter.courier_id {
            query = query.filter(package::Column::CourierId.eq(courier_id));
        }

        if filter.unassigned_only {
            query = query.filter(package::Column::CourierId.is_null());
        }

        let count = query
            .count(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        Ok(count)
    }

    async fn find_by_courier(&self, courier_id: Uuid) -> Result<Vec<Package>, RepositoryError> {
        let results = PackageEntity::find()
            .filter(package::Column::CourierId.eq(courier_id))
            .order_by_asc(package::Column::CreatedAt)
            .all(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        let mut packages = Vec::with_capacity(results.len());
        for model in results {
            let package = Package::try_from(model).map_err(RepositoryError::SerializationError)?;
            packages.push(package);
        }

        Ok(packages)
    }

    async fn delete(&self, id: PackageId) -> Result<(), RepositoryError> {
        let result = PackageEntity::delete_by_id(id.0)
            .exec(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        if result.rows_affected == 0 {
            return Err(RepositoryError::NotFound(id.0));
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::domain::delivery::common::v1 as proto_common;
    use crate::domain::model::domain::delivery::events::v1::{
        PackageAcceptedEvent, PackageAssignedEvent,
    };
    use crate::domain::model::package::{Address, DeliveryPeriod, Package, Priority};
    use crate::domain::model::vo::location::Location;
    use crate::infrastructure::repository::entities::outbox_message::Entity as OutboxEntity;
    use migration::{Migrator, MigratorTrait};
    use sea_orm::Database;
    use testcontainers::{runners::AsyncRunner, ImageExt};
    use testcontainers_modules::postgres::Postgres;

    use crate::test_support::ManagedAsyncContainer;

    async fn setup_db() -> (ManagedAsyncContainer<Postgres>, DatabaseConnection) {
        let container = ManagedAsyncContainer::new(
            Postgres::default()
                .with_tag("18-alpine")
                .start()
                .await
                .unwrap(),
        );
        let port = container.get_host_port_ipv4(5432).await.unwrap();
        let url = format!("postgres://postgres:postgres@localhost:{}/postgres", port);
        let db = Database::connect(&url).await.unwrap();

        Migrator::up(&db, None).await.unwrap();

        (container, db)
    }

    fn create_test_address() -> Address {
        Address::new(
            "123 Main St".to_string(),
            "Berlin".to_string(),
            Location::new(52.52, 13.405, 10.0).unwrap(),
        )
    }

    fn create_package_in_pool() -> Package {
        let now = chrono::Utc::now();
        let period = DeliveryPeriod::new(
            now + chrono::Duration::hours(2),
            now + chrono::Duration::hours(4),
        )
        .unwrap();

        let mut package = Package::new(
            Uuid::new_v4(),
            Uuid::new_v4(),
            None,
            None,
            None,
            None,
            create_test_address(),
            create_test_address(),
            period,
            2.5,
            Priority::Normal,
            "Berlin-101".to_string(),
        );
        package.move_to_pool().unwrap();
        package
    }

    #[tokio::test]
    async fn test_save_with_events_persists_package_and_outbox_row() {
        let (_container, db) = setup_db().await;
        let repo = PackagePostgresRepository::new(db.clone());
        let package = create_package_in_pool();

        let now = chrono::Utc::now();
        let events = vec![DomainEvent::PackageAccepted(PackageAcceptedEvent {
            package_id: package.id().0.to_string(),
            order_id: package.order_id().to_string(),
            customer_id: package.customer_id().to_string(),
            status: proto_common::PackageStatus::InPool as i32,
            created_at: Some(pbjson_types::Timestamp {
                seconds: package.created_at().timestamp(),
                nanos: package.created_at().timestamp_subsec_nanos() as i32,
            }),
            occurred_at: Some(pbjson_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
        })];

        repo.save_with_events(&package, &events).await.unwrap();

        let persisted = repo.find_by_id(*package.id()).await.unwrap();
        assert!(persisted.is_some());

        let outbox_rows = OutboxEntity::find().all(&db).await.unwrap();
        assert_eq!(outbox_rows.len(), 1);
        assert_eq!(outbox_rows[0].topic, "delivery.package.status.v1");
        assert_eq!(outbox_rows[0].event_type, "PackageAcceptedEvent");
        assert_eq!(outbox_rows[0].aggregate_id, package.id().0.to_string());
    }

    #[tokio::test]
    async fn test_save_with_events_rolls_back_when_event_encoding_fails() {
        let (_container, db) = setup_db().await;
        let repo = PackagePostgresRepository::new(db.clone());
        let mut package = create_package_in_pool();
        let courier_id = Uuid::new_v4();
        package.assign_to(courier_id).unwrap();

        let events = vec![DomainEvent::PackageAssigned(PackageAssignedEvent {
            package_id: package.id().0.to_string(),
            courier_id: courier_id.to_string(),
            status: proto_common::PackageStatus::Assigned as i32,
            assigned_at: None,
            pickup_address: None,
            delivery_address: None,
            delivery_period: None,
            customer_phone: String::new(),
            occurred_at: None,
        })];

        let err = repo.save_with_events(&package, &events).await.unwrap_err();
        assert!(matches!(err, RepositoryError::SerializationError(_)));

        let persisted = repo.find_by_id(*package.id()).await.unwrap();
        assert!(persisted.is_none());

        let outbox_rows = OutboxEntity::find().all(&db).await.unwrap();
        assert!(outbox_rows.is_empty());
    }
}
