# Pricer Service — Roadmap

## Обзор

Новый сервис **Pricer** с нуля: расчёт налогов и скидок для корзины/заказа. Заменяет `pricer_old` (Go). Ключевые решения:

- **Rust** — язык реализации
- **OPA (Open Policy Agent)** — движок политик (Rego)
- **BDD** — тесты в формате Gherkin (cucumber-rs)
- **DDD** — по аналогии с `delivery/` и `oms/`
- **gRPC + Buf** — API и регистрируемые модули в buf registry
- **Интеграция с OMS** — вызовы из OMS при оформлении корзины/заказа

## Контекст в экосистеме

```
┌─────────────────┐         ┌─────────────────┐
│   BFF / UI      │────────▶│      OMS        │
│   (Cart/Order)  │  gRPC   │  Cart / Order   │
└─────────────────┘         └────────┬────────┘
                                     │ gRPC: GetCartTotal, GetOrderTotal
                                     ▼
                            ┌─────────────────┐
                            │     Pricer       │
                            │  (Rust + OPA)    │
                            └─────────────────┘
```

OMS при отображении корзины или создании заказа вызывает Pricer для получения итогов (скидки, налоги, финальная сумма).

---

## Технологический стек

| Компонент | Выбор | Примечание |
|-----------|--------|------------|
| Язык | **Rust** | Консистентность с delivery, производительность, типобезопасность |
| Политики | **OPA / Rego** | Та же модель, что в pricer_old; политики можно переиспользовать/перенести |
| Вызов OPA | **opa-client** или **rego в process** | Либо sidecar/bundled OPA, либо embedded Rego (Rust SDK при наличии) |
| API | **gRPC** | Консистентность с OMS, delivery; Buf для схем и registry |
| Тесты | **BDD (Gherkin)** | cucumber-rs для сценариев; unit — стандартный `#[test]` |
| Структура | **DDD** | domain / usecases / infrastructure по аналогии с delivery и OMS |

---

## Фаза 1: Каркас проекта и DDD-структура

- [ ] Инициализировать Cargo workspace (Rust)
- [ ] Настроить структуру каталогов в стиле DDD (по аналогии с `delivery/`):

```
pricer/
├── src/
│   ├── main.rs
│   ├── lib.rs
│   ├── config.rs
│   ├── domain/                    # DDD Domain
│   │   ├── mod.rs
│   │   ├── model/                 # Агрегаты, value objects
│   │   │   ├── cart/              # Cart, CartItem (или общий pricing context)
│   │   │   └── vo/                # Money, Quantity при необходимости
│   │   ├── services/              # Domain services (если нужна чистая логика поверх OPA)
│   │   └── ports/                 # Traits: PolicyEvaluator, etc.
│   ├── usecases/                  # Application use cases
│   │   ├── mod.rs
│   │   ├── README.md              # Обзор юзкейсов
│   │   ├── calculate_cart_discount/   # Список товаров корзины → размер скидки
│   │   │   └── README.md
│   │   └── apply_promo_code/          # Применение промокода
│   │       └── README.md
│   └── infrastructure/           # Адаптеры
│       ├── mod.rs
│       ├── policy/                # OPA evaluator (Rego)
│       └── rpc/                    # gRPC server + handlers
├── policies/                      # Rego (discounts, taxes) — можно перенести из pricer_old
│   ├── discounts/
│   └── taxes/
├── tests/                         # Интеграционные / BDD
│   └── bdd/
│       └── features/
├── buf.yaml                       # Buf modules (см. фазу 4)
├── Cargo.toml
├── Makefile
└── ops/
    ├── dockerfile/
    └── Makefile/
```

- [ ] Добавить README с назначением сервиса и ссылкой на ROADMAP
- [ ] Опционально: ADR 0001-init (Rust, DDD, OPA)

---

## Фаза 2: OPA как основа политик

- [ ] Определить контракт входа в OPA: структура `input` (items, customerId, params для скидок/налогов)
- [ ] Перенести или переписать Rego-политики из `pricer_old/policies/`:
  - `policies/discounts/` — например total_brand_discount, general_discount
  - `policies/taxes/` — total_markup, vat
- [ ] Выбрать способ вызова OPA из Rust:
  - **Вариант A:** OPA как отдельный процесс / sidecar — HTTP или Go SDK не подходит; возможен subprocess или REST API OPA
  - **Вариант B:** Embedded Rego — использовать crate для выполнения Rego (например, если есть Rust-обёртка над OPA lib)
  - **Вариант C:** Вызов OPA REST API (bundled или внешний) — простой вариант для старта
- [ ] Реализовать порт `PolicyEvaluator` (discount, tax) и адаптер в `infrastructure/policy/`
- [ ] Кэширование результатов оценки (по аналогии с pricer_old — по входным данным) — опционально в первой итерации
- [ ] Unit-тесты для маппинга domain → OPA input и разбора результата

---

## Фаза 3: BDD для тестов

- [ ] Добавить в проект **cucumber-rs** (зависимость и настройка)
- [ ] Завести каталог сценариев: `tests/bdd/features/` (или `tests/features/`)
- [ ] Описать сценарии в Gherkin, например:
  - `calculate_cart_discount.feature` — подаём список товаров корзины, ожидаем размер скидки
  - `apply_promo_code.feature` — валидный/невалидный/истёкший промокод, min order, stacking
  - `discount_policy.feature` — скидки по бренду (Apple/Samsung), общая скидка
  - при необходимости: `tax_policy.feature` — налог/markup по правилам
- [ ] Реализовать step definitions для юзкейсов `CalculateCartDiscount` и `ApplyPromoCode` (in-memory или тестовый gRPC client)
- [ ] Интегрировать запуск BDD в `Makefile` / CI (`cargo test` или отдельная команда для cucumber)

---

## Фаза 4: gRPC API и Buf Registry

- [ ] Спроектировать gRPC API (минимум):
  - `CalculateCartDiscount(Cart + pricing context) -> DiscountAmount` — список товаров корзины (buf: shop-oms) → размер скидки
  - `ApplyPromoCode(Cart + promo_code) -> ApplyPromoResult` — применение промокода → новая скидка или ошибка валидации
  - при необходимости: `CalculateCartTotal` / `CalculateOrderTotal` для полного итога (налог + скидка)
- [ ] Завести proto в структуре, совместимой с Buf:
  - доменные сообщения (Cart, CartItem, CartTotal) — отдельный модуль или общий с OMS по соглашению
  - RPC-сервис (например `pricer.v1.PricerService`)
- [ ] Настроить `buf.yaml`: модули, линт, breaking, зависимости (googleapis; при необходимости `buf.build/shortlink-org/shop-oms` для общих типов)
- [ ] Генерация кода: `buf generate` → Rust (tonic, prost)
- [ ] Реализовать gRPC handlers в `infrastructure/rpc/` с вызовом use case
- [ ] Зарегистрировать модули в buf registry (buf.build/shortlink-org/...) по процессу организации
- [ ] Добавить Dockerfile и базовый Helm chart в `ops/`

---

## Фаза 5: Интеграция с OMS

- [ ] Уточнить сценарии вызова Pricer из OMS:
  - при запросе корзины (GetCart) — опционально обогащать итогами от Pricer;
  - при создании заказа (CreateOrderFromCart) — передать в заказ итоги (totalTax, totalDiscount, finalPrice) для консистентности и отображения.
- [ ] Определить контракт: либо OMS вызывает Pricer по gRPC, либо BFF вызывает оба и склеивает данные.
- [ ] Если вызов из OMS:
  - добавить в OMS gRPC-клиент к Pricer (или общий HTTP/gRPC client);
  - в workflow/use case корзины/заказа — вызов Pricer и подстановка итогов в ответ/событие.
- [ ] Документировать контракт (proto, примеры вызовов) и обновить C4/диаграммы (docs/ADR при необходимости).
- [ ] E2E/интеграционные тесты: OMS → Pricer (или BFF → OMS → Pricer) для ключевого сценария.

---

## Фаза 6: Операционная готовность и доработки

- [ ] Конфигурация: пути к политикам, параметры скидок/налогов, адрес OPA (если внешний)
- [ ] Логирование и метрики (tracing, Prometheus)
- [ ] Graceful shutdown
- [ ] CI: сборка, линт, тесты (unit + BDD), сбор образа
- [ ] Документация: README, ADR, описание API (Postman/аналог при необходимости)

---

## Зависимости между фазами

```
Фаза 1 (каркас, DDD) ─────────────────────────────────────────────────────┐
       │                                                                   │
       ▼                                                                   │
Фаза 2 (OPA) ────────────────┐                                            │
       │                      │                                            │
       ▼                      ▼                                            ▼
Фаза 3 (BDD) ◀───────────────┴─── использует domain + usecases + OPA
       │
       ▼
Фаза 4 (gRPC + Buf) ────────────── использует usecases
       │
       ▼
Фаза 5 (OMS) ───────────────────── использует gRPC API
       │
       ▼
Фаза 6 (ops, observability)
```

---

## Открытые решения

1. **OPA из Rust:** embedded vs отдельный процесс vs REST — зависит от доступных Rust-библиотек и требований к латентности.
2. **Общие proto с OMS:** Cart/CartItem могут быть в OMS; Pricer тогда зависит от `buf.build/shortlink-org/shop-oms` и использует те же типы для запроса.
3. **Кэш:** нужен ли кэш результатов OPA в первой версии или достаточно «каждый запрос — один вызов OPA».

После утверждения роадмапа можно детализировать каждую фазу в отдельных ADR или подзадачах.
