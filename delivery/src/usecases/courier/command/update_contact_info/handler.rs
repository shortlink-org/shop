//! Update Contact Info Handler
//!
//! Handles updating courier contact information.
//!
//! ## Flow
//! 1. Validate courier exists
//! 2. Check courier is not archived
//! 3. Validate new email/phone are unique if changed
//! 4. Update courier in repository
//! 5. Return update result

use std::sync::Arc;

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::boundary::ports::{
    CacheError, CommandHandlerWithResult, CourierCache, CourierRepository, RepositoryError,
};
use crate::domain::model::courier::CourierStatus;

use super::Command;

/// Errors that can occur during contact info update
#[derive(Debug, Error)]
pub enum UpdateContactInfoError {
    /// Courier not found
    #[error("Courier not found: {0}")]
    NotFound(Uuid),

    /// Courier is archived
    #[error("Cannot update archived courier: {0}")]
    CourierArchived(Uuid),

    /// Email already exists
    #[error("Email already registered: {0}")]
    EmailExists(String),

    /// Phone already exists
    #[error("Phone already registered: {0}")]
    PhoneExists(String),

    /// No fields to update
    #[error("No fields to update")]
    NoFieldsToUpdate,

    /// Repository error
    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),

    /// Cache error
    #[error("Cache error: {0}")]
    CacheError(#[from] CacheError),
}

/// Response from updating contact info
#[derive(Debug, Clone)]
pub struct Response {
    /// Courier ID
    pub courier_id: Uuid,
    /// Current phone
    pub phone: String,
    /// Current email
    pub email: String,
    /// Update timestamp
    pub updated_at: DateTime<Utc>,
}

/// Update Contact Info Handler
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
}

impl<R, C> CommandHandlerWithResult<Command, Response> for Handler<R, C>
where
    R: CourierRepository + Send + Sync,
    C: CourierCache + Send + Sync,
{
    type Error = UpdateContactInfoError;

    /// Handle the UpdateContactInfo command
    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // Check if there's anything to update
        if cmd.phone.is_none() && cmd.email.is_none() && cmd.push_token.is_none() {
            return Err(UpdateContactInfoError::NoFieldsToUpdate);
        }

        // 1. Validate courier exists
        let mut courier = self
            .repository
            .find_by_id(cmd.courier_id)
            .await?
            .ok_or(UpdateContactInfoError::NotFound(cmd.courier_id))?;

        // 2. Check courier is not archived
        if let Ok(Some(state)) = self.cache.get_state(cmd.courier_id).await {
            if state.status == CourierStatus::Archived {
                return Err(UpdateContactInfoError::CourierArchived(cmd.courier_id));
            }
        }

        // 3. Validate uniqueness for email if changing
        if let Some(ref new_email) = cmd.email {
            if new_email != courier.email() && self.repository.email_exists(new_email).await? {
                return Err(UpdateContactInfoError::EmailExists(new_email.clone()));
            }
        }

        // 3. Validate uniqueness for phone if changing
        if let Some(ref new_phone) = cmd.phone {
            if new_phone != courier.phone() && self.repository.phone_exists(new_phone).await? {
                return Err(UpdateContactInfoError::PhoneExists(new_phone.clone()));
            }
        }

        // 4. Update courier
        if let Some(phone) = cmd.phone {
            courier.update_phone(phone);
        }
        if let Some(email) = cmd.email {
            courier.update_email(email);
        }
        if let Some(push_token) = cmd.push_token {
            courier.update_push_token(Some(push_token));
        }

        self.repository.save(&courier).await?;

        let updated_at = Utc::now();

        Ok(Response {
            courier_id: cmd.courier_id,
            phone: courier.phone().to_string(),
            email: courier.email().to_string(),
            updated_at,
        })
    }
}

#[cfg(test)]
mod tests {
    // TODO: Add unit tests with mock repository and cache
}
