## Use Case: UC-2 Assign Order to Courier

### Описание
Назначение посылки на курьера. Может быть автоматическим (диспетчеризация) или ручным. При назначении курьер получает push-уведомление.

### Sequence Diagram

```mermaid
sequenceDiagram
    participant Dispatcher as Dispatcher/System
    participant Delivery as Delivery Service
    participant Dispatch as Dispatch Service
    participant CourierPool as Courier Pool
    participant Geolocation as Geolocation Service
    participant Courier as Courier (Mobile App)
    participant Push as Push Notification Service
  
    rect rgb(224, 236, 255)
        Dispatcher->>+Delivery: AssignOrder(AssignOrderRequest)
        Note right of Delivery: - Package ID<br>- Courier ID (optional)<br>- Auto-assign flag
    end
  
    rect rgb(224, 255, 239)
        alt Auto-assign (диспетчеризация)
            Delivery->>+Dispatch: FindNearestCourier(package)
            Dispatch->>+CourierPool: GetAvailableCouriers(zone, status)
            CourierPool-->>-Dispatch: List of available couriers
            Dispatch->>+Geolocation: GetCourierLocations(courier_ids)
            Geolocation-->>-Dispatch: Courier locations
            Dispatch->>Dispatch: Calculate distances
            Dispatch->>Dispatch: Filter by transport type
            Dispatch->>Dispatch: Filter by max distance
            Dispatch->>Dispatch: Filter by current load
            Dispatch->>Dispatch: Select nearest courier
            Dispatch-->>-Delivery: Selected courier
        else Manual assign
            Delivery->>+CourierPool: GetCourier(courier_id)
            CourierPool-->>-Delivery: Courier info
            Delivery->>Delivery: Validate courier availability
        end
    end
  
    rect rgb(255, 244, 224)
        Delivery->>Delivery: Assign package to courier
        Delivery->>Delivery: Update package status: ASSIGNED
        Delivery->>Delivery: Update courier status: BUSY
        Delivery->>Delivery: Increment courier load
        Delivery->>Delivery: Generate event: PackageAssigned
    end
  
    rect rgb(255, 230, 230)
        Delivery->>+Push: SendNotification(courier_id, notification)
        Note right of Push: - Package ID<br>- Pickup Address<br>- Delivery Address<br>- Customer Contacts
        Push->>Courier: Push notification
        Push-->>-Delivery: Notification sent
        Delivery-->>-Dispatcher: AssignOrderResponse
        Note left of Dispatcher: - Package ID<br>- Courier ID<br>- Assigned at timestamp
    end
```

### Request

```protobuf
message AssignOrderRequest {
  string package_id = 1;
  oneof assignment_type {
    string courier_id = 2; // Manual assignment
    bool auto_assign = 3;   // Auto dispatch
  }
}

message AssignOrderResponse {
  string package_id = 1;
  string courier_id = 2;
  google.protobuf.Timestamp assigned_at = 3;
  PackageStatus status = 4;
}
```

### Диспетчеризация (Auto-assign)

Алгоритм выбора ближайшего курьера:

1. **Фильтрация по зоне работы:**
   - Получить все свободные курьеры в зоне доставки
   - Статус: `FREE`
   - Зона работы пересекается с зоной доставки

2. **Расчет расстояния:**
   - Получить текущую геолокацию курьера через **Geolocation Service**: `GetCourierLocations(courier_ids)`
   - Рассчитать расстояние от курьера до точки забора
   - Использовать формулу Haversine для расчета расстояния
   - Если локация курьера недоступна, пропустить курьера из рассмотрения

3. **Фильтрация по возможностям:**
   - Тип транспорта курьера соответствует требованиям
   - Расстояние до точки забора ≤ максимальная дальность курьера
   - Текущая загрузка < максимальная загрузка

4. **Сортировка и выбор:**
   - Сортировать по расстоянию (ближайший первый)
   - Учитывать рейтинг курьера
   - Учитывать текущую загрузку (балансировка)
   - Выбрать оптимального курьера

### Business Rules

1. Курьер должен быть в статусе `FREE`
2. Курьер должен быть в рабочее время
3. Расстояние до точки забора не должно превышать максимальную дальность курьера
4. Текущая загрузка курьера должна быть меньше максимальной
5. При назначении статус посылки меняется на `ASSIGNED`
6. Статус курьера меняется на `BUSY`
7. Генерируется событие `PackageAssigned`
8. Отправляется push-уведомление курьеру

### Push Notification Content

```json
{
  "title": "Новый заказ назначен",
  "body": "Заказ #{{package_id}} готов к забору",
  "data": {
    "package_id": "uuid",
    "pickup_address": "ул. Примерная, 1",
    "pickup_coordinates": {
      "latitude": 55.7558,
      "longitude": 37.6173
    },
    "delivery_address": "ул. Доставки, 2",
    "delivery_coordinates": {
      "latitude": 55.7600,
      "longitude": 37.6200
    },
    "customer_phone": "+79001234567",
    "delivery_period": {
      "start": "2024-01-15T10:00:00Z",
      "end": "2024-01-15T12:00:00Z"
    }
  }
}
```

### Error Cases

- `PACKAGE_NOT_FOUND`: Посылка не найдена
- `COURIER_NOT_FOUND`: Курьер не найден
- `COURIER_NOT_AVAILABLE`: Курьер недоступен (занят/недоступен)
- `NO_AVAILABLE_COURIERS`: Нет доступных курьеров в зоне
- `INVALID_ASSIGNMENT`: Невозможно назначить (превышена загрузка, расстояние и т.д.)

