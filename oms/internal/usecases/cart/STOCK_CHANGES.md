# Обработка событий изменения остатков товаров (Stock Changes)

## Описание

Реализована функциональность автоматического удаления товаров из корзин при обнулении остатков на складе. При получении события `stockchanges` о том, что товар закончился (остаток = 0), система:

1. Находит все корзины, содержащие этот товар
2. Удаляет товар из найденных корзин
3. Отправляет уведомление пользователям через WebSocket

## Компоненты

### 1. Индекс товаров в корзинах (`internal/infrastructure/index/cart_goods_index.go`)

Индекс для быстрого поиска всех корзин, содержащих определенный товар. Индекс обновляется автоматически при:
- Добавлении товара в корзину (`Add`)
- Удалении товара из корзины (`Remove`)
- Очистке корзины (`Reset`)

### 2. Обработчик событий (`internal/usecases/cart/handle_stock_change.go`)

Метод `HandleStockChange` обрабатывает события изменения остатков:
- Если остаток > 0 - никаких действий не требуется
- Если остаток = 0 - товар удаляется из всех корзин, где он присутствует

### 3. HTTP endpoint для получения событий (`internal/infrastructure/http/stock_event.go`)

HTTP endpoint `POST /stock-changes` для получения событий от внешних сервисов (например, от Inventory Service через Kafka).

**Формат запроса:**
```json
{
  "good_id": "uuid",
  "old_quantity": 10,
  "new_quantity": 0
}
```

### 4. WebSocket сервер для уведомлений (`internal/infrastructure/websocket/`)

WebSocket сервер для отправки уведомлений в UI:
- Подключение: `WS /ws?customer_id={uuid}`
- Уведомление отправляется в формате JSON:
```json
{
  "type": "stock_depleted",
  "message": "Товар закончился на складе и был удален из вашей корзины",
  "data": {
    "good_id": "uuid",
    "message": "Товар закончился на складе и был удален из вашей корзины"
  }
}
```

## Интеграция

### 1. Инициализация WebSocket Notifier

В `internal/di/wire.go` нужно добавить инициализацию WebSocket notifier и передать его в cart service:

```go
// Создать notifier
notifier := websocket.NewNotifier(log)

// Установить в cart service
cartService.SetNotifier(notifier)
```

### 2. Регистрация HTTP endpoints

Добавить регистрацию HTTP endpoints для:
- Получения событий stockchanges: `POST /stock-changes`
- WebSocket подключений: `WS /ws?customer_id={uuid}`

### 3. Подписка на события

Настроить подписку на события `stockchanges` от Inventory Service (через Kafka). При получении события вызывать HTTP endpoint `/stock-changes` или напрямую метод `HandleStockChange`.

## Использование

### Отправка события stockchanges

```bash
curl -X POST http://localhost:50051/stock-changes \
  -H "Content-Type: application/json" \
  -d '{
    "good_id": "123e4567-e89b-12d3-a456-426614174000",
    "old_quantity": 10,
    "new_quantity": 0
  }'
```

### Подключение к WebSocket

```javascript
const customerId = '123e4567-e89b-12d3-a456-426614174000';
const ws = new WebSocket(`ws://localhost:50051/ws?customer_id=${customerId}`);

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  if (notification.type === 'stock_depleted') {
    console.log('Товар закончился:', notification.data.good_id);
    // Обновить UI корзины
  }
};
```

## Примечания

- Индекс товаров хранится в памяти. Для production рекомендуется использовать Redis или другую персистентную БД
- WebSocket соединения управляются автоматически (ping/pong, cleanup при отключении)
- При отсутствии WebSocket соединения для пользователя, уведомление не отправляется (не является ошибкой)

