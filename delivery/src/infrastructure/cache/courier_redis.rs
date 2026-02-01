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

use crate::domain::ports::{CacheError, CachedCourierState, CourierCache};
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
            CourierStatus::Archived => "archived",
        }
    }

    /// Convert string to CourierStatus
    fn string_to_status(s: &str) -> Option<CourierStatus> {
        match s.to_lowercase().as_str() {
            "unavailable" => Some(CourierStatus::Unavailable),
            "free" => Some(CourierStatus::Free),
            "busy" => Some(CourierStatus::Busy),
            "archived" => Some(CourierStatus::Archived),
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

    async fn update_status(
        &self,
        courier_id: Uuid,
        status: CourierStatus,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let _: () = conn
            .hset(&key, "status", Self::status_to_string(status))
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn update_max_load(&self, courier_id: Uuid, max_load: u32) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let key = Self::state_key(courier_id);

        let _: () = conn
            .hset(&key, "max_load", max_load)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn add_to_free_pool(&self, courier_id: Uuid, work_zone: &str) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let courier_id_str = courier_id.to_string();
        let zone_key = Self::zone_free_key(work_zone);

        let _: () = redis::pipe()
            .sadd(FREE_COURIERS_KEY, &courier_id_str)
            .sadd(&zone_key, &courier_id_str)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn remove_from_free_pool(
        &self,
        courier_id: Uuid,
        work_zone: &str,
    ) -> Result<(), CacheError> {
        let mut conn = self.conn.clone();
        let courier_id_str = courier_id.to_string();
        let zone_key = Self::zone_free_key(work_zone);

        let _: () = redis::pipe()
            .srem(FREE_COURIERS_KEY, &courier_id_str)
            .srem(&zone_key, &courier_id_str)
            .query_async(&mut conn)
            .await
            .map_err(|e| CacheError::OperationError(e.to_string()))?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use testcontainers::{runners::AsyncRunner, ContainerAsync, ImageExt};
    use testcontainers_modules::redis::Redis;

    async fn setup_redis() -> (ContainerAsync<Redis>, CourierRedisCache) {
        let container = Redis::default().with_tag("7-alpine").start().await.unwrap();
        let port = container.get_host_port_ipv4(6379).await.unwrap();
        let url = format!("redis://localhost:{}", port);

        let client = redis::Client::open(url).unwrap();
        let conn = redis::aio::ConnectionManager::new(client).await.unwrap();
        let cache = CourierRedisCache::new(conn);

        (container, cache)
    }

    fn test_state() -> CachedCourierState {
        CachedCourierState {
            status: CourierStatus::Unavailable,
            current_load: 0,
            max_load: 2,
            rating: 0.0,
            successful_deliveries: 0,
            failed_deliveries: 0,
        }
    }

    // ==================== Status Conversion Tests ====================

    #[test]
    fn test_status_to_string_all_variants() {
        assert_eq!(
            CourierRedisCache::status_to_string(CourierStatus::Unavailable),
            "unavailable"
        );
        assert_eq!(
            CourierRedisCache::status_to_string(CourierStatus::Free),
            "free"
        );
        assert_eq!(
            CourierRedisCache::status_to_string(CourierStatus::Busy),
            "busy"
        );
        assert_eq!(
            CourierRedisCache::status_to_string(CourierStatus::Archived),
            "archived"
        );
    }

    #[test]
    fn test_string_to_status_all_variants() {
        assert_eq!(
            CourierRedisCache::string_to_status("unavailable"),
            Some(CourierStatus::Unavailable)
        );
        assert_eq!(
            CourierRedisCache::string_to_status("free"),
            Some(CourierStatus::Free)
        );
        assert_eq!(
            CourierRedisCache::string_to_status("busy"),
            Some(CourierStatus::Busy)
        );
        assert_eq!(
            CourierRedisCache::string_to_status("archived"),
            Some(CourierStatus::Archived)
        );
    }

    #[test]
    fn test_string_to_status_case_insensitive() {
        assert_eq!(
            CourierRedisCache::string_to_status("FREE"),
            Some(CourierStatus::Free)
        );
        assert_eq!(
            CourierRedisCache::string_to_status("Free"),
            Some(CourierStatus::Free)
        );
    }

    #[test]
    fn test_string_to_status_unknown_returns_none() {
        assert_eq!(CourierRedisCache::string_to_status("unknown"), None);
        assert_eq!(CourierRedisCache::string_to_status(""), None);
    }

    // ==================== Key Generation Tests ====================

    #[test]
    fn test_state_key_format() {
        let id = Uuid::parse_str("123e4567-e89b-12d3-a456-426614174000").unwrap();
        let key = CourierRedisCache::state_key(id);
        assert_eq!(key, "courier:123e4567-e89b-12d3-a456-426614174000:state");
    }

    #[test]
    fn test_zone_free_key_format() {
        let key = CourierRedisCache::zone_free_key("Berlin-Mitte");
        assert_eq!(key, "couriers:zone:Berlin-Mitte:free");
    }

    // ==================== Integration Tests with Redis Container ====================

    #[tokio::test]
    async fn test_initialize_and_get_state() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let state = test_state();

        cache
            .initialize_state(courier_id, state.clone(), "Berlin-Mitte")
            .await
            .unwrap();

        let retrieved = cache.get_state(courier_id).await.unwrap();
        assert!(retrieved.is_some());
        let retrieved = retrieved.unwrap();
        assert_eq!(retrieved.status, CourierStatus::Unavailable);
        assert_eq!(retrieved.max_load, 2);
    }

    #[tokio::test]
    async fn test_get_state_not_found() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        let result = cache.get_state(courier_id).await.unwrap();
        assert!(result.is_none());
    }

    #[tokio::test]
    async fn test_update_status() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let state = test_state();

        cache
            .initialize_state(courier_id, state, "Berlin-Mitte")
            .await
            .unwrap();

        // Update status to Free
        cache
            .update_status(courier_id, CourierStatus::Free)
            .await
            .unwrap();

        let retrieved = cache.get_state(courier_id).await.unwrap().unwrap();
        assert_eq!(retrieved.status, CourierStatus::Free);
    }

    #[tokio::test]
    async fn test_update_status_to_archived() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let state = test_state();

        cache
            .initialize_state(courier_id, state, "Berlin-Mitte")
            .await
            .unwrap();

        cache
            .update_status(courier_id, CourierStatus::Archived)
            .await
            .unwrap();

        let retrieved = cache.get_state(courier_id).await.unwrap().unwrap();
        assert_eq!(retrieved.status, CourierStatus::Archived);
    }

    #[tokio::test]
    async fn test_update_max_load() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let state = test_state();

        cache
            .initialize_state(courier_id, state, "Berlin-Mitte")
            .await
            .unwrap();

        cache.update_max_load(courier_id, 5).await.unwrap();

        let retrieved = cache.get_state(courier_id).await.unwrap().unwrap();
        assert_eq!(retrieved.max_load, 5);
    }

    #[tokio::test]
    async fn test_add_to_free_pool() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        cache
            .add_to_free_pool(courier_id, "Berlin-Mitte")
            .await
            .unwrap();

        let all_free = cache.get_all_free_couriers().await.unwrap();
        assert!(all_free.contains(&courier_id));

        let zone_free = cache.get_free_couriers_in_zone("Berlin-Mitte").await.unwrap();
        assert!(zone_free.contains(&courier_id));
    }

    #[tokio::test]
    async fn test_remove_from_free_pool() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        // Add first
        cache
            .add_to_free_pool(courier_id, "Berlin-Mitte")
            .await
            .unwrap();

        // Then remove
        cache
            .remove_from_free_pool(courier_id, "Berlin-Mitte")
            .await
            .unwrap();

        let all_free = cache.get_all_free_couriers().await.unwrap();
        assert!(!all_free.contains(&courier_id));

        let zone_free = cache.get_free_couriers_in_zone("Berlin-Mitte").await.unwrap();
        assert!(!zone_free.contains(&courier_id));
    }

    #[tokio::test]
    async fn test_exists() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        // Should not exist initially
        let exists = cache.exists(courier_id).await.unwrap();
        assert!(!exists);

        // Initialize state
        cache
            .initialize_state(courier_id, test_state(), "Berlin-Mitte")
            .await
            .unwrap();

        // Should exist now
        let exists = cache.exists(courier_id).await.unwrap();
        assert!(exists);
    }

    #[tokio::test]
    async fn test_remove() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        // Initialize state and add to free pool
        let mut state = test_state();
        state.status = CourierStatus::Free;
        cache
            .initialize_state(courier_id, state, "Berlin-Mitte")
            .await
            .unwrap();
        cache
            .add_to_free_pool(courier_id, "Berlin-Mitte")
            .await
            .unwrap();

        // Remove
        cache.remove(courier_id, "Berlin-Mitte").await.unwrap();

        // Should not exist
        let exists = cache.exists(courier_id).await.unwrap();
        assert!(!exists);

        // Should not be in free pool
        let all_free = cache.get_all_free_couriers().await.unwrap();
        assert!(!all_free.contains(&courier_id));
    }
}
