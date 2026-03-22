#!/bin/bash

# Скрипт для автоматического тестирования системы складского учета
# Использование: ./test_system.sh [API_URL]

set -e  # Остановка при ошибке

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Настройки
API_URL="${1:-http://localhost:8080}"
AUTH_HEADER="X-Telegram-Init-Data: test_init_data"
TEMP_DIR="/tmp/inventory_test_$$"

# Счетчики
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

mkdir -p "$TEMP_DIR"

# Функция для вывода результата теста
print_result() {
  local test_name="$1"
  local result="$2"
  local message="$3"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))

  if [ "$result" = "PASS" ]; then
    echo -e "${GREEN}✓ PASS${NC} - $test_name"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}✗ FAIL${NC} - $test_name: $message"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
}

# Функция для выполнения API запроса
api_request() {
  local method="$1"
  local endpoint="$2"
  local data="$3"
  local expected_code="${4:-200}"

  if [ -z "$data" ]; then
    response=$(curl -s -w "\n%{http_code}" -X "$method" \
      -H "$AUTH_HEADER" \
      "$API_URL$endpoint")
  else
    response=$(curl -s -w "\n%{http_code}" -X "$method" \
      -H "$AUTH_HEADER" \
      -H "Content-Type: application/json" \
      -d "$data" \
      "$API_URL$endpoint")
  fi

  http_code=$(echo "$response" | tail -n 1)
  body=$(echo "$response" | sed '$d')

  echo "$body" > "$TEMP_DIR/last_response.json"

  if [ "$http_code" != "$expected_code" ]; then
    return 1
  fi

  return 0
}

echo "======================================"
echo "  Тестирование системы складского учета"
echo "  API URL: $API_URL"
echo "======================================"
echo ""

# ===========================
# БЛОК 1: Проверка доступности
# ===========================

echo "БЛОК 1: Проверка доступности сервисов"
echo "--------------------------------------"

# Тест 1.1: Health check
if api_request GET /health "" 200; then
  print_result "Health check endpoint" "PASS"
else
  print_result "Health check endpoint" "FAIL" "Сервер недоступен"
  echo -e "${RED}Критическая ошибка: Backend не запущен${NC}"
  exit 1
fi

echo ""

# ===========================
# БЛОК 2: Работа с товарами
# ===========================

echo "БЛОК 2: Операции с товарами"
echo "----------------------------"

# Тест 2.1: Получение списка товаров
if api_request GET "/api/items?limit=10&offset=0" "" 200; then
  item_count=$(jq '.items | length' "$TEMP_DIR/last_response.json" 2>/dev/null || echo "0")
  print_result "Получение списка товаров" "PASS"
  echo "  → Получено товаров: $item_count"
else
  print_result "Получение списка товаров" "FAIL" "Ошибка API"
fi

# Тест 2.2: Добавление нового товара
TEST_ITEM_NAME="Тестовый товар $(date +%s)"
if api_request POST /api/items "{\"name\":\"$TEST_ITEM_NAME\",\"initial_quantity\":10}" 201; then
  ITEM_ID=$(jq -r '.id' "$TEMP_DIR/last_response.json")
  print_result "Добавление товара без фото" "PASS"
  echo "  → ID созданного товара: $ITEM_ID"
else
  print_result "Добавление товара без фото" "FAIL" "Не удалось создать товар"
  ITEM_ID=""
fi

# Тест 2.3: Получение деталей товара
if [ -n "$ITEM_ID" ]; then
  if api_request GET "/api/items/$ITEM_ID" "" 200; then
    current_qty=$(jq -r '.item.current_quantity' "$TEMP_DIR/last_response.json")
    print_result "Получение деталей товара" "PASS"
    echo "  → Текущее количество: $current_qty"
  else
    print_result "Получение деталей товара" "FAIL" "Товар не найден"
  fi
else
  echo -e "${YELLOW}⊘ SKIP${NC} - Получение деталей товара (нет ID)"
fi

echo ""

# ===========================
# БЛОК 3: Управление количеством
# ===========================

echo "БЛОК 3: Управление количеством товаров"
echo "---------------------------------------"

if [ -n "$ITEM_ID" ]; then
  # Тест 3.1: Добавление товара на склад
  if api_request PUT "/api/items/$ITEM_ID/add" '{"quantity":5,"note":"Тест добавления"}' 200; then
    new_qty=$(jq -r '.item.current_quantity' "$TEMP_DIR/last_response.json")
    print_result "Добавление товара на склад (+5)" "PASS"
    echo "  → Новое количество: $new_qty (ожидается: 15)"

    if [ "$new_qty" != "15" ]; then
      print_result "Проверка количества после добавления" "FAIL" "Ожидалось 15, получено $new_qty"
    else
      print_result "Проверка количества после добавления" "PASS"
    fi
  else
    print_result "Добавление товара на склад" "FAIL" "Ошибка API"
  fi

  # Тест 3.2: Списание товара со склада
  if api_request PUT "/api/items/$ITEM_ID/remove" '{"quantity":3,"note":"Тест списания"}' 200; then
    new_qty=$(jq -r '.item.current_quantity' "$TEMP_DIR/last_response.json")
    print_result "Списание товара со склада (-3)" "PASS"
    echo "  → Новое количество: $new_qty (ожидается: 12)"

    if [ "$new_qty" != "12" ]; then
      print_result "Проверка количества после списания" "FAIL" "Ожидалось 12, получено $new_qty"
    else
      print_result "Проверка количества после списания" "PASS"
    fi
  else
    print_result "Списание товара со склада" "FAIL" "Ошибка API"
  fi

  # Тест 3.3: Валидация (попытка добавить 0)
  if ! api_request PUT "/api/items/$ITEM_ID/add" '{"quantity":0}' 400; then
    print_result "Валидация: отклонение quantity=0" "PASS"
  else
    print_result "Валидация: отклонение quantity=0" "FAIL" "Принято неверное значение"
  fi

  # Тест 3.4: Списание с переходом в минус
  if api_request PUT "/api/items/$ITEM_ID/remove" '{"quantity":20,"note":"Тест отрицательного остатка"}' 200; then
    new_qty=$(jq -r '.item.current_quantity' "$TEMP_DIR/last_response.json")
    print_result "Разрешение отрицательного остатка" "PASS"
    echo "  → Количество после списания: $new_qty (должно быть отрицательным)"
  else
    print_result "Разрешение отрицательного остатка" "FAIL" "Система не разрешает минус"
  fi
else
  echo -e "${YELLOW}⊘ SKIP${NC} - Блок 3 (нет ID товара)"
fi

echo ""

# ===========================
# БЛОК 4: История транзакций
# ===========================

echo "БЛОК 4: История транзакций"
echo "--------------------------"

if [ -n "$ITEM_ID" ]; then
  # Тест 4.1: Получение истории транзакций
  if api_request GET "/api/items/$ITEM_ID/transactions?limit=50" "" 200; then
    txn_count=$(jq '.transactions | length' "$TEMP_DIR/last_response.json")
    print_result "Получение истории транзакций" "PASS"
    echo "  → Количество транзакций: $txn_count (ожидается ≥4)"

    if [ "$txn_count" -ge 4 ]; then
      print_result "Проверка количества транзакций" "PASS"
    else
      print_result "Проверка количества транзакций" "FAIL" "Мало транзакций"
    fi
  else
    print_result "Получение истории транзакций" "FAIL" "Ошибка API"
  fi

  # Тест 4.2: Фильтрация по типу (только добавления)
  if api_request GET "/api/items/$ITEM_ID/transactions?type=addition" "" 200; then
    additions=$(jq '[.transactions[] | select(.delta > 0)] | length' "$TEMP_DIR/last_response.json")
    total=$(jq '.transactions | length' "$TEMP_DIR/last_response.json")

    if [ "$additions" -eq "$total" ]; then
      print_result "Фильтрация по типу (addition)" "PASS"
      echo "  → Все транзакции - добавления: $additions"
    else
      print_result "Фильтрация по типу (addition)" "FAIL" "Фильтр не работает"
    fi
  else
    print_result "Фильтрация по типу (addition)" "FAIL" "Ошибка API"
  fi

  # Тест 4.3: Фильтрация по типу (только списания)
  if api_request GET "/api/items/$ITEM_ID/transactions?type=reduction" "" 200; then
    reductions=$(jq '[.transactions[] | select(.delta < 0)] | length' "$TEMP_DIR/last_response.json")
    total=$(jq '.transactions | length' "$TEMP_DIR/last_response.json")

    if [ "$reductions" -eq "$total" ]; then
      print_result "Фильтрация по типу (reduction)" "PASS"
      echo "  → Все транзакции - списания: $reductions"
    else
      print_result "Фильтрация по типу (reduction)" "FAIL" "Фильтр не работает"
    fi
  else
    print_result "Фильтрация по типу (reduction)" "FAIL" "Ошибка API"
  fi

  # Тест 4.4: Фильтрация по дате
  START_DATE="2026-01-01T00:00:00Z"
  END_DATE="2026-12-31T23:59:59Z"

  if api_request GET "/api/items/$ITEM_ID/transactions?start_date=$START_DATE&end_date=$END_DATE" "" 200; then
    filtered=$(jq '.filtered' "$TEMP_DIR/last_response.json")

    if [ "$filtered" = "true" ]; then
      print_result "Фильтрация по дате" "PASS"
    else
      print_result "Фильтрация по дате" "FAIL" "Флаг filtered=false"
    fi
  else
    print_result "Фильтрация по дате" "FAIL" "Ошибка API"
  fi
else
  echo -e "${YELLOW}⊘ SKIP${NC} - Блок 4 (нет ID товара)"
fi

echo ""

# ===========================
# БЛОК 5: Безопасность
# ===========================

echo "БЛОК 5: Безопасность и авторизация"
echo "-----------------------------------"

# Тест 5.1: Запрос без авторизации
if ! curl -s -o /dev/null -w "%{http_code}" "$API_URL/api/items" | grep -q "401"; then
  print_result "Отклонение запроса без авторизации" "FAIL" "Принят неавторизованный запрос"
else
  print_result "Отклонение запроса без авторизации" "PASS"
fi

# Тест 5.2: Запрос несуществующего товара (404)
if ! api_request GET "/api/items/00000000-0000-0000-0000-000000000000" "" 404; then
  print_result "404 для несуществующего товара" "FAIL" "Неверный код ответа"
else
  print_result "404 для несуществующего товара" "PASS"
fi

echo ""

# ===========================
# БЛОК 6: Производительность
# ===========================

echo "БЛОК 6: Производительность"
echo "--------------------------"

# Тест 6.1: Время ответа для чтения
ITERATIONS=5
TOTAL_TIME=0

for i in $(seq 1 $ITERATIONS); do
  START=$(date +%s%N)
  api_request GET "/api/items?limit=50" "" 200 > /dev/null 2>&1
  END=$(date +%s%N)

  DIFF=$(( (END - START) / 1000000 ))  # миллисекунды
  TOTAL_TIME=$((TOTAL_TIME + DIFF))
done

AVG_TIME=$((TOTAL_TIME / ITERATIONS))

if [ $AVG_TIME -lt 200 ]; then
  print_result "Время ответа чтения (<200ms)" "PASS"
  echo "  → Среднее время: ${AVG_TIME}ms"
else
  print_result "Время ответа чтения (<200ms)" "FAIL" "Среднее: ${AVG_TIME}ms"
fi

echo ""

# ===========================
# ИТОГИ
# ===========================

echo "======================================"
echo "  ИТОГИ ТЕСТИРОВАНИЯ"
echo "======================================"
echo ""
echo "Всего тестов: $TESTS_TOTAL"
echo -e "${GREEN}Пройдено: $TESTS_PASSED${NC}"
echo -e "${RED}Не пройдено: $TESTS_FAILED${NC}"
echo ""

SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
echo "Процент успешных: $SUCCESS_RATE%"

# Очистка
rm -rf "$TEMP_DIR"

# Финальная оценка
if [ $TESTS_FAILED -eq 0 ]; then
  echo -e "${GREEN}✓ ВСЕ ТЕСТЫ ПРОЙДЕНЫ${NC}"
  echo "Система готова к использованию!"
  exit 0
elif [ $SUCCESS_RATE -ge 80 ]; then
  echo -e "${YELLOW}⚠ ЧАСТИЧНО РАБОТАЕТ${NC}"
  echo "Система работает, но есть проблемы. Требуется проверка."
  exit 1
else
  echo -e "${RED}✗ КРИТИЧЕСКИЕ ПРОБЛЕМЫ${NC}"
  echo "Система не готова к использованию. Требуется отладка."
  exit 2
fi
