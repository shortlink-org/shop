//! Domain Layer
//!
//! Contains all domain logic following DDD principles:
//!
//! ## Structure
//!
//! - `model/` - Domain models
//!   - `delivery/` - Proto-generated models (commands, events, queries)
//!   - `package/` - Package aggregate with state machine
//!   - `courier/` - Courier aggregate with capacity management
//!   - `vo/` - Value Objects (Location, etc.)
//!
//! - `ports/` - Port Interfaces (Hexagonal Architecture)
//!   - Traits defining contracts between domain and infrastructure
//!   - Repository ports, cache ports, handler traits
//!
//! - `services/` - Domain Services
//!   - Business logic that operates on multiple aggregates
//!   - No infrastructure dependencies
//!
//! ## Domain Services vs Use Cases
//!
//! | Layer | Responsibilities |
//! |-------|-----------------|
//! | Domain Services | Pure business logic, no I/O |
//! | Use Cases | Orchestration, repositories, external services |

pub mod model;
pub mod ports;
pub mod services;
