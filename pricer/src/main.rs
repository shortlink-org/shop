//! Pricer Service
//!
//! Calculates tax and discount for cart/order. Uses OPA for policies (Phase 2).

use tracing::info;
use tracing_subscriber::EnvFilter;

use pricer::config::Config;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::from_env().map_err(|e| {
        eprintln!("Configuration error: {}", e);
        e
    })?;

    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(&config.log_level)),
        )
        .init();

    info!("Pricer service starting");
    info!(policies_discounts = %config.policies_discounts, policies_taxes = %config.policies_taxes, "Policy paths");

    // TODO Phase 2: OPA evaluator
    // TODO Phase 4: gRPC server (CalculateCartDiscount, ApplyPromoCode)
    // TODO Phase 5: OMS integration

    Ok(())
}
