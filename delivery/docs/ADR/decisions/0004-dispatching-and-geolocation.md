# 4. Dispatching and Geolocation

Date: 2025-02-06

## Status

Accepted

## Context

Delivery Service must assign packages to couriers efficiently. The system needs:

- An algorithm to select the best available courier for a package (dispatching)
- Use of courier geolocation for distance-based selection
- Clear event flow so OMS and push notifications can react to assignment and delivery status changes

## Decision

### Dispatching algorithm (DispatchService)

The domain service [DispatchService](../../src/domain/services/dispatch.rs) implements courier selection:

1. **Filter by status** — only couriers with status FREE (can_accept_assignment)
2. **Filter by capacity** — courier must have available slots (capacity.can_accept)
3. **Filter by zone** — courier work_zone must match package delivery_zone
4. **Filter by distance** — distance from courier location to pickup must be within courier max_distance_km
5. **Distance calculation** — Haversine formula via Location value object
6. **Sort** — by distance to pickup (primary), then by rating (secondary)
7. **Return** — best match (DispatchResult with courier_id, distance_to_pickup_km, estimated_arrival_minutes)

Rejection reasons (RejectionReason) are: NotAvailable, AtFullCapacity, TooFarFromPickup, NoLocationData, WrongZone.

### Geolocation

- **Courier location** is updated via UpdateLocation use case and stored in location cache (e.g. Redis) and location history (PostgreSQL) for dispatch and analytics.
- **Location** value object holds latitude, longitude, and accuracy; provides distance_to() using Haversine.
- Use cases (e.g. AssignOrder with auto_assign) load candidate couriers with current_location from cache and call DispatchService.find_nearest_courier.

### Events and consumers

- **PackageAssigned** — published when a package is assigned to a courier. Consumers: push notification service (notify courier), OMS (optional if tracking assignment).
- **PackageDelivered** / **PackageNotDelivered** — published when courier confirms delivery outcome. Consumer: OMS (update order delivery status via Kafka).
- **PackageInTransit** — published when courier picks up the package. Consumer: OMS (update order delivery status).

Event definitions live in [domain/model/delivery/events/v1/events.proto](../../src/domain/model/delivery/events/v1/events.proto). Publishing is done via EventPublisher (Kafka implementation in infrastructure).

**Kafka event format (for OMS and other consumers):** Messages in `delivery.package.status.v1` are **protobuf**-encoded. Each message includes a Kafka header `event_type` with the proto message name (e.g. `PackageInTransitEvent`, `PackageDeliveredEvent`, `PackageNotDeliveredEvent`) so consumers can decode the payload with the correct type. OMS consumes these events via the same topic and uses the header for dispatch.

### Verification (event flow)

To verify the OMS ↔ Delivery event flow:

1. **Kafka config** — Both services must use the same topic (`delivery.package.status.v1`), brokers, and (for OMS) consumer group (e.g. `oms-delivery-consumer`). See Delivery `KafkaPublisherConfig` and OMS `DeliveryConsumerConfig`.
2. **Flow** — When Delivery changes a package status (e.g. PickUpOrder → InTransit, DeliverOrder → Delivered/NotDelivered), it publishes a protobuf message with the `event_type` header. OMS consumer receives it, decodes via the event type, maps to `DeliveryStatusEvent`, and calls `HandleDeliveryStatus`; [on_delivery_status handler](../../../../oms/internal/usecases/order/event/on_delivery_status/handler.go) updates the order’s delivery status (IN_TRANSIT, DELIVERED, NOT_DELIVERED) and persists.
3. **AcceptOrder errors** — Delivery returns gRPC `AlreadyExists` for duplicate `order_id` and `InvalidArgument` for invalid address/period; OMS should treat these as permanent failures (no retry) or handle retries according to policy.

## References

- [ADR-0002: C4 System](0002-c4-system.md) — architecture and container diagram
- [ADR-0003: Domain Structure](0003-domain-structure.md) — model/ and services/, DispatchService placement
