//! Notifications Infrastructure
//!
//! Contains implementations for push notification services.
//! Currently provides a stub implementation - FCM/APNs integration is TODO.

pub mod stub;

pub use stub::StubNotificationService;
