# Примеры API запросов

Готовые команды для тестирования API складской системы.

---

## 🔧 Настройка

### Переменные для удобства

```bash
# Базовый URL API
export API_URL="http://localhost:8080"

# Для тестирования без реальной авторизации используйте любой initData
export AUTH_HEADER="X-Telegram-Init-Data: test_init_data"

# Для продакшена получите реальный initData от Telegram WebApp
# export AUTH_HEADER="X-Telegram-Init-Data: query_id=AAHd..."
```

---

## 📝 Базовые операции

### 1. Health Check (без авторизации)

```bash
# Проверка работоспособности сервера
curl -X GET "$API_URL/health"
```

**Ожидаемый ответ**:
```json
{
  "status": "healthy"
}
```

---

## 📦 Работа с товарами

### 2. Получить список товаров

```bash
# Получить первые 50 товаров
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items?limit=50&offset=0"
```

**Параметры**:
- `limit` - количество товаров (по умолчанию 50, макс 100)
- `offset` - смещение для пагинации

**Пример ответа**:
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Ноутбук Dell XPS 15",
      "current_quantity": 10,
      "photo_file_id": null,
      "photo_s3_key": null,
      "created_at": "2026-03-17T10:00:00Z",
      "updated_at": "2026-03-17T10:00:00Z",
      "deleted_at": null,
      "created_by": 12345,
      "version": 1
    }
  ],
  "total_count": 8,
  "limit": 50,
  "offset": 0
}
```

---

### 3. Получить детали товара

```bash
# Замените {ITEM_ID} на реальный ID товара
ITEM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID"
```

**Пример ответа**:
```json
{
  "item": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ноутбук Dell XPS 15",
    "current_quantity": 10,
    "created_at": "2026-03-17T10:00:00Z",
    "version": 1
  },
  "transactions": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "item_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": 12345,
      "timestamp": "2026-03-17T10:00:00Z",
      "previous_quantity": 0,
      "new_quantity": 10,
      "delta": 10,
      "note": "Начальный остаток"
    }
  ],
  "total_count": 1
}
```

---

### 4. Добавить новый товар (без фото)

```bash
curl -X POST \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Тестовый товар",
    "initial_quantity": 15
  }' \
  "$API_URL/api/items"
```

**Пример ответа**:
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "Тестовый товар",
  "current_quantity": 15,
  "created_at": "2026-03-17T12:00:00Z",
  "version": 1
}
```

---

### 5. Добавить товар с фото

```bash
# Добавление товара с фото через multipart/form-data
curl -X POST \
  -H "$AUTH_HEADER" \
  -F "name=Товар с фото" \
  -F "initial_quantity=20" \
  -F "photo=@/path/to/photo.jpg" \
  "$API_URL/api/items"
```

**Примечание**: Замените `/path/to/photo.jpg` на реальный путь к фото.

---

## 📊 Управление количеством

### 6. Добавить товар на склад

```bash
# Добавить 5 единиц товара
ITEM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 5,
    "note": "Поступление от поставщика ABC"
  }' \
  "$API_URL/api/items/$ITEM_ID/add"
```

**Пример ответа**:
```json
{
  "success": true,
  "item": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ноутбук Dell XPS 15",
    "current_quantity": 15,
    "version": 2
  },
  "message": "stock added successfully"
}
```

---

### 7. Списать товар со склада

```bash
# Списать 3 единицы товара
ITEM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 3,
    "note": "Выдано на офис, корпус A"
  }' \
  "$API_URL/api/items/$ITEM_ID/remove"
```

**Пример ответа**:
```json
{
  "success": true,
  "item": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ноутбук Dell XPS 15",
    "current_quantity": 12,
    "version": 3
  },
  "message": "stock removed successfully"
}
```

---

### 8. Списание с переходом в минус (с предупреждением)

```bash
# Попытка списать больше, чем есть на складе
ITEM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 20,
    "note": "Списание по акту"
  }' \
  "$API_URL/api/items/$ITEM_ID/remove"
```

**Результат**: Количество уйдет в минус (например, 12 - 20 = -8). Система разрешает отрицательный остаток.

---

## 📈 История транзакций

### 9. Получить историю транзакций товара

```bash
# Все транзакции товара
ITEM_ID="550e8400-e29b-41d4-a716-446655440000"

curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?limit=50&offset=0"
```

**Пример ответа**:
```json
{
  "transactions": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "item_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": 12345,
      "timestamp": "2026-03-17T14:30:00Z",
      "previous_quantity": 15,
      "new_quantity": 12,
      "delta": -3,
      "note": "Выдано на офис, корпус A"
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "item_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": 12345,
      "timestamp": "2026-03-17T12:00:00Z",
      "previous_quantity": 10,
      "new_quantity": 15,
      "delta": 5,
      "note": "Поступление от поставщика ABC"
    }
  ],
  "total_count": 2,
  "filtered": false,
  "limit": 50,
  "offset": 0
}
```

---

### 10. Фильтрация транзакций по типу

```bash
# Только добавления (положительные транзакции)
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?type=addition"

# Только списания (отрицательные транзакции)
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?type=reduction"
```

---

### 11. Фильтрация транзакций по дате

```bash
# Транзакции за март 2026
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?start_date=2026-03-01T00:00:00Z&end_date=2026-03-31T23:59:59Z"

# Транзакции за последние 7 дней
START_DATE=$(date -u -v-7d +"%Y-%m-%dT00:00:00Z")  # macOS
END_DATE=$(date -u +"%Y-%m-%dT23:59:59Z")

curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?start_date=$START_DATE&end_date=$END_DATE"
```

**Формат даты**: RFC3339 (например, `2026-03-17T12:00:00Z`)

---

### 12. Комбинированные фильтры

```bash
# Только добавления за март 2026
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/$ITEM_ID/transactions?type=addition&start_date=2026-03-01T00:00:00Z&end_date=2026-03-31T23:59:59Z"
```

---

## ⚠️ Обработка ошибок

### 13. Конкурентное обновление (409 Conflict)

```bash
# Симуляция конфликта оптимистической блокировки
# 1. Получите текущую версию товара
curl -X GET -H "$AUTH_HEADER" "$API_URL/api/items/$ITEM_ID"

# 2. Одновременно из двух терминалов попробуйте изменить количество
# Одна из операций получит ошибку 409

curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"quantity": 10}' \
  "$API_URL/api/items/$ITEM_ID/add"
```

**Пример ответа при конфликте**:
```json
{
  "error": "Conflict",
  "message": "item was modified by another user, please refresh and try again",
  "code": 409
}
```

---

### 14. Ошибка валидации (400 Bad Request)

```bash
# Попытка добавить 0 или отрицательное количество
curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"quantity": 0}' \
  "$API_URL/api/items/$ITEM_ID/add"
```

**Пример ответа**:
```json
{
  "error": "Bad Request",
  "message": "quantity must be positive",
  "code": 400
}
```

---

### 15. Товар не найден (404 Not Found)

```bash
# Запрос несуществующего товара
curl -X GET \
  -H "$AUTH_HEADER" \
  "$API_URL/api/items/00000000-0000-0000-0000-000000000000"
```

**Пример ответа**:
```json
{
  "error": "Not Found",
  "message": "item not found",
  "code": 404
}
```

---

### 16. Ошибка авторизации (401 Unauthorized)

```bash
# Запрос без заголовка авторизации
curl -X GET "$API_URL/api/items"
```

**Пример ответа**:
```json
{
  "error": "Unauthorized",
  "message": "missing authentication header",
  "code": 401
}
```

---

## 🧪 Тестовые сценарии

### Сценарий 1: Полный цикл жизни товара

```bash
# 1. Добавить новый товар
RESPONSE=$(curl -s -X POST \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"name":"Тестовый ноутбук","initial_quantity":10}' \
  "$API_URL/api/items")

# Извлечь ID товара из ответа
ITEM_ID=$(echo $RESPONSE | jq -r '.id')
echo "Создан товар с ID: $ITEM_ID"

# 2. Просмотреть детали
curl -X GET -H "$AUTH_HEADER" "$API_URL/api/items/$ITEM_ID"

# 3. Добавить 5 единиц
curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"quantity":5,"note":"Поступление"}' \
  "$API_URL/api/items/$ITEM_ID/add"

# 4. Списать 3 единицы
curl -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"quantity":3,"note":"Выдано"}' \
  "$API_URL/api/items/$ITEM_ID/remove"

# 5. Просмотреть историю
curl -X GET -H "$AUTH_HEADER" "$API_URL/api/items/$ITEM_ID/transactions"

# Ожидаемый результат: 3 транзакции (начальный остаток, +5, -3)
# Финальное количество: 10 + 5 - 3 = 12
```

---

### Сценарий 2: Массовое добавление товаров

```bash
#!/bin/bash
# Скрипт для массового добавления тестовых товаров

ITEMS=(
  "Ноутбук HP Pavilion:15"
  "Монитор LG 24\":20"
  "Принтер Canon:5"
  "Роутер TP-Link:30"
  "USB кабель Type-C:100"
)

for item in "${ITEMS[@]}"; do
  NAME="${item%%:*}"
  QTY="${item##*:}"

  echo "Добавляем: $NAME (количество: $QTY)"

  curl -s -X POST \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$NAME\",\"initial_quantity\":$QTY}" \
    "$API_URL/api/items" | jq '.id'

  sleep 0.5
done

echo "Добавлено ${#ITEMS[@]} товаров"
```

---

### Сценарий 3: Проверка пагинации

```bash
#!/bin/bash
# Получить все товары с пагинацией

LIMIT=10
OFFSET=0
TOTAL=0

while true; do
  RESPONSE=$(curl -s -H "$AUTH_HEADER" \
    "$API_URL/api/items?limit=$LIMIT&offset=$OFFSET")

  ITEMS_COUNT=$(echo $RESPONSE | jq '.items | length')
  TOTAL_COUNT=$(echo $RESPONSE | jq '.total_count')

  echo "Страница: offset=$OFFSET, получено=$ITEMS_COUNT, всего=$TOTAL_COUNT"

  if [ $ITEMS_COUNT -eq 0 ]; then
    break
  fi

  TOTAL=$((TOTAL + ITEMS_COUNT))
  OFFSET=$((OFFSET + LIMIT))
done

echo "Всего получено товаров: $TOTAL"
```

---

### Сценарий 4: Тест производительности

```bash
#!/bin/bash
# Тест времени ответа API

ENDPOINT="$API_URL/api/items?limit=50"
ITERATIONS=10

echo "Тестируем производительность: $ENDPOINT"
echo "Количество запросов: $ITERATIONS"

TOTAL_TIME=0

for i in $(seq 1 $ITERATIONS); do
  START=$(date +%s%N)

  curl -s -H "$AUTH_HEADER" "$ENDPOINT" > /dev/null

  END=$(date +%s%N)
  DIFF=$(( (END - START) / 1000000 ))  # миллисекунды

  TOTAL_TIME=$((TOTAL_TIME + DIFF))

  echo "Запрос $i: ${DIFF}ms"
done

AVG_TIME=$((TOTAL_TIME / ITERATIONS))
echo "Среднее время: ${AVG_TIME}ms"

# Проверка критерия: < 200ms
if [ $AVG_TIME -lt 200 ]; then
  echo "✅ PASS: Производительность соответствует требованиям (<200ms)"
else
  echo "❌ FAIL: Производительность ниже требований (>200ms)"
fi
```

---

## 📋 Postman Collection

Для импорта в Postman создайте файл `postman_collection.json`:

```json
{
  "info": {
    "name": "Inventory Management API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "auth_header",
      "value": "test_init_data"
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/health",
          "host": ["{{base_url}}"],
          "path": ["health"]
        }
      }
    },
    {
      "name": "Get Items",
      "request": {
        "method": "GET",
        "header": [
          {
            "key": "X-Telegram-Init-Data",
            "value": "{{auth_header}}"
          }
        ],
        "url": {
          "raw": "{{base_url}}/api/items?limit=50&offset=0",
          "host": ["{{base_url}}"],
          "path": ["api", "items"],
          "query": [
            {"key": "limit", "value": "50"},
            {"key": "offset", "value": "0"}
          ]
        }
      }
    },
    {
      "name": "Add Stock",
      "request": {
        "method": "PUT",
        "header": [
          {
            "key": "X-Telegram-Init-Data",
            "value": "{{auth_header}}"
          },
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"quantity\": 5,\n  \"note\": \"Тест добавления\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/api/items/:id/add",
          "host": ["{{base_url}}"],
          "path": ["api", "items", ":id", "add"],
          "variable": [
            {"key": "id", "value": "ITEM_ID_HERE"}
          ]
        }
      }
    }
  ]
}
```

---

## 🐛 Отладка

### Включить verbose режим в curl

```bash
# Показать заголовки запроса и ответа
curl -v -X GET -H "$AUTH_HEADER" "$API_URL/api/items"

# Показать только заголовки ответа
curl -I -X GET -H "$AUTH_HEADER" "$API_URL/api/items"

# Сохранить ответ в файл
curl -X GET -H "$AUTH_HEADER" "$API_URL/api/items" -o response.json

# Показать время выполнения
curl -w "@curl-format.txt" -X GET -H "$AUTH_HEADER" "$API_URL/api/items"
```

Создайте файл `curl-format.txt`:
```
    time_namelookup:  %{time_namelookup}s\n
       time_connect:  %{time_connect}s\n
    time_appconnect:  %{time_appconnect}s\n
   time_pretransfer:  %{time_pretransfer}s\n
      time_redirect:  %{time_redirect}s\n
 time_starttransfer:  %{time_starttransfer}s\n
                    ----------\n
         time_total:  %{time_total}s\n
```

---

## 📚 Дополнительные ресурсы

- **Тестирование**: [`TESTING_GUIDE.md`](./TESTING_GUIDE.md)
- **Быстрый старт**: [`QUICKSTART_RU.md`](./QUICKSTART_RU.md)
- **Спецификация API**: см. код в `backend/internal/api/handlers.go`

---

**Версия**: 1.0
**Последнее обновление**: 2026-03-17
