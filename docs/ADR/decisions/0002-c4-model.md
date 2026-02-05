# 2. C4 Model for Shop boundary context

Date: 2024-01-01

## Status

Accepted

## Context

The Shop Boundary consists of several critical services (Admin, BFF, OMS, Pricer, Feed, Email Subscription) integral to our system's operations 
related to goods and services management. Given the complex interactions and processes handled by these services, 
it is crucial to have a detailed and clear visualization of the architecture. 

The system uses:
- **Oathkeeper** as authentication proxy: validates JWT/session (Kratos), injects identity headers, forwards to BFF and Admin
- **BFF (Cosmo Router)** as the GraphQL API gateway for frontend requests (behind Oathkeeper)
- **Kafka** for asynchronous communication and event-driven operations between services
- **Temporal** for workflow orchestration (OMS, Email Subscription)
- **GraphQL** for API communication through BFF and OMS-GraphQL
- **gRPC** for inter-service communication

**Pricer** is called by OMS (and optionally BFF) for cart/order totals (taxes, discounts). **Delivery Service** (Delivery Boundary) receives orders from OMS and sends delivery status updates back.

The C4 model is renowned for its ability to effectively map and document software architecture, 
making it ideal for our needs to ensure clarity and cohesion across the system.

## Decision

We will apply the C4 model to detail the architecture of the Shop Boundary Context. This includes 
creating System Context, Container, and Component diagrams, and optionally, Class diagrams, 
for each service within the boundary.

Authentication flow: UI sends API and admin traffic to Oathkeeper; Oathkeeper validates JWT/session and forwards to BFF and Admin (see [ADR-0005](./0005-bff-behind-oathkeeper.md)).

## Consequences

By applying the C4 model to the Shop Boundary, we anticipate the following benefits:

+ **Enhanced Understanding:** All stakeholders, from developers to business analysts, will have a clearer understanding of the system architecture.
+ **Improved Communication:** Facilitates better discussions and decision-making regarding changes and enhancements to the system.
+ **Streamlined Development and Maintenance:** With a well-documented architecture, new team members can onboard more quickly, 
and ongoing maintenance can be managed more efficiently.


### C4

#### Level 1: System Context diagram

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

LAYOUT_WITH_LEGEND()

title System Context diagram for Shop Boundary with External Contexts

Person_Ext(customer, "Customer", "A customer using the online shop.")

System_Boundary(sbs, "Shop Boundary Context") {
    System(ui_service, "UI Service (Next.js)", "UI for the shop boundary")
    System(oathkeeper, "Oathkeeper", "Auth proxy: validates JWT/session, injects X-User-ID, forwards to BFF and Admin.")
    System(wundergraph_bff, "BFF (Cosmo Router)", "GraphQL API Gateway - Handles frontend requests via GraphQL and coordinates with backend services.")
    System(admin_service, "Admin Service", "Administers shop settings and user permissions.")
    System(pricer_service, "Pricer Service", "Calculates taxes and discounts for cart/order using OPA (Rego) policies.")
    System(oms_graphql, "OMS-GraphQL", "GraphQL API Bridge for orders via GraphQL API.")
    System(oms_temporal, "OMS (Temporal)", "Service for work with carts and orders using Temporal workflows.")
    System(email_subscription_service, "Email Subscription (Temporal)", "Handles email subscriptions and notifications.")
    System(feed_service, "Feed Service", "Cron job in Go, generates feeds every 24h and saves them to Minio.")
}

SystemQueue_Ext(kafka, "Kafka", "Message Queue for asynchronous communication and event-driven operations.")
System_Ext(minio_store, "Minio (S3-like block store)", "Stores generated feeds.")
System_Ext(temporal, "Temporal", "Workflow orchestration service.")

System_Boundary(bbs, "Billing Boundary") {
    System_Ext(billing_service, "Billing Service", "Manages billing and invoices.")
}

System_Boundary(dbs, "Delivery Boundary") {
    System_Ext(delivery_service, "Delivery Service", "Handles logistics: package/courier pools, dispatch, delivery status; receives orders from OMS, sends status updates back.")
}

Rel(customer, ui_service, "Accesses shop UI through", "HTTP/HTTPS")
Rel(ui_service, oathkeeper, "API and admin requests via", "HTTPS")
Rel(oathkeeper, wundergraph_bff, "Forwards authenticated requests to", "HTTP")
Rel(oathkeeper, admin_service, "Forwards authenticated requests to", "HTTP")
Rel(wundergraph_bff, oms_graphql, "Coordinates shopping cart and checkout via", "GraphQL")
Rel(wundergraph_bff, admin_service, "Admin service requests via", "GraphQL/REST")
Rel(wundergraph_bff, pricer_service, "Price calculation requests via", "gRPC")
Rel(oms_graphql, oms_temporal, "Communicates with OMS via", "gRPC")
Rel(oms_temporal, temporal, "Uses for workflow orchestration", "Temporal API")
Rel(oms_temporal, kafka, "Publishes/Consumes events via", "Kafka")
Rel(oms_temporal, pricer_service, "Gets cart/order totals from", "gRPC")
Rel(email_subscription_service, temporal, "Uses for workflow orchestration", "Temporal API")
Rel(email_subscription_service, kafka, "Consumes events via", "Kafka")
Rel(email_subscription_service, customer, "Sends email notifications to", "SMTP")
Rel(oms_temporal, billing_service, "Submits order details to", "gRPC/Kafka")
Rel(oms_temporal, delivery_service, "Sends order info for delivery to", "gRPC/Kafka")
Rel(delivery_service, oms_temporal, "Sends delivery status updates to", "gRPC/Webhook")
Rel(feed_service, oms_graphql, "Fetches data via", "GraphQL")
Rel(feed_service, minio_store, "Saves generated feeds to", "S3 API")

@enduml
```

#### Level 2: Container diagram

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()

title Container diagram for Shop Boundary Context

Person(customer, "Customer", "A customer interacts with the online shopping system.")
SystemQueue_Ext(kafka, "Kafka", "Message Queue for asynchronous communication and event-driven operations.")
Container_Ext(payment_gateway, "Payment Gateway", "External Service", "Securely processes payment transactions and handles financial data exchange.")
System_Ext(temporal, "Temporal", "Workflow orchestration service.")
System_Boundary(sbs, "Shop Boundary Context") {
    Container(ui_service, "UI Service (Next.js)", "Service", "User interface for customers interacting with the shop.")
    Container(oathkeeper, "Oathkeeper", "Service", "Auth proxy: validates JWT/session (Kratos), injects X-User-ID, forwards to BFF and Admin.")
    Container(wundergraph_bff, "BFF (Cosmo Router)", "Service", "GraphQL API Gateway - Handles frontend requests via GraphQL and coordinates with backend services.")
    Container(admin_service, "Admin Service", "Service", "Administers shop settings, manages user roles and permissions, and performs back-end configuration tasks.")
    Container(pricer_service, "Pricer Service", "Service", "Calculates taxes and discounts for cart/order using OPA (Rego) policies; called by OMS and optionally BFF.")
    Container(oms_graphql, "OMS-GraphQL", "Service", "GraphQL API Bridge for work with orders via GraphQL API.")
    Container(oms_temporal, "OMS (Temporal)", "Service", "Service for work with carts and orders using Temporal workflows and gRPC.")
    Container(email_subscription_service, "Email Subscription Service (Temporal)", "Service", "Manages email subscriptions and notifications using Temporal workflows.")
    Container(feed_service, "Feed Service", "Service", "Cron job in Go, generates feeds every 24h and saves them to Minio.")
    ContainerDb(shop_db, "Shop Database", "Database", "Central repository for storing all orders, carts, and administrative data.")
    Container(shop_cache, "Shop Cache Server", "Cache", "Improves performance by caching frequently accessed data such as product details and prices.")
    Container_Ext(minio_store, "Minio (S3-like block store)", "External Storage", "Stores generated feeds.")
}

System_Boundary(bbs, "Billing Boundary") {
    Container_Ext(billing_service, "Billing Service", "Service", "Manages billing and invoices.")
}

System_Boundary(dbs, "Delivery Boundary") {
    Container_Ext(delivery_service, "Delivery Service", "Service", "Handles logistics: package/courier pools, dispatch, delivery status; receives orders from OMS, sends status updates back.")
}

Rel_Down(customer, ui_service, "Submits requests to", "HTTP/HTTPS")
Rel(ui_service, oathkeeper, "API and admin requests via", "HTTPS")
Rel(oathkeeper, wundergraph_bff, "Forwards authenticated requests to", "HTTP")
Rel(oathkeeper, admin_service, "Forwards authenticated requests to", "HTTP")
Rel(wundergraph_bff, oms_graphql, "Coordinates order management via", "GraphQL")
Rel(wundergraph_bff, admin_service, "Routes administrative requests to", "GraphQL/REST")
Rel(wundergraph_bff, pricer_service, "Price calculation requests", "gRPC")
Rel(oms_graphql, oms_temporal, "Communicates with for order management", "gRPC")
Rel(oms_temporal, temporal, "Uses for workflow orchestration", "Temporal API")
Rel(oms_temporal, shop_db, "Reads and writes data", "SQL")
Rel(oms_temporal, shop_cache, "Utilizes for faster data retrieval", "Redis")
Rel(oms_temporal, kafka, "Publishes/Consumes events", "Kafka")
Rel(oms_temporal, billing_service, "Submits order details to", "gRPC/Kafka")
Rel(oms_temporal, delivery_service, "Sends order info for delivery to", "gRPC/Kafka")
Rel(delivery_service, oms_temporal, "Sends delivery status updates to", "gRPC/Webhook")
Rel(oms_temporal, pricer_service, "Gets cart/order totals from", "gRPC")
Rel(oms_temporal, payment_gateway, "Connects for payment processing", "API")
Rel(admin_service, shop_db, "Reads and writes data", "SQL")
Rel(email_subscription_service, temporal, "Uses for workflow orchestration", "Temporal API")
Rel(email_subscription_service, kafka, "Consumes events", "Kafka")
Rel(feed_service, oms_graphql, "Fetches data via", "GraphQL")
Rel(feed_service, minio_store, "Saves generated feeds to", "S3 API")

@enduml
```
