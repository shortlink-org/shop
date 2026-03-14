//! Redis-backed pub/sub for customer-facing order tracking updates.

use redis::aio::{ConnectionManager, PubSubStream};
use redis::{AsyncCommands, Client, RedisResult};
use uuid::Uuid;

#[derive(Clone)]
pub struct RedisTrackingPubSub {
    client: Client,
    publisher: ConnectionManager,
}

impl RedisTrackingPubSub {
    pub fn new(client: Client, publisher: ConnectionManager) -> Self {
        Self { client, publisher }
    }

    pub async fn publish_order_update(&self, order_id: Uuid) -> RedisResult<()> {
        let mut publisher = self.publisher.clone();
        let _: usize = publisher
            .publish(Self::channel_name(order_id), order_id.to_string())
            .await?;
        Ok(())
    }

    pub async fn subscribe_order_updates(&self, order_id: Uuid) -> RedisResult<PubSubStream> {
        let pubsub = self.client.get_async_pubsub().await?;
        let (mut sink, stream) = pubsub.split();
        sink.subscribe(Self::channel_name(order_id)).await?;
        Ok(stream)
    }

    pub fn channel_name(order_id: Uuid) -> String {
        format!("delivery:tracking:{}", order_id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    use futures_util::StreamExt;
    use testcontainers::{runners::AsyncRunner, ImageExt};
    use testcontainers_modules::redis::Redis;
    use tokio::time::timeout;

    use crate::test_support::ManagedAsyncContainer;

    async fn setup_pubsub() -> (ManagedAsyncContainer<Redis>, RedisTrackingPubSub) {
        let container = ManagedAsyncContainer::new(
            Redis::default().with_tag("7-alpine").start().await.unwrap(),
        );
        let port = container.get_host_port_ipv4(6379).await.unwrap();
        let url = format!("redis://localhost:{port}");

        let client = redis::Client::open(url).unwrap();
        let publisher = ConnectionManager::new(client.clone()).await.unwrap();

        (container, RedisTrackingPubSub::new(client, publisher))
    }

    #[tokio::test]
    async fn publish_order_update_reaches_order_channel() {
        let (container, pubsub) = setup_pubsub().await;
        let order_id = Uuid::new_v4();
        let mut stream = pubsub.subscribe_order_updates(order_id).await.unwrap();

        pubsub.publish_order_update(order_id).await.unwrap();

        let message = timeout(Duration::from_secs(5), stream.next())
            .await
            .expect("timed out waiting for redis pubsub message")
            .expect("pubsub stream closed unexpectedly");

        let payload: String = message.get_payload().unwrap();
        assert_eq!(payload, order_id.to_string());
        assert_eq!(
            message.get_channel_name(),
            RedisTrackingPubSub::channel_name(order_id)
        );

        container.rm().await.unwrap();
    }
}
