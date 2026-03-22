#!/bin/bash

# Скрипт для сброса системы в начальное состояние
# Использование: ./reset_system.sh

set -e

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "======================================"
echo "  Сброс системы складского учета"
echo "======================================"
echo ""

# Проверка подтверждения
read -p "⚠️  Это удалит ВСЕ данные. Продолжить? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
  echo "Отменено."
  exit 0
fi

echo ""
echo "Начинаем сброс..."
echo ""

# Определяем, используется ли Docker
if [ -f "deployments/docker-compose.yml" ]; then
  USE_DOCKER=true
  echo "→ Обнаружен Docker Compose"
else
  USE_DOCKER=false
  echo "→ Docker не используется, работаем с локальной БД"
fi

# ===========================
# Остановка сервисов
# ===========================

echo ""
echo "1. Остановка сервисов..."

if [ "$USE_DOCKER" = true ]; then
  cd deployments
  docker-compose down -v
  echo -e "${GREEN}✓${NC} Docker контейнеры остановлены"
  cd ..
else
  # Пытаемся остановить локальный сервер (если запущен)
  pkill -f "go run cmd/server/main.go" 2>/dev/null || true
  echo -e "${GREEN}✓${NC} Локальные процессы остановлены"
fi

# ===========================
# Очистка базы данных
# ===========================

echo ""
echo "2. Очистка базы данных..."

if [ "$USE_DOCKER" = true ]; then
  # База будет пересоздана при запуске Docker Compose
  echo -e "${GREEN}✓${NC} База данных будет пересоздана"
else
  # Локальная база данных
  DB_NAME="${DB_NAME:-inventory_db}"

  if psql -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo "  → Удаление старой базы данных..."
    dropdb "$DB_NAME" 2>/dev/null || true

    echo "  → Создание новой базы данных..."
    createdb "$DB_NAME"

    echo "  → Применение миграций..."
    psql -d "$DB_NAME" -f backend/migrations/001_create_items.sql
    psql -d "$DB_NAME" -f backend/migrations/002_create_transactions.sql

    echo -e "${GREEN}✓${NC} База данных пересоздана"
  else
    echo "  → Создание базы данных..."
    createdb "$DB_NAME"

    echo "  → Применение миграций..."
    psql -d "$DB_NAME" -f backend/migrations/001_create_items.sql
    psql -d "$DB_NAME" -f backend/migrations/002_create_transactions.sql

    echo -e "${GREEN}✓${NC} База данных создана"
  fi
fi

# ===========================
# Очистка временных файлов
# ===========================

echo ""
echo "3. Очистка временных файлов..."

# Удаление скомпилированных файлов Go
find backend -name "*.test" -delete 2>/dev/null || true
find backend -name "*.out" -delete 2>/dev/null || true
rm -f backend/server 2>/dev/null || true

# Удаление логов (если есть)
rm -rf logs/*.log 2>/dev/null || true

echo -e "${GREEN}✓${NC} Временные файлы удалены"

# ===========================
# Перезапуск системы (опционально)
# ===========================

echo ""
read -p "Запустить систему заново? (yes/no): " START_AGAIN

if [ "$START_AGAIN" = "yes" ]; then
  echo ""
  echo "4. Запуск системы..."

  if [ "$USE_DOCKER" = true ]; then
    cd deployments
    docker-compose up -d

    echo "  → Ожидание запуска контейнеров..."
    sleep 5

    # Применение миграций в Docker
    docker-compose exec -T backend sh -c "psql \$DATABASE_URL -f /app/migrations/001_create_items.sql" 2>/dev/null || echo "    Миграции уже применены"
    docker-compose exec -T backend sh -c "psql \$DATABASE_URL -f /app/migrations/002_create_transactions.sql" 2>/dev/null || echo "    Миграции уже применены"

    cd ..
    echo -e "${GREEN}✓${NC} Система запущена в Docker"
  else
    # Запуск локального сервера в фоне
    cd backend
    nohup go run cmd/server/main.go > ../logs/server.log 2>&1 &
    SERVER_PID=$!
    echo "  → Backend запущен (PID: $SERVER_PID)"
    cd ..

    echo -e "${GREEN}✓${NC} Система запущена локально"
  fi

  # Проверка доступности
  echo ""
  echo "  → Проверка доступности API..."
  sleep 3

  if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo -e "${GREEN}✓ API доступен${NC}"
  else
    echo -e "${YELLOW}⚠ API пока недоступен, подождите немного${NC}"
  fi
fi

# ===========================
# Вставка тестовых данных (опционально)
# ===========================

echo ""
read -p "Добавить тестовые данные? (yes/no): " ADD_TEST_DATA

if [ "$ADD_TEST_DATA" = "yes" ]; then
  echo ""
  echo "5. Добавление тестовых данных..."

  # Создаем временный файл с тестовыми данными
  cat > /tmp/test_data.sql << 'EOF'
-- Тестовые товары
INSERT INTO items (name, current_quantity, created_by) VALUES
('Ноутбук Dell XPS 15', 10, 12345),
('Монитор Samsung 27"', 25, 12345),
('Клавиатура Logitech', 50, 12345),
('Мышь беспроводная', 100, 12345),
('Кабель HDMI 2м', 0, 12345),
('MacBook Pro 16"', 5, 12345),
('iPhone 14 Pro', 15, 12345),
('iPad Air', 8, 12345),
('AirPods Pro', 30, 12345),
('Magic Mouse', 20, 12345);

-- Начальные транзакции
INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
SELECT id, 12345, 0, current_quantity, current_quantity, 'Начальный остаток'
FROM items;

-- Дополнительные транзакции для истории
DO $$
DECLARE
    item_record RECORD;
BEGIN
    FOR item_record IN SELECT id, current_quantity FROM items WHERE current_quantity > 0 LIMIT 5 LOOP
        INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
        VALUES (item_record.id, 12345, item_record.current_quantity, item_record.current_quantity + 10, 10, 'Поступление от поставщика');

        INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
        VALUES (item_record.id, 12345, item_record.current_quantity + 10, item_record.current_quantity + 5, -5, 'Выдано на офис');
    END LOOP;
END $$;
EOF

  if [ "$USE_DOCKER" = true ]; then
    docker-compose exec -T postgres psql -U postgres -d inventory_db < /tmp/test_data.sql
  else
    psql -d inventory_db < /tmp/test_data.sql
  fi

  rm /tmp/test_data.sql

  echo -e "${GREEN}✓${NC} Тестовые данные добавлены"
  echo "  → Добавлено 10 товаров"
  echo "  → Создана история транзакций"
fi

# ===========================
# Финал
# ===========================

echo ""
echo "======================================"
echo -e "${GREEN}✓ СБРОС ЗАВЕРШЕН${NC}"
echo "======================================"
echo ""

if [ "$START_AGAIN" = "yes" ]; then
  echo "Система запущена и готова к использованию:"
  echo "  → Backend API: http://localhost:8080"
  echo "  → Health Check: http://localhost:8080/health"
  echo ""
  echo "Для тестирования системы запустите:"
  echo "  ./test_system.sh"
fi

echo ""
echo "Для просмотра документации:"
echo "  → Быстрый старт: QUICKSTART_RU.md"
echo "  → Руководство по тестированию: TESTING_GUIDE.md"
echo "  → Примеры API: API_EXAMPLES.md"
echo ""
