//! Redis Implementation of CourierCache
//!
//! Uses Redis for caching courier state (hot data).
//!
//! Redis Key Schema:
//! - `courier:{id}:state` - HASH containing status, current_load, max_load, rating, etc.
//! - `couriers:free` - SET of all free courier IDs
//! - `couriers:zone:{zone}:free` - SET of free courier IDs in a zone

use async_trait::async_trait;
use redis::aio::ConnectionManager;
use redis::AsyncCommands;
use uuid::Uuid;

use crate::boundary::ports::{CacheError, CachedCourierState, CourierCache};
use crate::domain::model::courier::CourierStatus;

/// Redis key prefix for courier state
const COURIER_STATE_PREFIX: &str = "courier";
/// Redis key for all free couriers set
const FREE_COURIERS_KEY: &str = "couriers:free";
/// Redis key prefix for zone-based free couriers
const ZONE_FREE_PREFIX: &str = "couriers:zone";

/// Redis implementation of CourierCache
pub struct CourierRedisCache {
    conn: ConnectionManager,
}

impl CourierRedisCache {
    /// Create a new Redis cache instance
    pub fn new(conn: ConnectionManager) -> Self {
        Self { conn }
    }

    /// Get the state key for a courier
    fn state_key(courier_id: Uuid) -> String {
        format!("{}:{}:state", COURIER_STATE_PREFIX, courier_id)
    }

    /// Get the zone-specific free couriers key
    fn zone_free_key(zone: &str) -> String {
        format!("{}:{}:free", ZONE_FREE_PREFIX, zone)
    }

    /// Convert CourierStatus to string
    fn status_to_string(status: CourierStatus) -> &'static str {
        match status {
            CourierStatus::Unavailable => "unavailable",
            CourierStatus::Free => "free",
            CourierStatus::Busy => "busy",
        }
    }

    /// Convert string to CourierStatus
    fn string_to_status(s: &str) -> Option<CourierStatus> {
        match s.to_lowercase().as_str() {
            "unavailable" => Some(CourierStatus::Unavailable),
            "free" => Some(CourierStatus::Free),
            "busy" => Some(CourierStatus::Busy),
            _ => None,
        }
    }
}

#[async_trait]
impl CourierCache for CourierRedisCache {
    async fn initialize_state(
        &self,
        courier_id: Uuid,
        state: CachedCourierState,
        work_zone: &str,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        // Store state as hash
        let _: () = redis::pipe()
            .hset(&key, "status", Self::status_to_string(state.status))
            .hset(&key, "current_load", state.current_load)
            .hset(&key, "max_load", state.max_load)
            .hset(&key, "rating", state.rating.to_string())
            .hset(&key, "successful_deliveries", state.successful_deliveries)
            .hset(&key, "failed_deliveries", state.failed_deliveries)
            .hset(&key, "work_zone", work_zone)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        // Add to free sets if status is Free
        if state.status == CourierStatus::Free {
            let _: () = redis::pipe()
                .sadd(FREE_COURIERS_KEY, courier_id.to_string())
                .sadd(Self::zone_free_key(work_zone), courier_id.to_string())
                .query_async(&mut conn)
                .await
                .map_err(|e| CacheError::OperationError(e.to_string()))?;
        }

        Ok(())
    }

    async fn get_state(&self, courier_id: Uuid) -> Result<Option<CachedCourierState>, CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let result: Option<Vec<(String, String)>> = conn
            .hgetall(&key)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        match result {
            Some(fields) if !fields.is_empty() => {
                let mut status = CourierStatus::Unavailable;
                let mut current_load = 0u32;
                let mut max_load = 0u32;
                let mut rating = 0.0f64;
                let mut successful_deliveries = 0u32;
                let mut failed_deliveries = 0u32;

                for (field, value) in fields {
                    match field.as_str() {
                        "status" => {
                            status = Self::string_to_status(&value)
                                .unwrap_or(CourierStatus::Unavailable);
                        }
                        "current_load" => {
                            current_load = value.parse().unwrap_or(0);
                        }
                        "max_load" => {
                            max_load = value.parse().unwrap_or(0);
                        }
                        "rating" => {
                            rating = value.parse().unwrap_or(0.0);
                        }
                        "successful_deliveries" => {
                            successful_deliveries = value.parse().unwrap_or(0);
                        }
                        "failed_deliveries" => {
                            failed_deliveries = value.parse().unwrap_or(0);
                        }
                        _ => {}
                    }
                }

                Ok(Some(CachedCourierState {
                    status,
                    current_load,
                    max_load,
                    rating,
                    successful_deliveries,
                    failed_deliveries,
                }))
            }
            _ => Ok(None),
        }
    }

    async fn set_status(
        &self,
        courier_id: Uuid,
        status: CourierStatus,
        work_zone: &str,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);
        let courier_id_str = courier_id.to_string();
        let zone_key = Self::zone_free_key(work_zone);

        // Update status in hash
        let _: () = conn
            .hset(&key, "status", Self::status_to_string(status))
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        // Update free sets
        match status {
            CourierStatus::Free => {
                // Add to free sets
                let _: () = redis::pipe()
                    .sadd(FREE_COURIERS_KEY, &courier_id_str)
                    .sadd(&zone_key, &courier_id_str)
                    .query_async(&mut conn)
                    .await
                    .map_err(|e| CacheError::OperationError(e.to_string()))?;
            }
            _ => {
                // Remove from free sets
                let _: () = redis::pipe()
                    .srem(FREE_COURIERS_KEY, &courier_id_str)
                    .srem(&zone_key, &courier_id_str)
                    .query_async(&mut conn)
                    .await
                    .map_err(|e| CacheError::OperationError(e.to_string()))?;
            }
        }

        Ok(())
    }

    async fn get_status(&self, courier_id: Uuid) -> Result<Option<CourierStatus>, CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let result: Option<String> = conn
            .hget(&key, "status")
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(result.and_then(|s| Self::string_to_status(&s)))
    }

    async fn update_load(
        &self,
        courier_id: Uuid,
        current_load: u32,
        max_load: u32,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let _: () = redis::pipe()
            .hset(&key, "current_load", current_load)
            .hset(&key, "max_load", max_load)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn update_stats(
        &self,
        courier_id: Uuid,
        rating: f64,
        successful_deliveries: u32,
        failed_deliveries: u32,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let _: () = redis::pipe()
            .hset(&key, "rating", rating.to_string())
            .hset(&key, "successful_deliveries", successful_deliveries)
            .hset(&key, "failed_deliveries", failed_deliveries)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn get_free_couriers_in_zone(&self, zone: &str) -> Result<Vec<Uuid>, CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::zone_free_key(zone);

        let members: Vec<String> = conn
            .smembers(&key)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        let mut uuids = Vec::with_capacity(members.len());
        for member in members {
            if let Ok(uuid) = Uuid::parse_str(&member) {
                uuids.push(uuid);
            }
        }

        Ok(uuids)
    }

    async fn get_all_free_couriers(&self) -> Result<Vec<Uuid>, CacheError> {
        let mut conn = self.conn.clone();

        let members: Vec<String> = conn
            .smembers(FREE_COURIERS_KEY)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        let mut uuids = Vec::with_capacity(members.len());
        for member in members {
            if let Ok(uuid) = Uuid::parse_str(&member) {
                uuids.push(uuid);
            }
        }

        Ok(uuids)
    }

    async fn remove(&self, courier_id: Uuid, work_zone: &str) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);
        let courier_id_str = courier_id.to_string();
        let zone_key = Self::zone_free_key(work_zone);

        let _: () = redis::pipe()
            .del(&key)
            .srem(FREE_COURIERS_KEY, &courier_id_str)
            .srem(&zone_key, &courier_id_str)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn exists(&self, courier_id: Uuid) -> Result<bool, CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let result: bool = conn
            .exists(&key)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(result)
    }
}
