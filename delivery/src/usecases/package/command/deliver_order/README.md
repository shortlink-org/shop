## Use Case: UC-3 Deliver Order

### Description

Courier confirms order delivery. Can be successful (delivered) or unsuccessful (not delivered with reason specified).

### Sequence Diagram

```mermaid
sequenceDiagram
    participant Courier as Courier (Mobile App)
    participant Delivery as Delivery Service
    participant PackagePool as Package Pool
    participant CourierPool as Courier Pool
    participant OMS as OMS Service
    participant Dispatcher as Dispatcher Service
  
    rect rgb(224, 236, 255)
        Courier->>+Delivery: DeliverOrder(DeliverOrderRequest)
        Note right of Delivery: - Package ID<br>- Courier ID<br>- Status: DELIVERED/NOT_DELIVERED<br>- Reason (if not delivered)<br>- Photo (optional)<br>- Customer signature (optional)
    end
  
    rect rgb(224, 255, 239)
        Delivery->>Delivery: Validate request
        Delivery->>+PackagePool: GetPackage(package_id)
        PackagePool-->>-Delivery: Package info
        Delivery->>Delivery: Verify courier assignment
    end
  
    rect rgb(255, 244, 224)
        alt Status: DELIVERED
            Delivery->>Delivery: Update package status: DELIVERED
            Delivery->>Delivery: Set delivered_at timestamp
            Delivery->>Delivery: Remove from package pool
            Delivery->>+CourierPool: UpdateCourier(courier_id)
            CourierPool->>CourierPool: Set status: FREE
            CourierPool->>CourierPool: Decrement load
            CourierPool->>CourierPool: Increment successful deliveries
            CourierPool->>CourierPool: Update rating
            CourierPool-->>-Delivery: Courier updated
            Delivery->>Delivery: Generate event: PackageDelivered
            Delivery->>+OMS: NotifyDeliveryCompleted(package_id)
            OMS-->>-Delivery: Acknowledged
        else Status: NOT_DELIVERED
            Delivery->>Delivery: Update package status: NOT_DELIVERED
            Delivery->>Delivery: Set not_delivered_reason
            Delivery->>Delivery: Set status: REQUIRES_HANDLING
            Delivery->>+CourierPool: UpdateCourier(courier_id)
            CourierPool->>CourierPool: Set status: FREE
            CourierPool->>CourierPool: Decrement load
            CourierPool-->>-Delivery: Courier updated
            Delivery->>Delivery: Generate event: PackageNotDelivered
            Delivery->>+OMS: NotifyDeliveryFailed(package_id, reason)
            OMS-->>-Delivery: Acknowledged
            Delivery->>+Dispatcher: CreateTask(package_id, reason)
            Dispatcher-->>-Delivery: Task created
        end
    end
  
    rect rgb(255, 230, 230)
        Delivery->>Geolocation: UpdateCourierLocation(courier_id, location)
        Delivery-->>-Courier: DeliverOrderResponse
        Note left of Courier: - Package ID<br>- Status<br>- Updated at
    end
```

### Request

```protobuf
message DeliverOrderRequest {
  string package_id = 1;
  string courier_id = 2;
  DeliveryStatus status = 3;
  string reason = 4; // Required if status is NOT_DELIVERED
  bytes photo = 5; // Optional: delivery confirmation photo
  bytes customer_signature = 6; // Optional: customer signature
  Location current_location = 7; // Courier's current location after delivery
}

enum DeliveryStatus {
  DELIVERY_STATUS_UNKNOWN = 0;
  DELIVERY_STATUS_DELIVERED = 1;
  DELIVERY_STATUS_NOT_DELIVERED = 2;
}

message Location {
  double latitude = 1;
  double longitude = 2;
  double accuracy = 3; // meters
  google.protobuf.Timestamp timestamp = 4;
}
```

### Response

```protobuf
message DeliverOrderResponse {
  string package_id = 1;
  PackageStatus status = 2;
  google.protobuf.Timestamp updated_at = 3;
}
```

### NOT_DELIVERED Reasons

- `CUSTOMER_NOT_AVAILABLE` - Customer not available
- `WRONG_ADDRESS` - Wrong address
- `CUSTOMER_REFUSED` - Customer refused the order
- `ACCESS_DENIED` - No access to address
- `PACKAGE_DAMAGED` - Package is damaged
- `OTHER` - Other reason (description required)

### Business Rules

**On successful delivery (DELIVERED):**

1. Package status changes to `DELIVERED`
2. `delivered_at` timestamp is set
3. Package is removed from package pool
4. Courier status changes to `FREE`
5. Courier's current load is decremented
6. Successful deliveries counter is incremented
7. Courier rating is updated
8. `PackageDelivered` event is generated
9. Notification is sent to OMS about delivery completion
10. Courier location is updated

**On unsuccessful delivery (NOT_DELIVERED):**

1. Package status changes to `NOT_DELIVERED`
2. Not delivered reason is set
3. Status changes to `REQUIRES_HANDLING`
4. Package is returned to pool or marked for dispatcher handling
5. Courier status changes to `FREE`
6. Courier's current load is decremented
7. `PackageNotDelivered` event is generated
8. Notification is sent to OMS about the problem
9. Task is created for dispatcher
10. Courier location is updated

### Error Cases

- `PACKAGE_NOT_FOUND`: Package not found
- `COURIER_NOT_ASSIGNED`: Package is not assigned to this courier
- `INVALID_STATUS`: Invalid delivery status
- `REASON_REQUIRED`: Reason required when status is NOT_DELIVERED
- `ALREADY_DELIVERED`: Package already delivered
