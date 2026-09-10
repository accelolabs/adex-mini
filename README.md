# mini-AdEx

Небольшой HTTP-сервис, который фильтрует DSP-партнёров по параметрам аукциона и параллельно отправляет им запросы через mock-клиент.

## Запуск

Требуются Docker и Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Compose поднимет PostgreSQL, применит goose-миграцию и запустит API на `http://localhost:8080`.

Проверка состояния:

```bash
curl http://localhost:8080/healthz
```

Остановка:

```bash
docker compose down
```

Чтобы удалить также локальные данные PostgreSQL:

```bash
docker compose down --volumes
```

## Пример аукциона

```bash
curl -X POST http://localhost:8080/auction \
  -H 'Content-Type: application/json' \
  -d '{
    "request_id": "3f0a1c9e-2b1d-4a8f-9c11-7e6b2d0a55f1",
    "country": "RU",
    "device_type": "mobile",
    "bid_floor": 1.5,
    "categories": ["news", "sport"]
  }'
```

Пример ответа:

```json
{
  "request_id": "3f0a1c9e-2b1d-4a8f-9c11-7e6b2d0a55f1",
  "matched_dsps": ["dsp-alpha", "dsp-gamma"],
  "sent": 2,
  "succeeded": 1,
  "duration_ms": 200
}
```

`dsp-alpha` успешно отвечает, а `dsp-gamma` имитирует таймаут. `dsp-beta` не подходит по стране и устройству, `dsp-disabled` выключен. Таймаут одного DSP не прерывает весь аукцион.

## Партнёры и mock DSP

Начальные партнёры находятся в [`migrations/00001_init.sql`](migrations/00001_init.sql). Чтобы добавить партнёра, создайте следующую goose-миграцию с `INSERT INTO partners`.

Поле `endpoint` определяет поведение встроенного mock-клиента:

- `mock://success` - успешный ответ;
- `mock://error` - ошибка DSP;
- `mock://timeout` - ожидание до общего таймаута аукциона.

Mock работает внутри процесса приложения, поэтому отдельный DSP-сервер запускать не требуется.

## Конфигурация

Приложение читает настройки из ENV через `cleanenv`:

| Переменная | Назначение | Значение по умолчанию |
|---|---|---|
| `HTTP_ADDR` | Адрес HTTP-сервера | `:8080` |
| `DATABASE_URL` | PostgreSQL DSN | обязательное |
| `AUCTION_TIMEOUT` | Общий таймаут DSP | `200ms` |
| `SHUTDOWN_TIMEOUT` | Таймаут завершения | `5s` |
| `LOG_LEVEL` | Уровень JSON-логов `slog` | `info` |

Файл `.env` предназначен только для локального запуска и исключён из Git. Исходный шаблон хранится в `.env.example`.

## Структура

```text
cmd/adex/main.go              сборка зависимостей и запуск
internal/config/config.go     ENV-конфигурация
internal/auction/model.go     чистые сущности
internal/auction/postgres.go  PostgreSQL-хранилище
internal/auction/service.go   оркестрация аукциона
internal/auction/pretargeting.go правила фильтрации
internal/auction/dto.go       HTTP DTO и валидация
internal/auction/handlers.go  HTTP handlers
```

HTTP-слой зависит от интерфейса сервиса, а сервис - от интерфейсов репозитория и DSP-клиента.

## Тесты и миграции

```bash
go test ./...
make migrate-up
make migrate-down
```

Unit-тесты покрывают правила претаргетинга, частичный успех партнёров и общий таймаут.

## Что можно улучшить

- заменить mock-клиент реальным HTTP-клиентом DSP;
- добавить метрики и трассировку;
- добавить интеграционные тесты PostgreSQL;
- логировать детальные причины фильтрации каждого партнёра.
