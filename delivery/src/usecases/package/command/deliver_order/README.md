## Use Case: UC-3 Deliver Order

### Описание
Курьер подтверждает доставку заказа. Может быть успешной (доставлено) или неуспешной (не доставлено с указанием причины).

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

### Причины не доставки (NOT_DELIVERED)

- `CUSTOMER_NOT_AVAILABLE` - Клиент недоступен
- `WRONG_ADDRESS` - Неправильный адрес
- `CUSTOMER_REFUSED` - Клиент отказался от заказа
- `ACCESS_DENIED` - Нет доступа к адресу
- `PACKAGE_DAMAGED` - Посылка повреждена
- `OTHER` - Другая причина (требуется описание)

### Business Rules

**При успешной доставке (DELIVERED):**

1. Статус посылки меняется на `DELIVERED`
2. Устанавливается `delivered_at` timestamp
3. Посылка удаляется из пула посылок
4. Статус курьера меняется на `FREE`
5. Уменьшается текущая загрузка курьера
6. Увеличивается счетчик успешных доставок
7. Обновляется рейтинг курьера
8. Генерируется событие `PackageDelivered`
9. Отправляется уведомление в OMS о завершении доставки
10. Обновляется геолокация курьера

**При неуспешной доставке (NOT_DELIVERED):**

1. Статус посылки меняется на `NOT_DELIVERED`
2. Устанавливается причина не доставки
3. Статус меняется на `REQUIRES_HANDLING`
4. Посылка возвращается в пул или помечается для обработки диспетчером
5. Статус курьера меняется на `FREE`
6. Уменьшается текущая загрузка курьера
7. Генерируется событие `PackageNotDelivered`
8. Отправляется уведомление в OMS о проблеме
9. Создается задача для диспетчера
10. Обновляется геолокация курьера

### Error Cases

- `PACKAGE_NOT_FOUND`: Посылка не найдена
- `COURIER_NOT_ASSIGNED`: Посылка не назначена на этого курьера
- `INVALID_STATUS`: Некорректный статус доставки
- `REASON_REQUIRED`: Требуется указать причину при статусе NOT_DELIVERED
- `ALREADY_DELIVERED`: Посылка уже доставлена

