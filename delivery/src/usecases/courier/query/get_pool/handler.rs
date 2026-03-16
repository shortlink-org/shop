//! Get Courier Pool Handler
//!
//! Handles retrieving couriers with filtering and pagination.
//!
//! ## Flow
//! 1. Build query from filters
//! 2. Get courier IDs from cache (for status-based filters)
//! 3. Load courier profiles from repository
//! 4. Optionally fetch current locations from Geolocation Service
//! 5. Return paginated results

use std::sync::Arc;

use thiserror::Error;

use crate::domain::model::courier::{Courier, CourierStatus};
use crate::domain::ports::{
    CacheError, CachedCourierState, CourierCache, CourierRepository, QueryHandler, RepositoryError,
};

use super::Query;

/// Errors that can occur during courier pool retrieval
#[derive(Debug, Error)]
pub enum GetCourierPoolError {
    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Courier data with state from cache
#[derive(Debug, Clone)]
pub struct CourierWithState {
    /// Courier profile from database
    pub courier: Courier,
    /// Cached state (status, load, rating)
    pub state: Option<CachedCourierState>,
}

/// Response from get courier pool query
#[derive(Debug)]
pub struct Response {
    /// List of couriers with their states
    pub couriers: Vec<CourierWithState>,
    /// Total count of matching couriers
    pub total_count: usize,
}

/// Get Courier Pool Handler
pub struct Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    repository: Arc<R>,
    cache: Arc<C>,
}

impl<R, C> Handler<R, C>
where
    R: CourierRepository,
    C: CourierCache,
{
    /// Create a new handler instance
    pub fn new(repository: Arc<R>, cache: Arc<C>) -> Self {
        Self { repository, cache }
    }

    fn to_repo_filter(query: &Query) -> crate::domain::ports::CourierFilter {
        let effective_status = if query.filter.available_only {
            Some(CourierStatus::Free)
        } else {
            query.filter.status
        };

        crate::domain::ports::CourierFilter {
            work_zone: query.filter.work_zone.clone(),
            status: effective_status,
            archived: Some(false),
        }
    }
}

impl<R, C> QueryHandler<Query, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = GetCourierPoolError;

    /// Handle the GetCourierPool query
    async fn handle(&self, query: Query) -> Result<Response, Self::Error> {
        let repo_filter = Self::to_repo_filter(&query);
        let couriers = self
            .repository
            .find_by_filter(
                repo_filter,
                query.limit.unwrap_or(50),
                query.offset.unwrap_or(0),
            )
            .await?;

        let mut couriers_with_state = Vec::with_capacity(couriers.len());

        for courier in couriers {
            if let Some(ref transport) = query.filter.transport_type {
                if courier.transport_type() != *transport {
                    continue;
                }
            }

            let state = self
                .cache
                .get_state(courier.id().0)
                .await
                .ok()
                .flatten()
                .or_else(|| Some(CachedCourierState::from(&courier)));

            couriers_with_state.push(CourierWithState { courier, state });
        }

        let total_count = couriers_with_state.len();

        Ok(Response {
            couriers: couriers_with_state,
            total_count,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::WorkHours;
    use crate::domain::model::vo::TransportType;
    use crate::domain::ports::QueryHandler;
    use crate::infrastructure::cache::CourierRedisCache;
    use crate::infrastructure::repository::CourierPostgresRepository;
    use chrono::NaiveTime;
    use migration::{Migrator, MigratorTrait};
    use sea_orm::{Database, DatabaseConnection};
    use testcontainers::{runners::AsyncRunner, ImageExt};
    use testcontainers_modules::{postgres::Postgres, redis::Redis};

    use crate::test_support::ManagedAsyncContainer;

    /// Test environment with PostgreSQL and Redis containers
    struct TestEnv {
        _pg_container: ManagedAsyncContainer<Postgres>,
        _redis_container: ManagedAsyncContainer<Redis>,
        repository: Arc<CourierPostgresRepository>,
        cache: Arc<CourierRedisCache>,
    }

    async fn setup_test_env() -> TestEnv {
        // Start PostgreSQL container
        let pg_container = ManagedAsyncContainer::new(
            Postgres::default()
                .with_tag("18-alpine")
                .start()
                .await
                .unwrap(),
        );
        let pg_port = pg_container.get_host_port_ipv4(5432).await.unwrap();
        let pg_url = format!(
            "postgres://postgres:postgres@localhost:{}/postgres",
            pg_port
        );
        let db: DatabaseConnection = Database::connect(&pg_url).await.unwrap();

        // Apply migrations
        Migrator::up(&db, None).await.unwrap();

        // Start Redis container
        let redis_container = ManagedAsyncContainer::new(
            Redis::default().with_tag("7-alpine").start().await.unwrap(),
        );
        let redis_port = redis_container.get_host_port_ipv4(6379).await.unwrap();
        let redis_url = format!("redis://localhost:{}", redis_port);
        let redis_client = redis::Client::open(redis_url).unwrap();
        let redis_conn = redis::aio::ConnectionManager::new(redis_client)
            .await
            .unwrap();

        TestEnv {
            _pg_container: pg_container,
            _redis_container: redis_container,
            repository: Arc::new(CourierPostgresRepository::new(db)),
            cache: Arc::new(CourierRedisCache::new(redis_conn)),
        }
    }

    fn create_test_courier(name: &str, phone: &str, email: &str, zone: &str) -> Courier {
        let work_hours = WorkHours::new(
            NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();

        Courier::builder(
            name.to_string(),
            phone.to_string(),
            email.to_string(),
            TransportType::Bicycle,
            10.0,
            zone.to_string(),
            work_hours,
        )
        .build()
        .unwrap()
    }

    #[tokio::test]
    async fn test_get_pool_returns_all_couriers_when_no_filters() {
        // Arrange
        let env = setup_test_env().await;

        // Create and save test couriers
        let courier1 = create_test_courier("Courier 1", "+49111000001", "c1@test.de", "Berlin");
        let courier2 = create_test_courier("Courier 2", "+49222000002", "c2@test.de", "Munich");
        let courier3 = create_test_courier("Courier 3", "+49333000003", "c3@test.de", "Berlin");

        env.repository.save(&courier1).await.unwrap();
        env.repository.save(&courier2).await.unwrap();
        env.repository.save(&courier3).await.unwrap();

        let handler = Handler::new(env.repository.clone(), env.cache.clone());
        let query = Query::all();

        // Act
        let result = handler.handle(query).await;

        // Assert
        assert!(result.is_ok());
        let response = result.unwrap();
        // Seed migration adds 10 couriers + our 3 test couriers = 13
        assert!(response.total_count >= 3);
        assert!(response.couriers.len() >= 3);
    }

    #[tokio::test]
    async fn test_get_pool_with_pagination() {
        // Arrange
        let env = setup_test_env().await;

        // Create 5 test couriers
        for i in 1..=5 {
            let courier = create_test_courier(
                &format!("Courier {}", i),
                &format!("+49{:09}", i),
                &format!("c{}@test.de", i),
                "TestZone",
            );
            env.repository.save(&courier).await.unwrap();
        }

        let handler = Handler::new(env.repository.clone(), env.cache.clone());

        // Act - get first 2
        let query = Query::all().with_limit(2).with_offset(0);
        let result = handler.handle(query).await.unwrap();

        // Assert
        assert_eq!(result.couriers.len(), 2);

        // Act - get next 2
        let query = Query::all().with_limit(2).with_offset(2);
        let result = handler.handle(query).await.unwrap();

        // Assert
        assert_eq!(result.couriers.len(), 2);
    }

    #[tokio::test]
    async fn test_get_pool_filters_by_zone() {
        // Arrange
        let env = setup_test_env().await;

        // Create couriers in different zones
        let courier_berlin = create_test_courier(
            "Berlin Courier",
            "+49111000011",
            "berlin@test.de",
            "Berlin-Test",
        );
        let courier_munich = create_test_courier(
            "Munich Courier",
            "+49222000022",
            "munich@test.de",
            "Munich-Test",
        );

        env.repository.save(&courier_berlin).await.unwrap();
        env.repository.save(&courier_munich).await.unwrap();

        let handler = Handler::new(env.repository.clone(), env.cache.clone());
        let query = Query::in_zone("Berlin-Test");

        // Act
        let result = handler.handle(query).await;

        // Assert
        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.total_count, 1);
        assert_eq!(response.couriers[0].courier.work_zone(), "Berlin-Test");
    }

    #[tokio::test]
    async fn test_get_pool_filters_free_couriers_from_cache() {
        // Arrange
        let env = setup_test_env().await;

        // Create and save a courier
        let mut courier =
            create_test_courier("Free Courier", "+49111000033", "free@test.de", "FreeZone");
        courier.go_online().unwrap();
        let courier_id = courier.id().0;
        env.repository.save(&courier).await.unwrap();
        env.cache.cache(&courier).await.unwrap();

        let handler = Handler::new(env.repository.clone(), env.cache.clone());
        let query = Query::new(super::super::CourierFilter {
            status: Some(CourierStatus::Free),
            ..Default::default()
        });

        // Act
        let result = handler.handle(query).await;

        // Assert
        assert!(result.is_ok());
        let response = result.unwrap();
        assert_eq!(response.total_count, 1);
        assert_eq!(response.couriers[0].courier.id().0, courier_id);
        assert!(response.couriers[0].state.is_some());
        assert_eq!(
            response.couriers[0].state.as_ref().unwrap().status,
            CourierStatus::Free
        );
    }

    #[tokio::test]
    async fn test_get_pool_returns_seeded_couriers() {
        // Arrange
        let env = setup_test_env().await;

        let handler = Handler::new(env.repository.clone(), env.cache.clone());
        let query = Query::all();

        // Act
        let result = handler.handle(query).await;

        // Assert
        assert!(result.is_ok());
        let response = result.unwrap();
        // Seed migration adds 10 couriers
        assert_eq!(response.total_count, 10);
        assert_eq!(response.couriers.len(), 10);
    }
}
