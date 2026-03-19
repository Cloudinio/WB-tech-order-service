# WB Tech Order Service

Сервис обработки заказов на Go.

Приложение получает заказы из Kafka, сохраняет их в PostgreSQL, кэширует в памяти для быстрого доступа и отдает через HTTP API.

## Quick start

```bash
make up
~/go/bin/goose -dir ./migrations postgres "postgres://orders_user:orders_pass@localhost:5432/orders_db?sslmode=disable" up
make run
```

## Что умеет сервис

- читает сообщения с заказами из Kafka
- валидирует входные данные
- сохраняет заказ в PostgreSQL
- кэширует заказы в памяти
- отдает заказ по `order_uid` через HTTP API
- прогревает кэш при старте батчами из БД
- обрабатывает дубликаты сообщений по `order_uid`
- отдает Prometheus-метрики
- корректно завершается через graceful shutdown

## Стек

- Go
- PostgreSQL
- Kafka
- Docker Compose
- pgx
- chi
- Prometheus
- goose
- DBeaver
- unit и integration tests

## Архитектура

Проект разбит на несколько слоев:

- `domain` — бизнес-сущности (`Order`, `Delivery`, `Payment`, `Item`)
- `repository/postgres` — работа с PostgreSQL
- `cache/memory` — in-memory cache
- `usecase` — бизнес-логика получения заказа через cache + repo
- `broker/kafka` — Kafka consumer и DTO сообщений
- `transport/http` — HTTP handlers и роутинг
- `metrics` — Prometheus-метрики

## Поток данных

### Запись заказа

```text
Kafka message
-> Consumer
-> JSON DTO
-> domain.Order
-> Validate()
-> PostgreSQL Save()
-> Cache Set()
```

### Чтение заказа

```text
HTTP GET /api/v1/orders/{order_uid}
-> Handler
-> Usecase
-> Cache Get()
   -> hit: вернуть из cache
   -> miss: взять из PostgreSQL, положить в cache, вернуть
```

### Прогрев кэша

```text
Application start
-> Repository.ListRecent()
-> batched loading from PostgreSQL
-> Cache.Warmup()
-> HTTP server start
```

## Структура проекта

```text
wb-tech-order-service/
├── api/
│   └── openapi.yaml
├── cmd/
│   └── app/
│       └── main.go
├── deployments/
│   └── docker-compose.yml
├── internal/
│   ├── broker/
│   │   └── kafka/
│   ├── cache/
│   │   └── memory/
│   ├── config/
│   ├── domain/
│   ├── metrics/
│   ├── repository/
│   │   └── postgres/
│   ├── transport/
│   │   └── http/
│   └── usecase/
├── migrations/
├── test/
│   └── integration/
├── Makefile
└── README.md
```

## Запуск проекта

### 1. Поднять инфраструктуру

```bash
make up
```

Или:

```bash
docker compose -f deployments/docker-compose.yml up -d
```

### 2. Применить миграции

```bash
~/go/bin/goose -dir ./migrations postgres "postgres://orders_user:orders_pass@localhost:5432/orders_db?sslmode=disable" up
```

### 3. Запустить приложение

```bash
make run
```

## Конфигурация

Пример `.env`:

```env
APP_PORT=8081

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=orders_db
POSTGRES_USER=orders_user
POSTGRES_PASSWORD=orders_pass

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=order-service
```

## HTTP API

### Health check

```http
GET /healthz
```

Ответ:

```text
ok
```

### Metrics

```http
GET /metrics
```

Возвращает Prometheus-метрики.

### Получить заказ по UID

```http
GET /api/v1/orders/{order_uid}
```

Пример:

```bash
curl http://localhost:8081/api/v1/orders/b563feb7b2b84b6test
```

Пример ответа:

```json
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```

## OpenAPI

Контракт API описан в файле:

```text
api/openapi.yaml
```

Его можно открыть через любой OpenAPI viewer, например Swagger Editor.

## Kafka

Сервис читает сообщения из Kafka topic:

```text
orders
```

Пример сообщения:

```json
{
  "order_uid": "kafka-order-001",
  "track_number": "TRACK-KAFKA-001",
  "entry": "WBIL",
  "delivery": {
    "name": "Kafka User",
    "phone": "+79990000000",
    "zip": "123456",
    "city": "Moscow",
    "address": "Pushkina 10",
    "region": "Moscow",
    "email": "kafka@test.com"
  },
  "payment": {
    "transaction": "tx-kafka-001",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "TRACK-KAFKA-001",
      "price": 453,
      "rid": "rid-kafka-001",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```

## Идемпотентность

Повторная доставка одного и того же заказа не ломает сервис.

В качестве естественного ключа используется:

```text
order_uid
```

Если сообщение с таким `order_uid` уже было обработано, сервис распознает его как дубликат и игнорирует повторную запись.

## Кэш

Используется in-memory cache на `map[string]domain.Order` + `sync.RWMutex`.

### Как это работает

- при чтении заказ сначала ищется в cache
- если найден — возвращается сразу
- если не найден — читается из PostgreSQL и затем кладется в cache

### Прогрев кэша

При старте сервиса кэш заполняется батчами из БД по `date_created DESC`.

## Метрики

Сервис отдает Prometheus-метрики через:

```http
GET /metrics
```

Примеры метрик:

- `kafka_messages_total`
- `kafka_invalid_messages_total`
- `kafka_duplicate_messages_total`
- `kafka_save_errors_total`
- `cache_hits_total`
- `cache_misses_total`
- `http_requests_total`

## Graceful shutdown

Приложение корректно обрабатывает `SIGINT` / `SIGTERM`:

- перестает принимать новые HTTP-запросы
- останавливает Kafka consumer
- отменяет общий context
- корректно завершает работу

## Тесты

### Unit tests

Покрыты:
- `domain.Order.Validate()`
- `kafka.OrderMessage.ToDomain()`
- `memory.OrderCache`
- `usecase.OrderService`

Запуск:

```bash
go test ./...
```

### Integration tests

Покрыты:
- `OrderRepository.Save()`
- `OrderRepository.GetByUID()`
- `OrderRepository.ListRecent()`
- HTTP endpoint
- Kafka consumer flow

Запуск:

```bash
go test ./test/integration -v
```
