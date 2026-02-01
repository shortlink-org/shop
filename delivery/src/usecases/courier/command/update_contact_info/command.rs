//! Update Contact Info Command
//!
//! Data structure representing the command to update courier contact info.

use uuid::Uuid;

/// Command to update courier contact information
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to update
    pub courier_id: Uuid,
    /// New phone number (optional)
    pub phone: Option<String>,
    /// New email address (optional)
    pub email: Option<String>,
    /// New push notification token (optional)
    pub push_token: Option<String>,
}

impl Command {
    /// Create a new UpdateContactInfo command
    pub fn new(
        courier_id: Uuid,
        phone: Option<String>,
        email: Option<String>,
        push_token: Option<String>,
    ) -> Self {
        Self {
            courier_id,
            phone,
            email,
            push_token,
        }
    }
}
