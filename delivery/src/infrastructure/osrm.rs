//! OSRM HTTP adapter built on the generated OpenAPI client.

use std::time::Duration;

use reqwest::{
    header::{HeaderMap, HeaderName, HeaderValue},
    Client,
};
use shop_osrm_client::apis::{configuration::Configuration, nearest_api};
use thiserror::Error;
use tracing::{info_span, Instrument};

#[derive(Debug, Clone)]
pub struct SnappedPoint {
    pub street: Option<String>,
    pub distance_meters: Option<f64>,
    pub latitude: f64,
    pub longitude: f64,
}

#[derive(Debug, Error)]
pub enum OsrmError {
    #[error("failed to build OSRM HTTP client: {0}")]
    ClientBuild(String),
    #[error("OSRM request failed: {0}")]
    Request(String),
    #[error("OSRM returned code {0}")]
    InvalidCode(String),
    #[error("OSRM returned no snapped point")]
    EmptyWaypoints,
    #[error("OSRM returned invalid waypoint location")]
    InvalidLocation,
}

#[derive(Debug, Clone)]
pub struct OsrmClient {
    configuration: Configuration,
}

impl OsrmClient {
    pub fn new(
        base_url: String,
        timeout: Duration,
        auth_header_name: Option<&str>,
        auth_header_value: Option<&str>,
    ) -> Result<Self, OsrmError> {
        let mut default_headers = HeaderMap::new();
        if let (Some(name), Some(value)) = (auth_header_name, auth_header_value) {
            let header_name = HeaderName::from_bytes(name.as_bytes())
                .map_err(|e| OsrmError::ClientBuild(e.to_string()))?;
            let header_value =
                HeaderValue::from_str(value).map_err(|e| OsrmError::ClientBuild(e.to_string()))?;
            default_headers.insert(header_name, header_value);
        }

        let client = Client::builder()
            .timeout(timeout)
            .default_headers(default_headers)
            .build()
            .map_err(|e| OsrmError::ClientBuild(e.to_string()))?;

        let mut configuration = Configuration::new();
        configuration.base_path = base_url;
        configuration.user_agent = Some("shortlink-shop-delivery-osrm".to_string());
        configuration.client = client;

        Ok(Self { configuration })
    }

    pub async fn nearest_driving(
        &self,
        latitude: f64,
        longitude: f64,
    ) -> Result<SnappedPoint, OsrmError> {
        let coordinates = format!("{longitude},{latitude}");
        let span = info_span!(
            "osrm.nearest",
            http.method = "GET",
            osrm.profile = "driving",
            osrm.coordinates = %coordinates
        );
        let response = nearest_api::nearest(
            &self.configuration,
            "driving",
            &coordinates,
            None,
            None,
            None,
            Some(1),
        )
        .instrument(span)
        .await
        .map_err(|e| OsrmError::Request(e.to_string()))?;

        if response.code != "Ok" {
            return Err(OsrmError::InvalidCode(response.code));
        }

        let waypoint = response
            .waypoints
            .and_then(|mut waypoints| waypoints.drain(..).next())
            .ok_or(OsrmError::EmptyWaypoints)?;

        let location = waypoint.location.ok_or(OsrmError::InvalidLocation)?;
        if location.len() < 2 {
            return Err(OsrmError::InvalidLocation);
        }

        Ok(SnappedPoint {
            street: waypoint.name.filter(|value| !value.trim().is_empty()),
            distance_meters: waypoint.distance,
            latitude: location[1],
            longitude: location[0],
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::net::TcpListener;

    #[test]
    fn test_osrm_client_builds() {
        let client = OsrmClient::new(
            "http://localhost:5000".to_string(),
            Duration::from_secs(3),
            None,
            None,
        );
        assert!(client.is_ok());
    }

    #[tokio::test]
    async fn test_nearest_driving_maps_waypoint() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();

        tokio::spawn(async move {
            let (mut stream, _) = listener.accept().await.unwrap();
            let mut buffer = [0_u8; 2048];
            let _ = stream.read(&mut buffer).await.unwrap();
            let body = r#"{"code":"Ok","waypoints":[{"name":"Unter den Linden","distance":4.2,"location":[13.40495,52.52001]}]}"#;
            let response = format!(
                "HTTP/1.1 200 OK\r\ncontent-type: application/json\r\ncontent-length: {}\r\nconnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(response.as_bytes()).await.unwrap();
        });

        let client = OsrmClient::new(
            format!("http://{address}"),
            Duration::from_secs(3),
            None,
            None,
        )
        .unwrap();
        let result = client.nearest_driving(52.52, 13.405).await.unwrap();

        assert_eq!(result.street.as_deref(), Some("Unter den Linden"));
        assert_eq!(result.distance_meters, Some(4.2));
        assert_eq!(result.latitude, 52.52001);
        assert_eq!(result.longitude, 13.40495);
    }

    #[tokio::test]
    async fn test_nearest_driving_rejects_non_ok_code() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();

        tokio::spawn(async move {
            let (mut stream, _) = listener.accept().await.unwrap();
            let mut buffer = [0_u8; 2048];
            let _ = stream.read(&mut buffer).await.unwrap();
            let body = r#"{"code":"NoSegment","waypoints":[]}"#;
            let response = format!(
                "HTTP/1.1 200 OK\r\ncontent-type: application/json\r\ncontent-length: {}\r\nconnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(response.as_bytes()).await.unwrap();
        });

        let client = OsrmClient::new(
            format!("http://{address}"),
            Duration::from_secs(3),
            None,
            None,
        )
        .unwrap();
        let result = client.nearest_driving(52.52, 13.405).await;

        assert!(matches!(result, Err(OsrmError::InvalidCode(code)) if code == "NoSegment"));
    }
}
