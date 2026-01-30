# 2. C4 System Architecture

Date: 2025-01-30

## Status

Accepted

## Context

Define the Courier Emulation Service architecture using the C4 model.

## Decision

Use C4 diagrams to document the architecture at different levels of abstraction.

## Consequences

- Clear visualization of system boundaries and interactions
- Understandable for both technical and non-technical stakeholders
- Simplifies onboarding and architectural discussions

### System Context Diagram

```plantuml
@startuml C4_Context_Diagram_for_Courier_Emulation
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

LAYOUT_WITH_LEGEND()

title System Context Diagram for Courier Emulation Service

Person(tester, "QA Engineer", "Runs test scenarios")
Person(developer, "Developer", "Debugs integrations")

System_Boundary(shop, "Shop Boundary") {
    System(emulation, "Courier Emulation Service", "Emulates courier behavior: movement, locations, statuses")
    System(delivery, "Delivery Service", "Manages deliveries and assignments")
    System(geolocation, "Geolocation Service", "Tracks courier locations")
}

System_Ext(osrm, "OSRM", "Open Source Routing Machine — route generation")

Rel(tester, emulation, "Runs scenarios", "gRPC")
Rel(developer, emulation, "Manages emulation", "gRPC")
Rel(emulation, delivery, "Receives assignments, confirms delivery", "gRPC")
Rel(emulation, geolocation, "Updates courier locations", "gRPC")
Rel(emulation, osrm, "Generates routes", "HTTP")

@enduml
```

### Container Diagram

```plantuml
@startuml C4_Container_Diagram_for_Courier_Emulation
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()

title Container Diagram for Courier Emulation Service

Person(tester, "QA Engineer", "Runs tests")

System_Boundary(emulation, "Courier Emulation Service") {
    Container(api, "gRPC API", "Go/gRPC", "Courier emulation management")
    Container(simulator, "Courier Simulator", "Go", "Courier movement and action emulation")
    Container(route_gen, "Route Generator", "Go", "Route generation and caching")
    Container(event_emitter, "Event Emitter", "Go", "Event publishing to delivery/geolocation")
    ContainerDb(route_cache, "Route Cache", "Redis/File", "Generated routes cache")
}

System_Ext(osrm, "OSRM", "Routing engine")
System_Ext(delivery, "Delivery Service", "Assignments")
System_Ext(geolocation, "Geolocation Service", "Locations")

Rel(tester, api, "Manages emulation", "gRPC")
Rel(api, simulator, "Starts simulation")
Rel(simulator, route_gen, "Requests route")
Rel(simulator, event_emitter, "Sends events")
Rel(route_gen, osrm, "Generates route", "HTTP")
Rel(route_gen, route_cache, "Caches routes")
Rel(event_emitter, delivery, "Delivery statuses", "gRPC")
Rel(event_emitter, geolocation, "Location updates", "gRPC")

@enduml
```

### Component Diagram (Courier Simulator)

```plantuml
@startuml C4_Component_Diagram_for_Courier_Simulator
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram for Courier Simulator

Container_Boundary(simulator, "Courier Simulator") {
    Component(courier_actor, "CourierActor", "Go", "Stateful emulated courier actor")
    Component(movement_engine, "MovementEngine", "Go", "Movement calculation along polyline route")
    Component(behavior_strategy, "BehaviorStrategy", "Go", "Behavior strategies: always_accept, realistic, random_failures")
    Component(clock, "SimulationClock", "Go", "Simulation time management (real-time / accelerated)")
}

Container_Boundary(route_gen, "Route Generator") {
    Component(osrm_client, "OSRMClient", "Go", "HTTP client for OSRM API")
    Component(polyline_decoder, "PolylineDecoder", "Go", "Polyline decoding to coordinates")
    Component(route_store, "RouteStore", "Go", "Route storage and selection")
}

Rel(courier_actor, movement_engine, "Requests next position")
Rel(courier_actor, behavior_strategy, "Determines event response")
Rel(movement_engine, clock, "Considers simulation speed")
Rel(movement_engine, polyline_decoder, "Gets coordinates")
Rel(route_store, osrm_client, "Generates new routes")
Rel(route_store, polyline_decoder, "Decodes polyline")

@enduml
```
