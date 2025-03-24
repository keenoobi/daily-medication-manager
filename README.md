# Сервис для управления приемом лекарств

## Общее описание
RESTful API для управления расписанием приёма лекарств с автоматическим округлением времени и поддержкой Docker-развёртывания.

---

## Быстрый старт с Docker

### Требования
1. Docker Engine 20.10+
2. Docker Compose 2.20+

### Запуск системы
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/yourusername/medication-scheduler.git
   cd medication-scheduler
   ```

2. Запустите сервисы:
   ```bash
   docker compose up --build
   ```

### Структура контейнеров
1. **postgres**: Контейнер с PostgreSQL 14
   - Порт: 5432
   - Volume: `postgres_data` для сохранения данных
   - Автоматическое создание БД при первом запуске

2. **app**: Основное приложение
   - Порт: 8080
   - Автоматически применяет миграции при запуске
   - Переменные окружения настраиваются через `.env`

---

## Конфигурация

### Файл .env
Создайте файл `.env` в корне проекта (пример):
```ini
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=scheduler

# Application
SERVER_PORT=8080
NEXT_TAKINGS_PERIOD=1h
LOG_LEVEL=info
```

### Переменные окружения
| Переменная               | По умолчанию     | Описание                          |
|--------------------------|------------------|-----------------------------------|
| POSTGRES_USER            | postgres         | Пользователь PostgreSQL           |
| POSTGRES_PASSWORD        | password         | Пароль PostgreSQL                 |
| POSTGRES_DB              | scheduler        | Название базы данных              |
| SERVER_PORT              | 8080             | Порт для HTTP-сервера             |
| NEXT_TAKINGS_PERIOD      | 1h               | Период для поиска ближайших приёмов |
| LOG_LEVEL                | info             | Уровень логирования (debug/info/warn/error) |

---

## API Endpoints

### 1. Создание расписания
`POST /schedule`
```bash
curl -X POST http://localhost:8080/schedule \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "medication": "Аспирин", "frequency": "1h", "duration": "24h"}'
```

### 2. Получение списка расписаний
`GET /schedules?user_id=123`
```bash
curl "http://localhost:8080/schedules?user_id=123"
```

### 3. Получение деталей расписания
`GET /schedule?user_id=123&schedule_id=1`
```bash
curl "http://localhost:8080/schedule?user_id=123&schedule_id=1"
```

### 4. Ближайшие приёмы лекарств
`GET /next_takings?user_id=123`
```bash
curl "http://localhost:8080/next_takings?user_id=123"
```

---

## Управление системой

### Команды Docker Compose
| Команда                          | Описание                              |
|----------------------------------|---------------------------------------|
| `docker-compose up --build`      | Сборка и запуск в foreground         |
| `docker-compose up -d`           | Запуск в фоновом режиме              |
| `docker-compose down`            | Остановка и удаление контейнеров     |
| `docker-compose logs -f app`     | Просмотр логов приложения            |
| `docker-compose exec postgres psql -U user db` | Доступ к PostgreSQL |

---

## Особенности реализации

### Автоматические миграции
- При запуске приложение автоматически применяет миграции из папки `migrations`
- Миграции выполняются при каждом старте контейнера

### Персистентность данных
- Данные PostgreSQL сохраняются в Docker volume `postgres_data`
- Для сброса данных:
  ```bash
  docker-compose down -v
  ```

## Технологический стек
- Go 1.20+
- Gin
- PostgreSQL 14+
- Docker Compose
