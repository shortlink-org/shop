# 5. Надёжная публикация доменных событий через Watermill Outbox (Forwarder)

Дата: 2026-02-03

## Статус

Предложено (Proposed)

## Контекст

Доменные события в OMS сейчас публикуются **после** коммита транзакции (хендлеры создания/отмены заказа, оформления заказа, обновления доставки). Схема:

1. Начало транзакции → сохранение агрегата → **Commit** → публикация событий (например в `InMemoryPublisher`).

Если публикация падает после коммита, события теряются, повторной отправки нет — код только логирует и продолжает работу. Это классическая проблема «ненадёжной публикации событий».

В OMS уже используется Watermill для потребления событий доставки (Kafka). Для надёжной публикации (outbox) используем **go-sdk/cqrs**.

### Выводы

**go-sdk/cqrs** ([shortlink-org/go-sdk/cqrs](https://github.com/shortlink-org/go-sdk/tree/main/cqrs)) даёт всё нужное для надёжной публикации:
- **EventBus** с `bus.WithOutbox(OutboxConfig{ ... })` и демоном `RunForwarder` / `CloseForwarder` (под капотом watermill/components/forwarder).
- Канонические имена сообщений, proto/JSON marshaler, трассировка.
- **Транзакционная публикация**: **`Publish(ctx, evt, bus.WithPublisher(txPublisher))`** — передаём publisher от текущей транзакции (например watermill-sql от `pgx.Tx`), запись в outbox попадает в ту же транзакцию, что и сохранение агрегата.

## Решение

Использовать **go-sdk/cqrs** для надёжной публикации доменных событий (EventBus + Outbox Forwarder).

**Стек**
- **go-sdk/cqrs** `EventBus` с `bus.WithOutbox(...)` и `RunForwarder` / `CloseForwarder`.
- Outbox-бэкенд: watermill-sql (publisher от `pgx.Tx` при публикации в транзакции UoW).
- Целевой транспорт Forwarder и существующие консьюмеры: **watermill-kafka**.

**Поток**
- В командных хендлерах в **той же** транзакции UoW (до `Commit`): сохранить агрегат, по каждому доменному событию вызвать **`eventBus.Publish(ctx, evt, bus.WithPublisher(txPublisher))`**, где `txPublisher` — watermill-sql publisher от `pgx.Tx` из контекста (при необходимости обёрнут в forwarder.NewPublisher). При `Commit` коммитятся и агрегат, и строки outbox. Forwarder (тот же процесс или отдельный) читает из outbox в Postgres и пересылает в Kafka.

**In-process подписчики (Temporal)**  
Оставить **InMemoryPublisher**: после `Commit` по-прежнему слать события в него, чтобы Temporal `OrderEventSubscriber` работал без подписки на Kafka. Позже можно перевести подписку на события из Kafka.

**Порт / адаптер**  
Ввести или расширить порт так, чтобы при `uow.HasTx(ctx)` вызывался cqrs EventBus с **WithPublisher(txPublisher)** (publisher от pgx.Tx из контекста); при необходимости после коммита — PublishInProcess для in-memory.

**Контекст и транзакция**  
Публикация **до** Commit использует тот же `ctx`, в котором лежит tx (и WithPublisher(txPublisher)). Публикация **после** Commit (например в InMemoryPublisher для Temporal) **не должна** получать этот же ctx: после Commit/Rollback транзакция невалидна, и если publisher или код по цепочке достанет tx из контекста и выполнит DB-операции — это опасно. После коммита использовать контекст без транзакции: `context.WithoutCancel(ctx)` или новый контекст без tx (например `context.Background()` или ctx с очищенным tx-маркером — в зависимости от реализации UoW). Паблишер после коммита не должен зависеть от DB-транзакции.

**Сериализация**  
Использовать marshaler/namer из cqrs; при необходимости DTO и соглашение по топикам (например `oms.order.created`, `oms.order.cancelled`) для консьюмеров Kafka.

## Последствия

- **Плюсы**
  - Доставка at-least-once: после успешного коммита события в БД и будут пересланы в Kafka (Forwarder с повторами).
  - Нет окна «коммит прошёл, публикация упала»: сохранение и запись в outbox в одной транзакции при использовании WithPublisher(txPublisher).
  - Совместимость с текущим использованием Watermill/Kafka в OMS; watermill-sql v4 поддерживает pgx.
- **Минусы**
  - Доп. зависимости: `watermill-sql/v4`, `watermill/components/forwarder`, пакет **go-sdk/cqrs**.
  - Нужен запущенный Forwarder и схема outbox (миграции, см. README cqrs — без автосоздания таблиц).
  - Изменение потока в хендлерах: публикация событий **до** Commit и при необходимости уведомление in-memory после коммита.

## Заметки по реализации

- **Хендлеры**: вместо «Commit → Publish» делать «Publish (в tx, с WithPublisher(txPublisher)) → Commit → при необходимости PublishInProcess(pubCtx, …)», где **pubCtx** — контекст без tx (например `context.WithoutCancel(ctx)` или `context.Background()`), чтобы код публикации после коммита не обращался к уже закрытой транзакции.
- **UoW**: без изменений; по-прежнему передавать `pgx.Tx` в контексте для создания tx-scoped publisher. После Commit в контексте для паблишера tx не использовать.
- **Топик Forwarder**: один внутренний топик outbox в Postgres (например `oms_domain_events_outbox`); Forwarder читает из него и публикует в топики Kafka по типу события или в один `oms.domain_events` с типом в метаданных.
- **go-sdk/cqrs**: использовать `EventBus` с `WithOutbox`; схему outbox создавать миграциями. Для транзакционной публикации вызывать **`Publish(ctx, evt, bus.WithPublisher(txPublisher))`**, где `txPublisher` — watermill-sql publisher от `pgx.Tx` из контекста (при необходимости обёрнут в forwarder.NewPublisher).

## Ссылки

- [go-sdk/cqrs](https://github.com/shortlink-org/go-sdk/tree/main/cqrs) — CQRS с опциональным Outbox Forwarder (`WithOutbox`, `RunForwarder`, `CloseForwarder`) и **PublishOption** (`WithPublisher`) для транзакционной публикации; [README](https://github.com/shortlink-org/go-sdk/blob/main/cqrs/README.md)
- [Watermill – Forwarder (outbox)](https://watermill.io/advanced/forwarder)
- [watermill-sql v4 (PostgreSQL, pgx)](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql) — `TxFromPgx`, `NewPublisher`
- [Watermill components/forwarder](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/forwarder)
- ADR-0004 (Hexagonal Architecture) — event publisher как порт, outbox как адаптер
