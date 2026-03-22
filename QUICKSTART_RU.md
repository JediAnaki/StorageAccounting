# Быстрый старт - Система складского учета

**Проект**: Warehouse Inventory Management via Telegram Mini App
**Язык**: Go 1.25 + PostgreSQL + Vanilla JS

---

## 🚀 Запуск за 5 минут

### Способ 1: Docker Compose (рекомендуется)

```bash
# 1. Перейдите в директорию deployments
cd deployments

# 2. Создайте файл .env с настройками
cat > .env << 'EOF'
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/inventory_db?sslmode=disable
TELEGRAM_BOT_TOKEN=your_bot_token_here
S3_ENDPOINT=your_s3_endpoint
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=inventory-photos
SERVER_PORT=8080
EOF

# 3. Запустите контейнеры
docker-compose up -d

# 4. Проверьте статус
docker-compose ps

# 5. Примените миграции
docker-compose exec backend sh -c "psql \$DATABASE_URL -f /app/migrations/001_create_items.sql"
docker-compose exec backend sh -c "psql \$DATABASE_URL -f /app/migrations/002_create_transactions.sql"

# 6. Проверьте health check
curl http://localhost:8080/health
```

**Ожидаемый ответ**:
```json
{"status":"healthy"}
```

---

### Способ 2: Локальная разработка

#### Предварительные требования
- Go 1.25+
- PostgreSQL 14+
- Node.js (для фронтенда, опционально)

#### Шаги

**1. Настройка базы данных**

```bash
# Создайте базу данных
createdb inventory_db

# Примените миграции
cd backend
psql -d inventory_db -f migrations/001_create_items.sql
psql -d inventory_db -f migrations/002_create_transactions.sql
```

**2. Настройка переменных окружения**

```bash
# Создайте файл .env в директории backend/
cat > backend/.env << 'EOF'
DATABASE_URL=postgresql://localhost:5432/inventory_db?sslmode=disable
TELEGRAM_BOT_TOKEN=your_bot_token_here
S3_ENDPOINT=your_s3_endpoint
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=inventory-photos
SERVER_PORT=8080
EOF

# Экспортируйте переменные
export $(cat backend/.env | xargs)
```

**3. Запуск backend**

```bash
cd backend

# Установите зависимости
go mod download

# Запустите сервер
go run cmd/server/main.go
```

**4. Запуск frontend (для локального тестирования)**

```bash
cd frontend

# Простой HTTP сервер
python3 -m http.server 8081

# Или используйте любой другой статический сервер
# npm install -g http-server
# http-server -p 8081
```

**5. Откройте в браузере**

```
http://localhost:8081
```

---

## 📊 Тестовые данные

Для быстрого тестирования создайте тестовые данные:

```bash
# Создайте файл test_data.sql
cat > test_data.sql << 'EOF'
-- Тестовые товары
INSERT INTO items (name, current_quantity, created_by) VALUES
('Ноутбук Dell XPS 15', 10, 12345),
('Монитор Samsung 27"', 25, 12345),
('Клавиатура Logitech', 50, 12345),
('Мышь беспроводная', 100, 12345),
('Кабель HDMI 2м', 0, 12345),
('MacBook Pro 16"', 5, 12345),
('iPhone 14 Pro', 15, 12345),
('iPad Air', 8, 12345);

-- Тестовые транзакции
INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
SELECT id, 12345, 0, current_quantity, current_quantity, 'Начальный остаток'
FROM items;

-- Дополнительные транзакции для истории
DO $$
DECLARE
    item_record RECORD;
BEGIN
    FOR item_record IN SELECT id, current_quantity FROM items WHERE current_quantity > 0 LOOP
        -- Добавление
        INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
        VALUES (item_record.id, 12345, item_record.current_quantity, item_record.current_quantity + 10, 10, 'Поступление от поставщика');

        -- Списание
        INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
        VALUES (item_record.id, 12345, item_record.current_quantity + 10, item_record.current_quantity + 5, -5, 'Выдано на офис');
    END LOOP;
END $$;
EOF

# Примените тестовые данные
psql -d inventory_db -f test_data.sql
```

---

## 🧪 Быстрые тесты

### Проверка API

```bash
# 1. Health check
curl http://localhost:8080/health

# 2. Список товаров (требуется авторизация)
curl -H "X-Telegram-Init-Data: test_data" \
  http://localhost:8080/api/items?limit=10

# 3. Детали товара
curl -H "X-Telegram-Init-Data: test_data" \
  http://localhost:8080/api/items/{item_id}

# 4. История транзакций
curl -H "X-Telegram-Init-Data: test_data" \
  http://localhost:8080/api/items/{item_id}/transactions

# 5. Добавление товара на склад
curl -X PUT \
  -H "X-Telegram-Init-Data: test_data" \
  -H "Content-Type: application/json" \
  -d '{"quantity":5,"note":"Тестовое добавление"}' \
  http://localhost:8080/api/items/{item_id}/add

# 6. Списание товара
curl -X PUT \
  -H "X-Telegram-Init-Data: test_data" \
  -H "Content-Type: application/json" \
  -d '{"quantity":3,"note":"Тестовое списание"}' \
  http://localhost:8080/api/items/{item_id}/remove
```

### Проверка базы данных

```bash
# Количество товаров
psql -d inventory_db -c "SELECT COUNT(*) FROM items WHERE deleted_at IS NULL;"

# Последние 5 товаров
psql -d inventory_db -c "SELECT name, current_quantity FROM items ORDER BY created_at DESC LIMIT 5;"

# Количество транзакций
psql -d inventory_db -c "SELECT COUNT(*) FROM transactions;"

# История изменений товара
psql -d inventory_db -c "SELECT delta, note, timestamp FROM transactions WHERE item_id = (SELECT id FROM items LIMIT 1) ORDER BY timestamp DESC LIMIT 10;"
```

---

## 📱 Интеграция с Telegram

### 1. Создание бота

```bash
# 1. Откройте @BotFather в Telegram
# 2. Создайте нового бота: /newbot
# 3. Получите токен бота
# 4. Настройте Mini App: /newapp
# 5. Укажите URL вашего frontend (HTTPS обязателен!)
```

### 2. Настройка webhook (опционально)

```bash
# Установите webhook для бота
curl -X POST "https://api.telegram.org/bot{YOUR_BOT_TOKEN}/setWebhook" \
  -d "url=https://your-domain.com/webhook"
```

### 3. Тестирование в Telegram

1. Откройте бота в Telegram
2. Запустите команду `/start`
3. Откройте Mini App через меню бота
4. Проверьте, что приложение загружается
5. Проверьте аутентификацию (должен определиться ваш Telegram ID)

---

## 🔧 Полезные команды

### Docker

```bash
# Посмотреть логи
docker-compose logs -f backend

# Остановить контейнеры
docker-compose down

# Пересобрать и запустить
docker-compose up -d --build

# Войти в контейнер backend
docker-compose exec backend sh

# Войти в PostgreSQL
docker-compose exec postgres psql -U postgres -d inventory_db
```

### База данных

```bash
# Сброс базы данных
psql -d inventory_db -c "DROP TABLE IF EXISTS transactions, items CASCADE;"

# Повторное применение миграций
psql -d inventory_db -f backend/migrations/001_create_items.sql
psql -d inventory_db -f backend/migrations/002_create_transactions.sql

# Бэкап базы данных
pg_dump inventory_db > backup_$(date +%Y%m%d).sql

# Восстановление из бэкапа
psql -d inventory_db < backup_20260317.sql
```

### Разработка

```bash
# Форматирование Go кода
cd backend
go fmt ./...

# Проверка линтером
golangci-lint run

# Запуск с auto-reload (требуется air)
# go install github.com/cosmtrek/air@latest
air
```

---

## 📖 Документация

- **Полное руководство по тестированию**: [`TESTING_GUIDE.md`](./TESTING_GUIDE.md)
- **Спецификация**: [`specs/001-inventory-management/spec.md`](./specs/001-inventory-management/spec.md)
- **План реализации**: [`specs/001-inventory-management/plan.md`](./specs/001-inventory-management/plan.md)
- **Список задач**: [`specs/001-inventory-management/tasks.md`](./specs/001-inventory-management/tasks.md)

---

## 🎯 Основные endpoint'ы API

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|-------------|
| GET | `/health` | Health check | Нет |
| GET | `/api/items` | Список товаров | Да |
| GET | `/api/items/:id` | Детали товара | Да |
| GET | `/api/items/:id/transactions` | История транзакций | Да |
| POST | `/api/items` | Добавить товар | Да |
| PUT | `/api/items/:id/add` | Добавить на склад | Да |
| PUT | `/api/items/:id/remove` | Списать со склада | Да |

### Параметры фильтрации транзакций

```bash
GET /api/items/:id/transactions?limit=50&offset=0&type=addition&start_date=2026-03-01T00:00:00Z&end_date=2026-03-17T23:59:59Z
```

**Параметры**:
- `limit` - количество записей (по умолчанию 50, макс 100)
- `offset` - смещение для пагинации
- `type` - тип транзакции: `addition` (добавление), `reduction` (списание), пусто = все
- `start_date` - начало диапазона (формат RFC3339)
- `end_date` - конец диапазона (формат RFC3339)

---

## ❓ Решение проблем

### Backend не запускается

```bash
# Проверьте переменные окружения
env | grep DATABASE_URL
env | grep TELEGRAM_BOT_TOKEN

# Проверьте соединение с базой данных
psql $DATABASE_URL -c "SELECT 1"

# Проверьте логи
docker-compose logs backend
```

### База данных недоступна

```bash
# Проверьте, запущен ли PostgreSQL
docker-compose ps postgres

# Перезапустите контейнер
docker-compose restart postgres

# Проверьте подключение
docker-compose exec postgres psql -U postgres -l
```

### Frontend не загружается

```bash
# Проверьте, что статический сервер запущен на порту 8081
lsof -i :8081

# Проверьте CORS настройки в backend
# Backend должен разрешать запросы с frontend домена
```

### Ошибки аутентификации 401

```bash
# Убедитесь, что заголовок X-Telegram-Init-Data передается
# Для локального тестирования используйте любое значение:
curl -H "X-Telegram-Init-Data: test" http://localhost:8080/api/items

# Для продакшена требуется валидный initData от Telegram
```

---

## 🎓 Следующие шаги

1. ✅ Запустите локально и протестируйте базовый функционал
2. ✅ Создайте тестового Telegram бота
3. ✅ Разверните на сервер с HTTPS (обязательно для Telegram)
4. ✅ Настройте webhook и протестируйте в Telegram
5. ✅ Загрузите тестовые фото и проверьте сжатие
6. ✅ Протестируйте с несколькими пользователями одновременно
7. ✅ Настройте автоматические бэкапы базы данных
8. ✅ Настройте мониторинг и алерты

---

## 📞 Поддержка

При возникновении проблем:
1. Проверьте [`TESTING_GUIDE.md`](./TESTING_GUIDE.md) - раздел "Решение проблем"
2. Просмотрите логи: `docker-compose logs -f`
3. Создайте issue на GitHub с описанием проблемы и логами

---

**Версия**: 1.0
**Последнее обновление**: 2026-03-17
