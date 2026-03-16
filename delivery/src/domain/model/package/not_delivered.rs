//! Structured not-delivered details for package state.

use std::{fmt, str::FromStr};

/// Machine-readable code for a failed delivery attempt.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum NotDeliveredReasonCode {
    CustomerUnavailable,
    WrongAddress,
    Refused,
    AccessDenied,
    Other,
}

impl NotDeliveredReasonCode {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::CustomerUnavailable => "CUSTOMER_NOT_AVAILABLE",
            Self::WrongAddress => "WRONG_ADDRESS",
            Self::Refused => "CUSTOMER_REFUSED",
            Self::AccessDenied => "ACCESS_DENIED",
            Self::Other => "OTHER",
        }
    }
}

impl fmt::Display for NotDeliveredReasonCode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

impl FromStr for NotDeliveredReasonCode {
    type Err = String;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value.trim().to_ascii_uppercase().as_str() {
            "CUSTOMER_NOT_AVAILABLE" => Ok(Self::CustomerUnavailable),
            "WRONG_ADDRESS" => Ok(Self::WrongAddress),
            "CUSTOMER_REFUSED" => Ok(Self::Refused),
            "ACCESS_DENIED" => Ok(Self::AccessDenied),
            "OTHER" => Ok(Self::Other),
            _ => Err(format!("unknown not delivered reason code: {}", value)),
        }
    }
}

/// Structured details about why a package was not delivered.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct NotDeliveredDetails {
    code: NotDeliveredReasonCode,
    description: Option<String>,
}

impl NotDeliveredDetails {
    pub fn new(code: NotDeliveredReasonCode, description: Option<String>) -> Self {
        let description = description.and_then(|value| {
            let trimmed = value.trim();
            if trimmed.is_empty() {
                None
            } else {
                Some(trimmed.to_string())
            }
        });

        Self { code, description }
    }

    pub fn code(&self) -> NotDeliveredReasonCode {
        self.code
    }

    pub fn description(&self) -> Option<&str> {
        self.description.as_deref()
    }
}

impl fmt::Display for NotDeliveredDetails {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.description() {
            Some(description) => write!(f, "{}: {}", self.code, description),
            None => write!(f, "{}", self.code),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn display_without_description_uses_code_only() {
        let details = NotDeliveredDetails::new(NotDeliveredReasonCode::WrongAddress, None);

        assert_eq!(details.to_string(), "WRONG_ADDRESS");
    }

    #[test]
    fn display_with_description_keeps_structure() {
        let details = NotDeliveredDetails::new(
            NotDeliveredReasonCode::Other,
            Some("customer asked to reschedule".to_string()),
        );

        assert_eq!(details.to_string(), "OTHER: customer asked to reschedule");
    }
}
