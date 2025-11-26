# Delivery Service

Delivery Service - сервис управления доставкой заказов, реализованный на Rust.

## Структура проекта

```
delivery/
├── src/                    # Rust исходный код
│   ├── main.rs            # Точка входа
│   └── domain/            # Сгенерированный код из proto (не в git)
├── domain/                 # Protobuf определения
│   ├── common/v1/         # Общие типы
│   ├── commands/v1/       # Команды (CQRS)
│   └── events/v1/         # События (Event Sourcing)
├── usecases/              # Use cases документация
├── ops/                   # Операционные файлы
│   └── proto/             # Конфигурация генерации proto
├── Cargo.toml             # Rust зависимости
├── build.rs               # Скрипт генерации proto кода
└── buf.yaml               # Конфигурация buf для proto

```

## Генерация кода из proto

```bash
# Установить зависимости
cargo build

# Или использовать buf напрямую
buf generate
```

Код будет сгенерирован в `src/domain/` при сборке проекта.

## Use Cases

См. [usecases/README.md](./usecases/README.md) для полного списка use cases.

## Domain Models

См. [domain/README.md](./domain/README.md) для описания proto файлов и моделей данных.

## Зависимости

- **prost** - Protobuf компилятор для Rust
- **tonic** - gRPC фреймворк для Rust
- **tokio** - Асинхронный runtime

## Разработка

```bash
# Сборка проекта
cargo build

# Запуск тестов
cargo test

# Проверка кода
cargo clippy

# Форматирование
cargo fmt
```

