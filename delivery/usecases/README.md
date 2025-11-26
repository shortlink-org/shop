# Delivery Service Use Cases

## Обзор

Данный документ описывает все use cases для Delivery Service - сервиса управления доставкой заказов.

## Use Cases

### UC-1: Accept Order
**Принять заказ от OMS**

Прием заказа от OMS для доставки. Заказ добавляется в пул посылок со статусом "Принят к доставке".

- [Детали](./accept_order/README.md)

### UC-2: Assign Order to Courier
**Назначить заказ на курьера**

Назначение посылки на курьера. Может быть автоматическим (диспетчеризация) или ручным. При назначении курьер получает push-уведомление.

- [Детали](./assign_order/README.md)

### UC-3: Deliver Order
**Доставить заказ**

Курьер подтверждает доставку заказа. Может быть успешной (доставлено) или неуспешной (не доставлено с указанием причины).

- [Детали](./deliver_order/README.md)

### UC-4: Update Courier Location
**Обновить геолокацию курьера**

Обновление геолокации курьера в реальном времени. Используется для отслеживания перемещений и диспетчеризации.

- [Детали](./update_courier_location/README.md)

### UC-5: Register Courier
**Зарегистрировать курьера**

Регистрация нового курьера в системе. Курьер получает учетные данные и может начать принимать заказы.

- [Детали](./register_courier/README.md)

### UC-6: Get Package Pool
**Получить пул посылок**

Получение списка посылок в пуле с возможностью фильтрации по статусу, приоритету, региону и другим параметрам.

- [Детали](./get_package_pool/README.md)

### UC-7: Get Courier Pool
**Получить пул курьеров**

Получение списка курьеров с возможностью фильтрации по статусу, типу транспорта, зоне работы и другим параметрам.

- [Детали](./get_courier_pool/README.md)

## Workflow доставки

```
Принят (ACCEPTED) 
  ↓
В пуле (IN_POOL) 
  ↓
Назначен курьеру (ASSIGNED) 
  ↓
В пути (IN_TRANSIT) 
  ↓
Доставлено (DELIVERED) / Не доставлено (NOT_DELIVERED)
```

**При статусе "Не доставлено":**
```
Не доставлено (NOT_DELIVERED) 
  ↓
Требует обработки (REQUIRES_HANDLING) 
  ↓
(возврат в пул или отмена)
```

## События

- `PackageAccepted` - посылка принята в пул
- `PackageAssigned` - посылка назначена на курьера → **Push-уведомление курьеру**
- `PackageInTransit` - посылка в пути
- `PackageDelivered` - посылка доставлена → Уведомление OMS
- `PackageNotDelivered` - посылка не доставлена → Уведомление OMS и диспетчера
- `CourierRegistered` - курьер зарегистрирован
- `CourierLocationUpdated` - обновлена геолокация курьера

## Статусы посылок

- `ACCEPTED` - Принят в пул
- `IN_POOL` - В пуле, ожидает назначения
- `ASSIGNED` - Назначен курьеру
- `IN_TRANSIT` - В пути
- `DELIVERED` - Доставлен
- `NOT_DELIVERED` - Не доставлен
- `REQUIRES_HANDLING` - Требует обработки

## Статусы курьеров

- `UNAVAILABLE` - Недоступен (начальный статус)
- `FREE` - Свободен, готов к работе
- `BUSY` - Занят, выполняет доставку

## Типы транспорта

- `WALKING` - Пеший
- `BICYCLE` - Велосипед
- `MOTORCYCLE` - Мотоцикл
- `CAR` - Автомобиль

## Приоритеты

- `NORMAL` - Обычный
- `URGENT` - Срочный

## Интеграция с Geolocation Service

Delivery Service интегрируется с **Geolocation Service** для отслеживания местоположения курьеров в реальном времени.

### Использование Geolocation Service

**UC-2: Assign Order (Диспетчеризация)**
- Получение текущих локаций курьеров: `GetCourierLocations(courier_ids)`
- Используется для расчета расстояния до точки забора
- Выбор ближайшего курьера на основе геолокации

**UC-4: Update Courier Location**
- Сохранение локации курьера: `SaveLocation(courier_id, location)`
- Обновление текущей локации
- Добавление в историю локаций
- Уведомление Dispatch Service об обновлении

**UC-6: Get Courier Pool**
- Опциональное получение текущих локаций: `GetCurrentLocations(courier_ids)`
- Используется при `include_location = true`

**UC-3: Deliver Order**
- Обновление локации курьера после доставки
- Сохранение финальной позиции

### API Geolocation Service

```protobuf
// Сохранение локации курьера
service GeolocationService {
  rpc SaveLocation(SaveLocationRequest) returns (SaveLocationResponse);
  rpc GetCourierLocation(GetCourierLocationRequest) returns (GetCourierLocationResponse);
  rpc GetCourierLocations(GetCourierLocationsRequest) returns (GetCourierLocationsResponse);
  rpc GetLocationHistory(GetLocationHistoryRequest) returns (GetLocationHistoryResponse);
}

message SaveLocationRequest {
  string courier_id = 1;
  Location location = 2;
}

message GetCourierLocationsRequest {
  repeated string courier_ids = 1;
}

message GetCourierLocationsResponse {
  map<string, Location> locations = 1; // courier_id -> location
}
```

### Зависимости

- **Delivery Service** → **Geolocation Service** (gRPC)
- Геолокация используется для:
  - Диспетчеризации (выбор ближайшего курьера)
  - Отслеживания маршрутов курьеров
  - Мониторинга доставок в реальном времени
  - Оптимизации маршрутов

### См. также

- [Geolocation Service](../geolocation/README.md)

