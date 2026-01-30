# 3. Domain Structure - Model and Services

Date: 2024-01-30

## Status

Accepted

## Context

The domain layer needs clear organization following DDD principles. We need to separate:

- Data models and aggregates
- Business logic that spans multiple aggregates
- Pure domain logic from infrastructure concerns

## Decision

Split domain into two main parts:

### model/

Contains all data structures:

- **Aggregates**: `Package`, `Courier` with state machines
- **Value Objects**: `Location`, `Address`, `DeliveryPeriod`
- **Proto definitions**: Commands, Events, Queries

### services/

Contains domain services:

- **DispatchService**: Courier selection algorithm (Haversine distance)
- **AssignmentValidationService**: Business rules validation

### Separation from Use Cases

| Layer           | Responsibility      | I/O                         |
|-----------------|---------------------|-----------------------------|
| Domain Services | Pure business logic | None                        |
| Use Cases       | Orchestration       | Repositories, External APIs |

## Consequences

### Positive

- Clear separation of concerns
- Domain services are easy to unit test
- Aggregates encapsulate their own state transitions
- Reusable business logic

### Negative

- More files and directories to maintain
- Need discipline to keep domain services pure
