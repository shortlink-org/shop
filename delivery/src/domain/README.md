# Delivery Domain Models

## Структура

```
src/domain/
├── delivery/            # Protobuf определения
│   ├── common/v1/       # Общие типы и enums
│   │   └── common.proto
│   ├── commands/v1/     # Команды (CQRS)
│   │   ├── commands.proto
│   │   └── responses.proto
│   ├── events/v1/       # События (Event Sourcing)
│   │   └── events.proto
│   └── queries/v1/      # Запросы (Queries)
│       └── queries.proto
├── vo/                  # Rust value objects
│   └── location.rs      # Location value object
└── mod.rs              # Модуль домена
```

## Общие типы (common/v1)

- `Address` - Адрес с координатами
- `Location` - GPS локация
- `DeliveryPeriod` - Период доставки
- `PackageInfo` - Характеристики посылки
- `WorkHours` - Рабочие часы курьера
- Enums: `Priority`, `PackageStatus`, `CourierStatus`, `TransportType`, `DeliveryStatus`, `NotDeliveredReason`

## Команды (commands/v1)

### AcceptOrderCommand
Принять заказ от OMS для доставки.

### AssignOrderCommand
Назначить посылку на курьера (автоматически или вручную).

### DeliverOrderCommand
Подтвердить доставку курьером.

### UpdateCourierLocationCommand
Обновить геолокацию курьера.

### RegisterCourierCommand
Зарегистрировать нового курьера.

### GetPackagePoolCommand
Получить список посылок с фильтрацией.

### GetCourierPoolCommand
Получить список курьеров с фильтрацией.

## События (events/v1)

### PackageAcceptedEvent
Посылка принята в пул.

### PackageAssignedEvent
Посылка назначена на курьера → Push-уведомление.

### PackageInTransitEvent
Посылка в пути.

### PackageDeliveredEvent
Посылка доставлена → Уведомление OMS.

### PackageNotDeliveredEvent
Посылка не доставлена → Уведомление OMS и диспетчера.

### CourierRegisteredEvent
Курьер зарегистрирован.

### CourierLocationUpdatedEvent
Обновлена геолокация курьера.

### CourierStatusChangedEvent
Изменен статус курьера.

### PackageRequiresHandlingEvent
Посылка требует обработки диспетчером.

## Использование

### Генерация Rust кода

```bash
# Установить buf (если еще не установлен)
# https://buf.build/docs/installation

# Установить prost плагины для buf
buf mod update

# Сгенерировать код
buf generate
```

### Зависимости в Cargo.toml

```toml
[dependencies]
prost = "0.12"
prost-types = "0.12"
pbjson-types = "0.6"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
```

### Пример использования команды

```rust
use domain::delivery::commands::v1::AcceptOrderCommand;
use domain::delivery::common::v1::{Address, Priority};

let cmd = AcceptOrderCommand {
    order_id: "order-123".to_string(),
    customer_id: "customer-456".to_string(),
    pickup_address: Some(Address {
        street: "ул. Примерная, 1".to_string(),
        city: "Москва".to_string(),
        latitude: 55.7558,
        longitude: 37.6173,
        ..Default::default()
    }),
    priority: Priority::Normal as i32,
    ..Default::default()
};
```

### Пример обработки события

```rust
use domain::delivery::events::v1::PackageAssignedEvent;

fn handle_package_assigned(event: PackageAssignedEvent) {
    // Отправить push-уведомление курьеру
    send_push_notification(&event.courier_id, &event);
}
```

### Структура модулей после генерации

```
src/
└── domain/
    ├── delivery/              # Сгенерированный код из proto
    │   ├── common::v1::*      # Address, Location, Priority, etc.
    │   ├── commands::v1::*    # Commands
    │   ├── events::v1::*      # Events
    │   └── queries::v1::*     # Queries
    ├── vo::*                  # Value objects (vo::location::*)
    └── mod.rs
```

