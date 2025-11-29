# Доменные сервисы (Domain Services)

## Обзор

Доменные сервисы содержат бизнес-логику, которая не принадлежит одному агрегату (entity) или требует работы с несколькими агрегатами одновременно. Они инкапсулируют важные бизнес-правила и операции домена.

## Созданные доменные сервисы

### 1. StockCartService (`cart/v1/stock_cart_service.go`)

**Назначение:** Обработка изменений стока и их влияния на корзины покупателей.

**Где применять:**
- В use case `handle_stock_change.go` - заменить текущую логику обработки изменений стока
- При получении событий об изменении стока из внешних систем

**Пример использования:**

```go
// В usecase/cart/handle_stock_change.go
func (uc *UC) HandleStockChange(ctx context.Context, goodId uuid.UUID, newQuantity uint32) error {
    if newQuantity > 0 {
        return nil
    }

    // Получить список клиентов с этим товаром в корзине
    customerIds := uc.goodsIndex.GetCustomersWithGood(goodId)
    
    // Использовать доменный сервис для обработки
    stockCartService := v1.NewStockCartService(uc.cartRepository)
    results, err := stockCartService.HandleStockDepletion(ctx, goodId, customerIds)
    if err != nil {
        return err
    }

    // Обработать результаты и отправить уведомления
    for _, result := range results {
        if result.Removed && uc.notifier != nil {
            uc.notifier.NotifyStockDepleted(result.CustomerID, result.GoodID)
        }
    }

    return nil
}
```

### 2. CartValidationService (`cart/v1/cart_validation_service.go`)

**Назначение:** Валидация операций с корзиной (добавление/удаление товаров).

**Где применять:**
- В use case `add.go` - перед добавлением товаров в корзину
- В use case `remove.go` - перед удалением товаров из корзины (опционально)
- При создании заказа из корзины

**Пример использования:**

```go
// В usecase/cart/add.go
func (uc *UC) Add(ctx context.Context, in *v1.CartState) error {
    // Валидация перед добавлением
    validationService := v1.NewCartValidationService(uc.stockChecker)
    validationResult := validationService.ValidateAddItems(ctx, in.GetItems())
    
    if !validationResult.Valid {
        // Вернуть ошибки валидации
        return fmt.Errorf("validation failed: %v", validationResult.Errors)
    }

    // Если есть предупреждения, залогировать их
    if len(validationResult.Warnings) > 0 {
        uc.log.Warn("Validation warnings", "warnings", validationResult.Warnings)
    }

    // Продолжить с добавлением товаров
    // ... остальная логика
}
```

### 3. OrderCreationService (`order/v1/order_creation_service.go`)

**Назначение:** Создание заказа из корзины с валидацией и применением бизнес-правил.

**Где применять:**
- В use case `order/create.go` - при создании заказа из корзины
- При конвертации корзины в заказ

**Пример использования:**

```go
// В usecase/order/create.go
func (uc *UC) CreateFromCart(ctx context.Context, customerId uuid.UUID) error {
    // Использовать доменный сервис для создания заказа
    orderCreationService := v1.NewOrderCreationService(
        uc.cartRepository,
        uc.stockChecker,
    )

    result, err := orderCreationService.CreateOrderFromCart(ctx, customerId)
    if err != nil {
        return fmt.Errorf("failed to create order from cart: %w", err)
    }

    // Если есть предупреждения, залогировать их
    if len(result.Warnings) > 0 {
        uc.log.Warn("Order creation warnings", "warnings", result.Warnings)
    }

    // Создать заказ с полученными товарами
    return uc.Create(ctx, result.OrderID, customerId, result.OrderItems)
}
```

## Принципы применения доменных сервисов

### Когда использовать доменные сервисы:

1. **Логика затрагивает несколько агрегатов**
   - Пример: обработка изменений стока влияет на множество корзин
   - Пример: создание заказа требует работы с корзиной и проверки стока

2. **Сложные бизнес-правила**
   - Пример: валидация корзины с проверкой стока, лимитов, ограничений
   - Пример: правила конвертации корзины в заказ

3. **Операции, не принадлежащие одному entity**
   - Пример: проверка доступности товара для добавления в корзину
   - Пример: расчет доступного количества товара с учетом резерва

### Когда НЕ использовать доменные сервисы:

1. **Простая логика одного агрегата** - должна быть в методах entity
   - Пример: `CartState.AddItem()` - простая операция добавления
   - Пример: `CartState.RemoveItem()` - простая операция удаления

2. **Инфраструктурные операции** - должны быть в infrastructure слое
   - Пример: сохранение в БД
   - Пример: отправка уведомлений через WebSocket

3. **Координация use cases** - должна быть в use case слое
   - Пример: вызов Temporal workflow
   - Пример: обновление индексов

## Миграция существующего кода

### Шаг 1: Рефакторинг handle_stock_change.go

**Текущая реализация:** Вся логика находится в use case.

**Целевая реализация:** 
- Доменная логика → `StockCartService`
- Координация и инфраструктура → use case

### Шаг 2: Добавление валидации в add.go

**Текущая реализация:** Нет валидации перед добавлением.

**Целевая реализация:**
- Валидация через `CartValidationService`
- Обработка ошибок и предупреждений

### Шаг 3: Улучшение создания заказа

**Текущая реализация:** Простое создание заказа без валидации корзины.

**Целевая реализация:**
- Использование `OrderCreationService` для валидации и конвертации
- Обработка ошибок и предупреждений

## Интерфейсы для dependency injection

Доменные сервисы используют интерфейсы для зависимостей:

- `CartRepository` - для работы с корзинами
- `StockChecker` - для проверки стока

Эти интерфейсы должны быть реализованы в infrastructure слое и переданы через dependency injection в use cases.

## Тестирование

Доменные сервисы легко тестировать, так как они:
- Не зависят от инфраструктуры (используют интерфейсы)
- Содержат чистую бизнес-логику
- Возвращают структурированные результаты

Пример теста:

```go
func TestStockCartService_HandleStockDepletion(t *testing.T) {
    mockRepo := &MockCartRepository{}
    service := v1.NewStockCartService(mockRepo)
    
    results, err := service.HandleStockDepletion(ctx, goodId, customerIds)
    // Проверки...
}
```

