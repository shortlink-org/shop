//! PostgreSQL Implementation of LocationRepository
//!
//! Uses Sea-ORM for database operations on courier_location_history table.

use async_trait::async_trait;
use chrono::{Duration, Utc};
use sea_orm::{
    ActiveModelTrait, ColumnTrait, DatabaseConnection, EntityTrait, QueryFilter, QueryOrder,
    QuerySelect,
};
use uuid::Uuid;

use crate::domain::model::courier_location::{LocationHistoryEntry, TimeRange};
use crate::domain::ports::{LocationRepository, LocationRepositoryError};
use crate::infrastructure::repository::entities::location_history::{
    ActiveModel, Column, Entity as LocationHistoryEntity,
};

/// PostgreSQL implementation of LocationRepository using Sea-ORM
pub struct LocationPostgresRepository {
    db: DatabaseConnection,
}

impl LocationPostgresRepository {
    /// Create a new repository instance
    pub fn new(db: DatabaseConnection) -> Self {
        Self { db }
    }
}

#[async_trait]
impl LocationRepository for LocationPostgresRepository {
    async fn save(&self, entry: &LocationHistoryEntry) -> Result<(), LocationRepositoryError> {
        let model = ActiveModel::from(entry);

        model
            .insert(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    async fn save_batch(
        &self,
        entries: &[LocationHistoryEntry],
    ) -> Result<(), LocationRepositoryError> {
        if entries.is_empty() {
            return Ok(());
        }

        let models: Vec<ActiveModel> = entries.iter().map(ActiveModel::from).collect();

        LocationHistoryEntity::insert_many(models)
            .exec(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        Ok(())
    }

    async fn get_history(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
        let results = LocationHistoryEntity::find()
            .filter(Column::CourierId.eq(courier_id))
            .filter(Column::Timestamp.gte(time_range.start()))
            .filter(Column::Timestamp.lte(time_range.end()))
            .order_by_asc(Column::Timestamp)
            .all(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        let mut entries = Vec::with_capacity(results.len());
        for model in results {
            let entry = LocationHistoryEntry::try_from(model)
                .map_err(LocationRepositoryError::DataError)?;
            entries.push(entry);
        }

        Ok(entries)
    }

    async fn get_history_paginated(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
        limit: u32,
        offset: u32,
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
        let results = LocationHistoryEntity::find()
            .filter(Column::CourierId.eq(courier_id))
            .filter(Column::Timestamp.gte(time_range.start()))
            .filter(Column::Timestamp.lte(time_range.end()))
            .order_by_desc(Column::Timestamp)
            .limit(limit as u64)
            .offset(offset as u64)
            .all(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        let mut entries = Vec::with_capacity(results.len());
        for model in results {
            let entry = LocationHistoryEntry::try_from(model)
                .map_err(LocationRepositoryError::DataError)?;
            entries.push(entry);
        }

        Ok(entries)
    }

    async fn get_last_location(
        &self,
        courier_id: Uuid,
    ) -> Result<Option<LocationHistoryEntry>, LocationRepositoryError> {
        let result = LocationHistoryEntity::find()
            .filter(Column::CourierId.eq(courier_id))
            .order_by_desc(Column::Timestamp)
            .one(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        match result {
            Some(model) => {
                let entry = LocationHistoryEntry::try_from(model)
                    .map_err(LocationRepositoryError::DataError)?;
                Ok(Some(entry))
            }
            None => Ok(None),
        }
    }

    async fn get_last_locations(
        &self,
        courier_ids: &[Uuid],
    ) -> Result<Vec<LocationHistoryEntry>, LocationRepositoryError> {
        if courier_ids.is_empty() {
            return Ok(Vec::new());
        }

        // Get last location for each courier using a subquery approach
        // For simplicity, we'll query each courier individually
        // In production, you might want to use a window function query
        let mut entries = Vec::with_capacity(courier_ids.len());

        for &courier_id in courier_ids {
            if let Some(entry) = self.get_last_location(courier_id).await? {
                entries.push(entry);
            }
        }

        Ok(entries)
    }

    async fn count_history(
        &self,
        courier_id: Uuid,
        time_range: TimeRange,
    ) -> Result<u64, LocationRepositoryError> {
        let count = LocationHistoryEntity::find()
            .filter(Column::CourierId.eq(courier_id))
            .filter(Column::Timestamp.gte(time_range.start()))
            .filter(Column::Timestamp.lte(time_range.end()))
            .count(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        Ok(count)
    }

    async fn delete_old_history(
        &self,
        older_than_days: u32,
    ) -> Result<u64, LocationRepositoryError> {
        let cutoff = Utc::now() - Duration::days(older_than_days as i64);

        let result = LocationHistoryEntity::delete_many()
            .filter(Column::CreatedAt.lt(cutoff))
            .exec(&self.db)
            .await
            .map_err(|e| LocationRepositoryError::QueryError(e.to_string()))?;

        Ok(result.rows_affected)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::vo::location::Location;

    fn create_test_entry(courier_id: Uuid) -> LocationHistoryEntry {
        let location = Location::new(52.52, 13.405, 10.0).unwrap();
        LocationHistoryEntry::new(courier_id, location, Utc::now(), Some(35.0), Some(180.0))
    }

    #[test]
    fn test_create_entry() {
        let courier_id = Uuid::new_v4();
        let entry = create_test_entry(courier_id);
        assert_eq!(entry.courier_id(), courier_id);
        assert!((entry.latitude() - 52.52).abs() < 0.001);
    }
}
