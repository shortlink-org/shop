# Delivery GraphQL Subgraph

GraphQL subgraph для Delivery service. Проксирует gRPC API в GraphQL используя [Tailcall](https://tailcall.run/).

## Архитектура

```
admin-ui → BFF (Cosmo Router) → delivery-graphql → Delivery Service (gRPC)
```

## Запуск

```bash
# Установка зависимостей
pnpm install

# Запуск (порт 8102)
pnpm start

# Разработка с hot-reload
pnpm dev

# Проверка конфигурации
pnpm check
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DELIVERY_GRPC_URL` | URL Delivery gRPC service | `http://localhost:50051` |

## GraphQL API

### Queries

- `couriers(filter, page, pageSize)` — список курьеров с фильтрацией
- `courier(id)` — получить курьера по ID
- `courierDeliveries(courierId, limit)` — история доставок курьера

### Mutations

- `registerCourier(input)` — регистрация курьера
- `activateCourier(id)` — активировать курьера
- `deactivateCourier(id, reason)` — деактивировать курьера
- `archiveCourier(id, reason)` — архивировать курьера
- `updateCourierContact(id, input)` — обновить контакты
- `updateCourierSchedule(id, input)` — обновить расписание
- `changeCourierTransport(id, transportType)` — сменить транспорт

## Ссылки

- [Tailcall Documentation](https://tailcall.run/docs/)
- Delivery Service proto: `delivery/src/infrastructure/rpc/delivery.proto`
