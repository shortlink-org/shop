# Geolocation Service Use Cases

## Обзор

Geolocation Service предоставляет функциональность для работы с геолокацией курьеров в Delivery Boundary.

## Use Cases

### UC-1: Save Location
**Описание:** Сохранение текущей геолокации курьера. Обновляет текущую позицию и добавляет запись в историю.

**См.:** [command/save_location/README.md](./command/save_location/README.md)

### UC-2: Get Courier Locations
**Описание:** Получение текущих геолокаций одного или нескольких курьеров одновременно. Поддерживает запрос как одного курьера (для простых случаев), так и множества курьеров (для диспетчеризации и массовых операций).

**См.:** [query/get_courier_locations/README.md](./query/get_courier_locations/README.md)

**Примечание:** UC-2 заменил отдельный use case для одного курьера, так как это подмножество общего случая. Для получения локации одного курьера просто передайте массив с одним элементом.

### UC-3: Get Location History
**Описание:** Получение истории геолокаций курьера за указанный период времени.

**См.:** [query/get_location_history/README.md](./query/get_location_history/README.md)

### UC-4: Create Geofence
**Описание:** Создание геозоны (geofence) — виртуальной географической границы для отслеживания входа/выхода курьеров. Геозона может быть кругом, многоугольником или прямоугольником. При пересечении границы могут выполняться автоматические действия (уведомления, события, изменение статуса).

**См.:** [command/create_geofence/README.md](./command/create_geofence/README.md)

**Что такое Geofence?** Геозона — это виртуальная граница на карте, которая отслеживает, когда курьер входит или выходит из определенной области. Используется для:
- Определения рабочих зон курьеров
- Автоматического уведомления о прибытии к точке доставки
- Контроля выхода из рабочей зоны
- Аналитики времени нахождения в зонах

### UC-5: Check Geofence (TBD)
**Описание:** Проверка нахождения курьера внутри геозоны.

**Статус:** Запланировано

## Интеграция с Delivery Service

Delivery Service использует Geolocation Service для:

1. **Диспетчеризация (Assign Order):**
   - `GetCourierLocations(courier_ids)` - получение локаций всех доступных курьеров
   - Расчет расстояний до точки забора
   - Выбор ближайшего курьера

2. **Обновление локации:**
   - `SaveLocation(courier_id, location)` - сохранение локации от мобильного приложения курьера

3. **Отслеживание маршрутов:**
   - `GetLocationHistory(courier_id, time_range)` - получение истории перемещений
   - Анализ эффективности доставок

4. **Мониторинг:**
   - `GetCourierLocations([courier_id])` - получение текущей позиции для отслеживания (передать массив с одним элементом)

## Модель данных

### Location

```protobuf
message Location {
  double latitude = 1;
  double longitude = 2;
  double accuracy = 3; // meters
  google.protobuf.Timestamp timestamp = 4;
  optional double speed = 5; // km/h (optional)
  optional double heading = 6; // degrees 0-360 (optional)
}
```

### LocationHistoryEntry

```protobuf
message LocationHistoryEntry {
  Location location = 1;
  google.protobuf.Timestamp recorded_at = 2;
}
```

## База данных

### Текущая локация
- Таблица: `courier_current_locations`
- Ключ: `courier_id` (unique)
- Индексы: по координатам для пространственных запросов

### История локаций
- Таблица: `courier_location_history`
- Ключ: `(courier_id, timestamp)`
- Партиционирование: по дате для оптимизации
- TTL: автоматическое удаление старых записей (например, старше 30 дней)

## Производительность

**Рекомендации по частоте обновления:**
- В пути: каждые 30-60 секунд
- Стоит на месте: каждые 5 минут
- При доставке: каждые 10 секунд

**Оптимизация:**
- Индексация по координатам (PostGIS, spatial index)
- Партиционирование истории по времени
- Кэширование текущих локаций в Redis
- Батчинг для массовых операций

## Связанные документы

- [Geolocation Service README](../../README.md)
- [Delivery Service Use Cases](../../../delivery/src/usecases/README.md)

