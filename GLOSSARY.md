## Terms

### Core Domain

- **Good**: A product or item available for purchase in the shop catalog. This is the primary term used throughout the system to refer to merchandise/trade goods. Managed by Admin Service, displayed in UI Service, and referenced in orders and carts.
  - **Note**: In UI layer, the TypeScript type `Product` is used for compatibility with GraphQL/Shopify API, but it represents a `Good` from the domain perspective.
  - **Note**: `Item` in Merch Service (deprecated) refers to SKU-based inventory units, which is a different level of abstraction.

- **Cart Item**: An entry in a shopping cart that references a Good with quantity and price information.
- **Order Item**: An entry in an order that references a Good with quantity, price, and order-specific details.

- **Merch**: Short for merchandise, referring to the goods that are being delivered.

### Delivery

- **Geolocation**: The use of technology to determine the geographical location of a user or device, crucial for delivery tracking and routing.
- **Delivery Tracking**: A system for monitoring the progress and location of an order from the point of sale to delivery.
- **Support**: Customer service functions related to the delivery process, addressing inquiries and issues.
- **Routing**: The process of determining the most efficient path for delivery.
- **Order Management**: The administration of business processes related to orders for goods or services.
- **Customer Feedback**: Responses and reviews from customers regarding their delivery experience.