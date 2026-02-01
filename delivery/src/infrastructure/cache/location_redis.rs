//! Redis Implementation of LocationCache
//!
//! Uses Redis for caching courier current locations (hot data).
//!
//! Redis Key Schema:
//! - `location:{courier_id}` - HASH containing location data
//! - `locations:active` - SET of courier IDs with active locations

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use redis::aio::ConnectionManager;
use redis::AsyncCommands;
use uuid::Uuid;

use crate::domain::model::vo::location::Location;
use crate::domain::model::CourierLocation;
use crate::domain::ports::{LocationCache, LocationCacheError};

/// Redis key prefix for courier locations
const LOCATION_PREFIX: &str = "location";
/// Redis key for active courier locations set
const ACTIVE_LOCATIONS_KEY: &str = "locations:active";
/// Default TTL for location cache entries (5 minutes)
const DEFAULT_TTL_SECONDS: u64 = 300;

/// Redis implementation of LocationCache
pub struct RedisLocationCache {
    conn: ConnectionManager,
    ttl_seconds: u64,
}

impl RedisLocationCache {
    /// Create a new Redis location cache instance
    pub fn new(conn: ConnectionManager) -> Self {
        Self {
            conn,
            ttl_seconds: DEFAULT_TTL_SECONDS,
        }
    }

    /// Create with custom TTL
    pub fn with_ttl(conn: ConnectionManager, ttl_seconds: u64) -> Self {
        Self { conn, ttl_seconds }
    }

    /// Get the location key for a courier
    fn location_key(courier_id: Uuid) -> String {
        format!("{}:{}", LOCATION_PREFIX, courier_id)
    }
}

#[async_trait]
impl LocationCache for RedisLocationCache {
    async fn set_location(
        &self,
        location: &CourierLocation,
    ) -> Result<(), LocationCacheError> {
        let mut conn = self.conn.clone();
        let key = Self::location_key(location.courier_id());
        let courier_id_str = location.courier_id().to_string();

        // Store location as hash with TTL
        let mut pipe = redis::pipe();
        pipe.hset(&key, "latitude", location.latitude().to_string())
            .hset(&key, "longitude", location.longitude().to_string())
            .hset(&key, "accuracy", location.accuracy().to_string())
            .hset(&key, "timestamp", location.timestamp().to_rfc3339());

        if let Some(speed) = location.speed() {
            pipe.hset(&key, "speed", speed.to_string());
        }
        if let Some(heading) = location.heading() {
            pipe.hset(&key, "heading", heading.to_string());
        }

        // Set TTL and add to active set
        pipe.expire(&key, self.ttl_seconds as i64)
            .sadd(ACTIVE_LOCATIONS_KEY, &courier_id_str);

        let _: () = pipe
            .query_async(&mut conn)
            .await
            .map_err(|e| LocationCacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn get_location(
        &self,
        courier_id: Uuid,
    ) -> Result<Option<CourierLocation>, LocationCacheError> {
        let mut conn = self.conn.clone();
        let key = Self::location_key(courier_id);

        let result: Option<Vec<(String, String)>> = conn
            .hgetall(&key)
            .await
            .map_err(|e| LocationCacheError::OperationError(e.to_string()))?;

        match result {
            Some(fields) if !fields.is_empty() => {
                let mut latitude: Option<f64> = None;
                let mut longitude: Option<f64> = None;
                let mut accuracy: f64 = 10.0;
                let mut timestamp: Option<DateTime<Utc>> = None;
                let mut speed: Option<f64> = None;
                let mut heading: Option<f64> = None;

                for (field, value) in fields {
                    match field.as_str() {
                        "latitude" => latitude = value.parse().ok(),
                        "longitude" => longitude = value.parse().ok(),
                        "accuracy" => accuracy = value.parse().unwrap_or(10.0),
                        "timestamp" => {
                            timestamp = DateTime::parse_from_rfc3339(&value)
                                .ok()
                                .map(|dt| dt.with_timezone(&Utc));
                        }
                        "speed" => speed = value.parse().ok(),
                        "heading" => heading = value.parse().ok(),
                        _ => {}
                    }
                }

                match (latitude, longitude, timestamp) {
                    (Some(lat), Some(lon), Some(ts)) => {
                        let location = Location::new(lat, lon, accuracy).map_err(|e| {
                            LocationCacheError::SerializationError(format!(
                                "Invalid location: {}",
                                e
                            ))
                        })?;

                        let courier_location =
                            CourierLocation::from_stored(courier_id, location, ts, speed, heading)
                                .map_err(|e| {
                                    LocationCacheError::SerializationError(format!(
                                        "Invalid courier location: {}",
                                        e
                                    ))
                                })?;

                        Ok(Some(courier_location))
                    }
                    _ => Ok(None),
                }
            }
            _ => Ok(None),
        }
    }

    async fn get_locations(
        &self,
        courier_ids: &[Uuid],
    ) -> Result<Vec<CourierLocation>, LocationCacheError> {
        let mut locations = Vec::with_capacity(courier_ids.len());

        for &courier_id in courier_ids {
            if let Some(location) = self.get_location(courier_id).await? {
                locations.push(location);
            }
        }

        Ok(locations)
    }

    async fn delete_location(&self, courier_id: Uuid) -> Result<(), LocationCacheError> {
        let mut conn = self.conn.clone();
        let key = Self::location_key(courier_id);
        let courier_id_str = courier_id.to_string();

        let _: () = redis::pipe()
            .del(&key)
            .srem(ACTIVE_LOCATIONS_KEY, &courier_id_str)
            .query_async(&mut conn)
            .await
            .map_err(|e| LocationCacheError::OperationError(e.to_string()))?;

        Ok(())
    }

    async fn has_location(&self, courier_id: Uuid) -> Result<bool, LocationCacheError> {
        let mut conn = self.conn.clone();
        let key = Self::location_key(courier_id);

        let exists: bool = conn
            .exists(&key)
            .await
            .map_err(|e| LocationCacheError::OperationError(e.to_string()))?;

        Ok(exists)
    }

    async fn get_all_locations(&self) -> Result<Vec<CourierLocation>, LocationCacheError> {
        let courier_ids = self.get_active_courier_ids().await?;
        self.get_locations(&courier_ids).await
    }

    async fn get_active_courier_ids(&self) -> Result<Vec<Uuid>, LocationCacheError> {
        let mut conn = self.conn.clone();

        let members: Vec<String> = conn
            .smembers(ACTIVE_LOCATIONS_KEY)
            .await
            .map_err(|e| LocationCacheError::OperationError(e.to_string()))?;

        let mut uuids = Vec::with_capacity(members.len());
        for member in members {
            if let Ok(uuid) = Uuid::parse_str(&member) {
                // Verify the location still exists (not expired)
                if self.has_location(uuid).await.unwrap_or(false) {
                    uuids.push(uuid);
                }
            }
        }

        Ok(uuids)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;
    use testcontainers::{runners::AsyncRunner, ContainerAsync, ImageExt};
    use testcontainers_modules::redis::Redis;

    async fn setup_redis() -> (ContainerAsync<Redis>, RedisLocationCache) {
        let container = Redis::default().with_tag("7-alpine").start().await.unwrap();
        let port = container.get_host_port_ipv4(6379).await.unwrap();
        let url = format!("redis://localhost:{}", port);

        let client = redis::Client::open(url).unwrap();
        let conn = redis::aio::ConnectionManager::new(client).await.unwrap();
        let cache = RedisLocationCache::new(conn);

        (container, cache)
    }

    fn create_test_location(courier_id: Uuid) -> CourierLocation {
        let location = Location::new(52.52, 13.405, 10.0).unwrap();
        CourierLocation::from_stored(courier_id, location, Utc::now(), Some(35.0), Some(180.0))
            .unwrap()
    }

    #[tokio::test]
    async fn test_set_and_get_location() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let location = create_test_location(courier_id);

        cache.set_location(&location).await.unwrap();

        let retrieved = cache.get_location(courier_id).await.unwrap();
        assert!(retrieved.is_some());

        let retrieved = retrieved.unwrap();
        assert_eq!(retrieved.courier_id(), courier_id);
        assert!((retrieved.latitude() - 52.52).abs() < 0.001);
        assert!((retrieved.longitude() - 13.405).abs() < 0.001);
        assert_eq!(retrieved.speed(), Some(35.0));
        assert_eq!(retrieved.heading(), Some(180.0));
    }

    #[tokio::test]
    async fn test_get_location_not_found() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();

        let result = cache.get_location(courier_id).await.unwrap();
        assert!(result.is_none());
    }

    #[tokio::test]
    async fn test_delete_location() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let location = create_test_location(courier_id);

        cache.set_location(&location).await.unwrap();
        cache.delete_location(courier_id).await.unwrap();

        let result = cache.get_location(courier_id).await.unwrap();
        assert!(result.is_none());
    }

    #[tokio::test]
    async fn test_has_location() {
        let (_container, cache) = setup_redis().await;
        let courier_id = Uuid::new_v4();
        let location = create_test_location(courier_id);

        assert!(!cache.has_location(courier_id).await.unwrap());

        cache.set_location(&location).await.unwrap();

        assert!(cache.has_location(courier_id).await.unwrap());
    }

    #[tokio::test]
    async fn test_get_active_courier_ids() {
        let (_container, cache) = setup_redis().await;
        let courier1 = Uuid::new_v4();
        let courier2 = Uuid::new_v4();

        cache
            .set_location(&create_test_location(courier1))
            .await
            .unwrap();
        cache
            .set_location(&create_test_location(courier2))
            .await
            .unwrap();

        let active = cache.get_active_courier_ids().await.unwrap();
        assert!(active.contains(&courier1));
        assert!(active.contains(&courier2));
    }

    #[tokio::test]
    async fn test_get_locations_batch() {
        let (_container, cache) = setup_redis().await;
        let courier1 = Uuid::new_v4();
        let courier2 = Uuid::new_v4();
        let courier3 = Uuid::new_v4(); // Not added

        cache
            .set_location(&create_test_location(courier1))
            .await
            .unwrap();
        cache
            .set_location(&create_test_location(courier2))
            .await
            .unwrap();

        let locations = cache
            .get_locations(&[courier1, courier2, courier3])
            .await
            .unwrap();

        assert_eq!(locations.len(), 2);
    }
}
