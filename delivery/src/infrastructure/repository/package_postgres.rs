//! PostgreSQL Implementation of PackageRepository
//!
//! Uses Sea-ORM for database operations.

use async_trait::async_trait;
use sea_orm::{
    ActiveModelTrait, ColumnTrait, DatabaseConnection, EntityTrait, PaginatorTrait, QueryFilter,
    QueryOrder, QuerySelect,
};
use uuid::Uuid;

use crate::boundary::ports::{PackageFilter, PackageRepository, RepositoryError};
use crate::domain::model::package::{Package, PackageId, PackageStatus};
use crate::infrastructure::repository::entities::package::{
    self, ActiveModel, Entity as PackageEntity,
};

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
}

#[async_trait]
impl PackageRepository for PackagePostgresRepository {
    async fn save(&self, package: &Package) -> Result<(), RepositoryError> {
        let model = ActiveModel::from(package);

        // Check if package exists
        let existing = PackageEntity::find_by_id(package.id().0)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        if existing.is_some() {
            // Update with optimistic locking
            let existing_version = existing.as_ref().unwrap().version;
            if existing_version != package.version() as i32 - 1 {
                return Err(RepositoryError::VersionConflict {
                    expected: package.version() - 1,
                    actual: existing_version as u32,
                });
            }

            model
                .update(&self.db)
                .await
                .map_err(|e| RepositoryError::QueryError(e.to_string()))?;
        } else {
            model
                .insert(&self.db)
                .await
                .map_err(|e| RepositoryError::QueryError(e.to_string()))?;
        }

        Ok(())
    }

    async fn find_by_id(&self, id: PackageId) -> Result<Option<Package>, RepositoryError> {
        let result = PackageEntity::find_by_id(id.0)
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let package = Package::try_from(model)
                    .map_err(|e| RepositoryError::SerializationError(e))?;
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
                let package = Package::try_from(model)
                    .map_err(|e| RepositoryError::SerializationError(e))?;
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
            let status_strings: Vec<String> = statuses.iter().map(|s| Self::status_to_string(*s)).collect();
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
            let package = Package::try_from(model)
                .map_err(|e| RepositoryError::SerializationError(e))?;
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
            let package = Package::try_from(model)
                .map_err(|e| RepositoryError::SerializationError(e))?;
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
