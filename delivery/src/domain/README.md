# Delivery Domain Layer

## Структура

```
domain/
├── model/                          # Domain Models
│   ├── delivery/                   # Proto-generated models
│   │   ├── common/v1/              # Common types (Address, Location, enums)
│   │   ├── commands/v1/            # CQRS Commands
│   │   ├── events/v1/              # Domain Events
│   │   └── queries/v1/             # Queries
│   ├── package/                    # Package Aggregate
│   │   └── state.rs                # Status state machine
│   ├── courier/                    # Courier Aggregate
│   │   └── state.rs                # Status, capacity management
│   └── vo/                         # Value Objects
│       └── location.rs             # GPS Location with Haversine
│
├── services/                       # Domain Services
│   ├── dispatch.rs                 # Courier selection algorithm
│   └── assignment_validation.rs    # Assignment business rules
│
└── mod.rs
```

## Domain Models (`model/`)

### Proto-generated Models (`model/delivery/`)

Protobuf определения, из которых генерируются Rust структуры:

- **common/v1** - общие типы: `Address`, `Location`, `DeliveryPeriod`, `PackageInfo`, `WorkHours`, enums
- **commands/v1** - CQRS команды: `AcceptOrderCommand`, `AssignOrderCommand`, `DeliverOrderCommand`, etc.
- **events/v1** - доменные события: `PackageAcceptedEvent`, `PackageAssignedEvent`, etc.
- **queries/v1** - запросы: `GetPackagePoolQuery`, `GetCourierPoolQuery`

### Package Aggregate (`model/package/`)

```rust
pub enum PackageStatus {
    Accepted,
    InPool,
    Assigned,
    InTransit,
    Delivered,
    NotDelivered,
    RequiresHandling,
}

// Valid transitions:
// Accepted -> InPool -> Assigned -> InTransit -> Delivered | NotDelivered
// NotDelivered -> RequiresHandling -> InPool (return to pool)
```

### Courier Aggregate (`model/courier/`)

```rust
pub enum CourierStatus {
    Unavailable,  // Off-duty
    Free,         // Ready for assignments
    Busy,         // Has active deliveries
}

pub struct CourierCapacity {
    current_load: u32,
    max_load: u32,
}
```

### Value Objects (`model/vo/`)

- **Location** - GPS координаты с валидацией и расчётом расстояния (Haversine)

## Domain Services (`services/`)

Domain Services содержат бизнес-логику, которая:
- Работает с несколькими агрегатами
- Не принадлежит одной сущности
- Не имеет инфраструктурных зависимостей

### DispatchService (`dispatch.rs`)

Алгоритм выбора оптимального курьера:

```rust
impl DispatchService {
    /// 1. Filter by status (FREE)
    /// 2. Filter by capacity
    /// 3. Filter by zone
    /// 4. Filter by max distance
    /// 5. Calculate Haversine distance
    /// 6. Sort by distance, then rating
    pub fn find_nearest_courier(
        couriers: &[CourierForDispatch],
        package: &PackageForDispatch,
    ) -> Option<DispatchResult>;
}
```

### AssignmentValidationService (`assignment_validation.rs`)

Валидация бизнес-правил назначения:

```rust
impl AssignmentValidationService {
    /// Validates:
    /// - Package status (must be InPool)
    /// - Courier status (must be Free)
    /// - Working hours
    /// - Courier capacity
    /// - Distance to pickup
    pub fn validate(
        courier: &CourierAvailability,
        package: &PackageForValidation,
        current_hour: u8,
    ) -> Result<(), Vec<AssignmentValidationError>>;
}
```

## Domain Services vs Use Cases

| Aspect | Domain Services | Use Cases |
|--------|----------------|-----------|
| **Location** | `domain/services/` | `usecases/` |
| **Dependencies** | Only domain objects | Repositories, external services |
| **I/O** | None | Database, HTTP, messaging |
| **Purpose** | Business logic | Orchestration |

### Example

```rust
// Domain Service - pure logic
fn find_nearest_courier(couriers: &[Courier], location: &Location) -> Option<&Courier>

// Use Case - orchestration
async fn assign_order(package_id: &str) -> Result<()> {
    let package = repo.get(package_id).await?;           // I/O
    let couriers = courier_repo.get_available().await?;   // I/O
    let courier = dispatch_service.find_nearest(&couriers, &package.pickup)?; // Domain
    repo.save(package).await?;                            // I/O
    push.notify(courier).await?;                          // I/O
}
```

## Workflow

```
                    ┌──────────────────────────────────────────┐
                    │              Use Case Layer               │
                    │  (orchestration, repositories, I/O)      │
                    └─────────────────┬────────────────────────┘
                                      │
                    ┌─────────────────▼────────────────────────┐
                    │           Domain Services                 │
                    │  (DispatchService, ValidationService)    │
                    └─────────────────┬────────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        │                             │                             │
┌───────▼───────┐           ┌────────▼────────┐           ┌────────▼────────┐
│   Package     │           │    Courier      │           │  Value Objects  │
│   Aggregate   │           │    Aggregate    │           │   (Location)    │
└───────────────┘           └─────────────────┘           └─────────────────┘
```
