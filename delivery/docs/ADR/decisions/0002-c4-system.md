# 2. C4 System Architecture

Date: 2024-01-15

## Status

Accepted

## Context

Define the system architecture for Delivery Service using C4 model.

## Decision

Adopt C4 diagrams to document the architecture at different levels of abstraction.

## Consequences

- Clear visualization of system boundaries and interactions
- Easy to understand for both technical and non-technical stakeholders
- Facilitates onboarding and architectural discussions

### System Context Diagram

```plantuml
@startuml C4_Context_Diagram_for_Delivery
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

LAYOUT_WITH_LEGEND()

title System Context Diagram for Delivery Service

Person(dispatcher, "Dispatcher", "Manages package assignments and handles exceptions.")
Person(courier, "Courier", "Delivers packages to customers.")

System_Boundary(shop, "Shop Boundary") {
    System(delivery, "Delivery Service", "Manages package lifecycle, courier assignments, and delivery tracking.")
    System(oms, "OMS Service", "Order Management System - creates orders for delivery.")
    System(geolocation, "Geolocation Service", "Tracks courier locations in real-time.")
}

System_Ext(push, "Push Notification Service", "Sends notifications to couriers.")

Rel(oms, delivery, "Sends orders for delivery", "gRPC")
Rel(delivery, geolocation, "Gets/Updates courier locations", "gRPC")
Rel(delivery, push, "Sends notifications", "gRPC")
Rel(dispatcher, delivery, "Manages assignments", "gRPC")
Rel(courier, delivery, "Updates delivery status", "gRPC")
Rel(delivery, oms, "Notifies delivery status", "Events")

@enduml
```

### Container Diagram

```plantuml
@startuml C4_Container_Diagram_for_Delivery
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()

title Container Diagram for Delivery Service

Person(dispatcher, "Dispatcher", "Manages assignments.")
Person(courier, "Courier", "Delivers packages.")

System_Boundary(delivery, "Delivery Service") {
    Container(api, "gRPC API", "Rust/Tonic", "Handles incoming gRPC requests.")
    Container(usecases, "Use Cases", "Rust", "Application layer - orchestrates workflows.")
    Container(domain, "Domain Layer", "Rust", "Business logic - aggregates, services, value objects.")
    ContainerDb(db, "Database", "PostgreSQL", "Stores packages and couriers.")
    ContainerQueue(mq, "Message Queue", "Kafka/NATS", "Publishes domain events.")
}

System_Ext(geolocation, "Geolocation Service", "Courier tracking.")
System_Ext(push, "Push Service", "Notifications.")

Rel(dispatcher, api, "Assigns packages", "gRPC")
Rel(courier, api, "Updates status", "gRPC")
Rel(api, usecases, "Delegates to")
Rel(usecases, domain, "Uses")
Rel(usecases, db, "Reads/Writes", "SQL")
Rel(usecases, mq, "Publishes events", "Messages")
Rel(usecases, geolocation, "Gets locations", "gRPC")
Rel(usecases, push, "Sends notifications", "gRPC")

@enduml
```

### Component Diagram (Domain Layer)

```plantuml
@startuml C4_Component_Diagram_for_Delivery_Domain
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram for Delivery Domain Layer

Container_Boundary(domain, "Domain Layer") {
    Component(package_agg, "Package Aggregate", "Rust", "State machine for package lifecycle.")
    Component(courier_agg, "Courier Aggregate", "Rust", "Courier status and capacity.")
    Component(location_vo, "Location VO", "Rust", "GPS coordinates with Haversine distance.")
    Component(dispatch_svc, "DispatchService", "Rust", "Finds nearest available courier.")
    Component(validation_svc, "AssignmentValidationService", "Rust", "Validates business rules.")
}

Container_Boundary(model, "Model") {
    Component(proto, "Proto Models", "Protobuf", "Commands, Events, Queries.")
}

Rel(dispatch_svc, courier_agg, "Filters by status/capacity")
Rel(dispatch_svc, location_vo, "Calculates distance")
Rel(validation_svc, package_agg, "Checks package status")
Rel(validation_svc, courier_agg, "Checks courier availability")
Rel(package_agg, proto, "Uses proto types")
Rel(courier_agg, proto, "Uses proto types")

@enduml
```
