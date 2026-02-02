# Admin UI для управления Delivery

## Обзор

Современная админ-панель для управления сервисом доставки на базе **Next.js** и **Refine.dev**.
Заменяет Django-админку управления курьерами из сервиса `admin/`.

## Технологический стек

- **Фреймворк**: Next.js 14+ (App Router)
- **Admin-фреймворк**: [Refine.dev](https://refine.dev/) — React CRUD-фреймворк
- **UI библиотека**: Ant Design (встроенная интеграция с Refine)
- **Стили**: Tailwind CSS — для кастомных компонентов и утилит
- **State Management**: React Query (встроен в Refine)
- **API**: GraphQL через BFF (Cosmo Router) + новый `delivery-graphql` subgraph
- **Авторизация**: Ory Oathkeeper (JWT) — как в текущей админке
- **Карты**: Leaflet/MapLibre для визуализации локаций курьеров

## Архитектура

```
┌─────────────────┐         ┌─────────────────┐
│   admin-ui      │────────▶│      BFF        │
│   (Next.js +    │ GraphQL │  (Cosmo Router) │
│    Refine)      │         │   Federation    │
└─────────────────┘         └────────┬────────┘
         │                           │
         │                  ┌────────┴────────┐
         ▼                  ▼                 ▼
┌─────────────────┐  ┌─────────────┐  ┌─────────────────┐
│  Ory Oathkeeper │  │   admin     │  │ delivery-graphql│ ◀── НОВЫЙ
│  (Auth Proxy)   │  │  subgraph   │  │    subgraph     │
└─────────────────┘  └─────────────┘  └────────┬────────┘
                                               │ gRPC
                                      ┌────────▼────────┐
                                      │    Delivery     │
                                      │    Service      │
                                      │   (Rust/gRPC)   │
                                      └─────────────────┘
```

### Компоненты

- **admin-ui** — React приложение (этот проект)
- **BFF** — GraphQL Federation Router, объединяет все subgraphs
- **delivery-graphql** — новый subgraph, обёртка gRPC→GraphQL для Delivery service
- **Delivery Service** — существующий Rust сервис с gRPC API

---

## Фаза 1: Настройка проекта и инфраструктура

- [x] Инициализировать Next.js проект с TypeScript
- [x] Установить и настроить Refine.dev с Ant Design
- [x] Настроить Tailwind CSS
- [x] Настроить структуру проекта по конвенциям Refine
- [x] Настроить переменные окружения
- [x] Создать Docker и Helm charts (ops/)
- [x] Добавить в GitLab CI (matrix_build_docker.yml, matrix_build_helm.yml)
- [x] Настроить auth provider для интеграции с Ory

## Фаза 2: delivery-graphql Subgraph

Создание нового сервиса `delivery-graphql/` — GraphQL обёртка для Delivery gRPC API.

- [x] Инициализировать проект (Tailcall — декларативный gRPC→GraphQL)
- [x] Подключить gRPC клиент к Delivery service (через Tailcall конфиг)
- [x] Определить GraphQL схему для Courier, Delivery
- [x] Реализовать резолверы (Query, Mutation) — через Tailcall directives
- [x] Добавить в BFF federation (обновить `bff/graph.yaml`)
- [x] Настроить Docker и Helm charts

### GraphQL схема (draft)

```graphql
type Courier {
  id: ID!
  name: String!
  phone: String!
  email: String!
  transportType: TransportType!
  status: CourierStatus!
  rating: Float!
  workZone: String!
  currentLoad: Int!
  maxLoad: Int!
  currentLocation: Location
  workHours: WorkHours
  successfulDeliveries: Int!
  failedDeliveries: Int!
}

type Query {
  couriers(filter: CourierFilter, page: Int, pageSize: Int): CourierList!
  courier(id: ID!): Courier
}

type Mutation {
  registerCourier(input: RegisterCourierInput!): Courier!
  activateCourier(id: ID!): Courier!
  deactivateCourier(id: ID!, reason: String): Courier!
  archiveCourier(id: ID!, reason: String): Boolean!
  updateCourierContact(id: ID!, input: UpdateContactInput!): Courier!
  updateCourierSchedule(id: ID!, input: UpdateScheduleInput!): Courier!
  changeCourierTransport(id: ID!, transportType: TransportType!): Courier!
}
```

## Фаза 3: API Layer (admin-ui)

- [x] Настроить GraphQL клиент (Apollo Client) — настроен, но пока mock
- [x] Создать Refine data provider — mock provider готов
- [ ] Подключить реальный GraphQL через BFF
- [ ] Сгенерировать TypeScript типы из GraphQL схемы (codegen)
- [ ] Реализовать auth provider (работа с JWT)
- [x] Добавить обработку ошибок и уведомления

## Фаза 4: Управление курьерами (MVP)

### 4.1 Страница списка курьеров
- [x] Таблица с колонками: Имя, Статус, Транспорт, Зона, Рейтинг, Загрузка
- [x] Фильтры: Статус, Тип транспорта (UI готов, данные mock)
- [ ] Пагинация (серверная) — UI готов, нужно подключить к GraphQL
- [ ] Сортировка по колонкам
- [x] Быстрые действия: Активировать/Деактивировать

### 4.2 Страница деталей курьера
- [x] Карточка курьера (имя, контакты, бейдж статуса)
- [x] Секция рабочего расписания (часы, дни, зона)
- [x] Информация о транспорте (тип, макс. дистанция, вместимость)
- [x] Статистика (количество доставок, рейтинг)
- [ ] Список недавних доставок — placeholder готов
- [x] Кнопки действий: Активировать, Деактивировать, Архивировать

### 4.3 Создание/редактирование курьера
- [x] Форма регистрации с валидацией
- [x] Форма контактной информации (телефон, email)
- [x] Форма рабочего расписания (выбор часов, дней)
- [x] Выбор типа транспорта
- [ ] Превью вместимости при выборе транспорта

### 4.4 Карта курьеров
- [ ] Компонент карты с маркерами курьеров
- [ ] Кластеризация маркеров для производительности
- [ ] Обновление локаций в реальном времени (WebSocket/polling)
- [ ] Фильтрация курьеров на карте по статусу
- [ ] Popup с краткой информацией о курьере

## Фаза 5: Управление доставками

- [ ] Список доставок с фильтрацией
- [ ] Страница деталей доставки
- [ ] Назначение/переназначение курьера на доставку
- [ ] Отслеживание доставки

## Фаза 6: Дашборд и аналитика

- [x] Обзорный дашборд с KPI (базовый, mock данные)
- [ ] График доступности курьеров
- [ ] Статистика доставок
- [ ] Тепловая карта по зонам

## Фаза 7: Доработка и продакшен

- [ ] Адаптивный дизайн (поддержка планшетов)
- [ ] Поддержка темной темы
- [ ] Интернационализация (i18n)
- [ ] E2E тесты на Playwright
- [ ] Оптимизация производительности
- [ ] Документация

## Фаза 8: Миграция и cleanup

После полного переноса функционала в admin-ui — удаление дублирующегося кода из Django админки.

### Удалить из `admin/`

- [x] `src/domain/couriers/` — views, forms, templates, urls
- [x] `src/infrastructure/grpc/delivery_client.py` — gRPC клиент
- [x] Убрать couriers из `admin/src/admin/urls.py`
- [x] Удалить тесты `tests/domain/couriers/`
- [x] Обновить `dashboard.py` — убрать компоненты курьеров
- [ ] Обновить документацию

### Проверить

- [ ] Убедиться что все функции курьеров работают в admin-ui
- [ ] Проверить что Django админка работает без couriers модуля
- [ ] Обновить CI/CD pipelines если нужно

---

## Модели данных (из Delivery Service)

### Курьер
```typescript
interface Courier {
  courier_id: string;
  name: string;
  phone: string;
  email: string;
  transport_type: 'WALKING' | 'BICYCLE' | 'MOTORCYCLE' | 'CAR';
  max_distance_km: number;
  status: 'UNAVAILABLE' | 'FREE' | 'BUSY' | 'ARCHIVED';
  current_load: number;
  max_load: number;
  rating: number;
  work_hours?: WorkHours;
  work_zone: string;
  current_location?: Location;
  successful_deliveries: number;
  failed_deliveries: number;
  created_at?: string;
  last_active_at?: string;
}

interface WorkHours {
  start_time: string; // "HH:MM"
  end_time: string;   // "HH:MM"
  work_days: number[]; // 0-6 (Воскресенье-Суббота)
}

interface Location {
  latitude: number;
  longitude: number;
}
```

### Запись о доставке
```typescript
interface DeliveryRecord {
  package_id: string;
  order_id: string;
  status: 'ACCEPTED' | 'IN_POOL' | 'ASSIGNED' | 'IN_TRANSIT' | 'DELIVERED' | 'NOT_DELIVERED' | 'REQUIRES_HANDLING';
  pickup_address?: Address;
  delivery_address?: Address;
  assigned_at?: string;
  delivered_at?: string;
  priority: 'NORMAL' | 'URGENT';
}

interface Address {
  street: string;
  city: string;
  postal_code: string;
  country: string;
  latitude: number;
  longitude: number;
}
```

---

## Маппинг API

Соответствие gRPC методов Delivery service → GraphQL → Refine операций:

| gRPC метод | GraphQL | Refine операция |
|------------|---------|-----------------|
| GetCourierPool | `couriers` query | getList |
| GetCourier | `courier` query | getOne |
| RegisterCourier | `registerCourier` mutation | create |
| UpdateContactInfo | `updateCourierContact` mutation | update (custom) |
| UpdateWorkSchedule | `updateCourierSchedule` mutation | update (custom) |
| ChangeTransportType | `changeCourierTransport` mutation | update (custom) |
| ActivateCourier | `activateCourier` mutation | custom |
| DeactivateCourier | `deactivateCourier` mutation | custom |
| ArchiveCourier | `archiveCourier` mutation | delete (soft) |
| GetCourierDeliveries | `courierDeliveries` query | custom |

---

## Структура директорий

### admin-ui (этот проект)

```
admin-ui/
├── app/                      # Next.js App Router
│   ├── layout.tsx
│   ├── page.tsx              # Дашборд
│   └── couriers/
│       ├── page.tsx          # Список
│       ├── create/page.tsx   # Форма создания
│       ├── [id]/
│       │   ├── page.tsx      # Просмотр/Детали
│       │   └── edit/page.tsx # Форма редактирования
│       └── map/page.tsx      # Карта
├── components/
│   ├── couriers/
│   │   ├── CourierStatusBadge.tsx
│   │   ├── CourierCard.tsx
│   │   └── CourierMap.tsx
│   └── common/
├── providers/
│   ├── data-provider.ts      # Refine GraphQL data provider
│   └── auth-provider.ts      # Интеграция с Ory
├── graphql/
│   ├── queries/              # GraphQL queries
│   ├── mutations/            # GraphQL mutations
│   └── generated/            # Сгенерированные типы (codegen)
├── lib/
│   └── apollo-client.ts      # Apollo Client настройка
├── ops/
│   ├── dockerfile/
│   └── Helm/
├── public/
├── package.json
├── codegen.ts                # GraphQL codegen конфиг
├── next.config.js
├── tailwind.config.js
└── tsconfig.json
```

### delivery-graphql (новый сервис)

```
delivery-graphql/
├── src/
│   ├── schema/               # GraphQL схема
│   │   └── courier.graphql
│   ├── resolvers/            # Резолверы
│   │   └── courier.ts
│   └── grpc/                 # gRPC клиент к Delivery
│       └── client.ts
├── ops/
│   ├── dockerfile/
│   └── Helm/
├── package.json
└── tsconfig.json
```

---

## Начало работы (следующие шаги)

### Шаг 1: Создать delivery-graphql subgraph

```bash
# В корне shop/
mkdir delivery-graphql
cd delivery-graphql
# Инициализировать проект (Node.js + Apollo или Go + gqlgen)
```

### Шаг 2: Инициализировать admin-ui

```bash
cd admin-ui
npx create-refine-app@latest . -- --preset refine-nextjs
# Выбрать: Ant Design, Next.js App Router, GraphQL
```

### Шаг 3: Добавить Tailwind CSS

```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### Шаг 4: Настроить GraphQL

- Настроить Apollo Client / urql для работы с BFF
- Настроить GraphQL codegen для генерации типов

### Шаг 5: Создать первый ресурс

- Создать страницу списка курьеров
- Подключить к GraphQL через Refine data provider

### Шаг 6: Добавить auth provider

- Интеграция с Ory Oathkeeper JWT

---

## Ссылки

- [Документация Refine.dev](https://refine.dev/docs/)
- [Refine GraphQL Data Provider](https://refine.dev/docs/data/packages/graphql/)
- [Гайд Refine + Next.js](https://refine.dev/docs/guides-concepts/routing/integrations/next-js/)
- [Компоненты Ant Design](https://ant.design/components/overview/)
- [Apollo Client](https://www.apollographql.com/docs/react/)
- [GraphQL Code Generator](https://the-guild.dev/graphql/codegen)
- [Cosmo Router (BFF)](https://cosmo-docs.wundergraph.com/)
- Текущая Django админка: `admin/src/domain/couriers/`
- Delivery Service proto: `delivery/proto/`
- BFF конфигурация: `bff/`
