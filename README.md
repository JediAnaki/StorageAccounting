<div align="center">

# 📦 Складской Учёт

### Система управления складом через Telegram Mini App

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Telegram](https://img.shields.io/badge/Telegram-Mini_App-26A5E4?style=for-the-badge&logo=telegram)](https://core.telegram.org/bots/webapps)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)

**Современное решение для учёта товаров на складе с фото-карточками, историей движений и мобильным интерфейсом**

[Демо](#-демонстрация) • [Возможности](#-возможности) • [Быстрый старт](#-быстрый-старт) • [API](#-api-документация)

---

<!-- Место для GIF/видео/скриншота -->
### 🎬 Демонстрация

> 📸 _Здесь будет размещена анимация работы приложения_

```
┌─────────────────────────────────────┐
│                                     │
│     [ Здесь ваша GIF-анимация ]     │
│                                     │
│  или скриншот интерфейса приложения │
│                                     │
└─────────────────────────────────────┘
```

</div>

---

## 🎯 О Проекте

**Складской Учёт** — это Telegram Mini App для управления товарами на складе производственного предприятия. Приложение позволяет вести учёт товаров с фотографиями, отслеживать все движения (поступления и убавления) и вести полный аудит изменений количества.

### 🎨 Для какой платформы

- 📱 **Telegram Mini App** — работает прямо в Telegram на iOS, Android и Desktop
- 🌐 **Web-интерфейс** — доступен через браузер с мобильной оптимизацией
- 🐳 **Docker** — готовая контейнеризация для развёртывания на сервере

### ⚡ Технологический Стек

<table>
<tr>
<td>

**Backend**
- 🐹 Go 1.25
- 🗄️ PostgreSQL 16
- 🔄 ACID транзакции
- 📸 Сжатие изображений

</td>
<td>

**Frontend**
- 🎨 Vanilla JavaScript
- 💬 Telegram WebApp SDK
- 📱 Mobile-First дизайн
- 🌓 Тёмная/светлая тема

</td>
<td>

**Deployment**
- 🐳 Docker + Compose
- 🔒 nginx + HTTPS
- 📦 Telegram File API
- ☁️ S3-совместимое хранилище

</td>
</tr>
</table>

---

## ✨ Возможности

<table>
<tr>
<td width="33%">

### 📋 Управление товарами
- Просмотр товаров с фото
- Карточки с текущим остатком
- Быстрый поиск и фильтры
- Визуальное отображение нулевых остатков

</td>
<td width="33%">

### 📊 Учёт движений
- Поступления (+) и убавления (-)
- Полная история изменений
- Пометки и комментарии
- Защита от конфликтов данных

</td>
<td width="33%">

### 🔐 Безопасность
- Telegram аутентификация
- Привязка к пользователю
- Immutable аудит-лог
- Автоматические бэкапы

</td>
</tr>
</table>

---

## 📁 Структура Проекта

```
📦 StorageAccounting
┣ 📂 backend/                 # Backend сервис на Go
┃ ┣ 📂 cmd/server/            # Точка входа приложения
┃ ┣ 📂 internal/              # Приватный код приложения
┃ ┃ ┣ 📂 api/                 # HTTP обработчики и middleware
┃ ┃ ┣ 📂 models/              # Модели данных
┃ ┃ ┣ 📂 repository/          # Слой доступа к данным
┃ ┃ ┣ 📂 service/             # Бизнес-логика
┃ ┃ ┗ 📂 telegram/            # Интеграция с Telegram Bot API
┃ ┗ 📂 migrations/            # Миграции базы данных
┣ 📂 frontend/                # Telegram Mini App
┃ ┣ 📂 css/                   # Стили
┃ ┣ 📂 js/                    # JavaScript модули
┃ ┗ 📂 assets/                # Статические ресурсы
┣ 📂 deployments/             # Docker и конфигурация деплоя
┗ 📂 tests/                   # Интеграционные и unit-тесты
```

---

## 🚀 Быстрый Старт

### 📋 Требования

- 🐹 **Go 1.25+**
- 🐳 **Docker** и **Docker Compose**
- 🗄️ **PostgreSQL 16+**
- 🤖 **Telegram Bot Token**

### 💻 Локальная Разработка

#### 1️⃣ Клонирование репозитория

```bash
git clone https://github.com/ваш-username/StorageAccounting.git
cd StorageAccounting
```

**Вывод:**
```
Cloning into 'StorageAccounting'...
remote: Enumerating objects: 45, done.
remote: Counting objects: 100% (45/45), done.
✅ Repository successfully cloned
```

#### 2️⃣ Настройка окружения

```bash
cp .env.example .env
nano .env  # или используйте ваш редактор
```

**Файл .env должен содержать:**
```env
PORT=8080
DATABASE_URL=postgres://storage_user:storage_password@localhost:5432/storage_accounting
TELEGRAM_BOT_TOKEN=your_bot_token_here
S3_ENDPOINT=https://s3.example.com
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=storage-photos
```

#### 3️⃣ Запуск через Docker

```bash
cd deployments
docker-compose up -d
```

**Вывод:**
```
[+] Running 3/3
 ✔ Container storage_accounting_db       Started
 ✔ Container storage_accounting_backend  Started
 ✔ Container storage_accounting_nginx    Started
```

#### 4️⃣ Проверка работы

```bash
curl http://localhost:8080/health
```

**Вывод:**
```json
{"status":"ok"}
```

### 🎯 Доступ к Приложению

| Сервис | URL | Описание |
|--------|-----|----------|
| 🌐 **Frontend** | `http://localhost` | Telegram Mini App интерфейс |
| 🔧 **Backend API** | `http://localhost:8080` | REST API сервер |
| ❤️ **Health Check** | `http://localhost:8080/health` | Проверка состояния |

---

## 🛠️ Разработка без Docker

### 1️⃣ Запуск PostgreSQL

```bash
psql -U postgres
CREATE DATABASE storage_accounting;
CREATE USER storage_user WITH PASSWORD 'storage_password';
GRANT ALL PRIVILEGES ON DATABASE storage_accounting TO storage_user;
```

### 2️⃣ Запуск Backend

```bash
cd backend
go run cmd/server/main.go
```

**Вывод:**
```
2025/03/16 12:00:00 Server starting on :8080
2025/03/16 12:00:00 ✅ Database connected
2025/03/16 12:00:00 ✅ Telegram bot initialized
```

### 3️⃣ Запуск Frontend

```bash
# Используйте любой статический сервер, например:
python3 -m http.server 3000 --directory frontend
```

---

## 📖 API Документация

### 🔑 Аутентификация

Все API endpoints требуют Telegram WebApp аутентификации через заголовок:
```
X-Telegram-Init-Data: <telegram_init_data>
```

### 📡 Endpoints

<details>
<summary><b>GET /health</b> — Проверка состояния сервера</summary>

**Запрос:**
```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{"status":"ok"}
```
</details>

<details>
<summary><b>GET /api/items</b> — Получить список товаров</summary>

**Параметры:**
- `limit` (optional) — количество записей (default: 50)
- `offset` (optional) — смещение (default: 0)

**Запрос:**
```bash
curl -H "X-Telegram-Init-Data: ..." \
  http://localhost:8080/api/items?limit=10&offset=0
```

**Ответ:**
```json
[
  {
    "id": 1,
    "name": "Винт М8х20",
    "current_quantity": 1500,
    "photo_url": "https://...",
    "created_at": "2025-03-16T12:00:00Z"
  }
]
```
</details>

<details>
<summary><b>GET /api/items/:id</b> — Получить детали товара</summary>

**Запрос:**
```bash
curl -H "X-Telegram-Init-Data: ..." \
  http://localhost:8080/api/items/1
```

**Ответ:**
```json
{
  "id": 1,
  "name": "Винт М8х20",
  "current_quantity": 1500,
  "photo_url": "https://...",
  "transactions": [...]
}
```
</details>

<details>
<summary><b>POST /api/items</b> — Создать новый товар</summary>

**Запрос:**
```bash
curl -X POST -H "X-Telegram-Init-Data: ..." \
  -F "name=Гайка М8" \
  -F "quantity=1000" \
  -F "photo=@photo.jpg" \
  http://localhost:8080/api/items
```

**Ответ:**
```json
{
  "id": 2,
  "name": "Гайка М8",
  "current_quantity": 1000,
  "photo_url": "https://...",
  "created_at": "2025-03-16T12:05:00Z"
}
```
</details>

<details>
<summary><b>PUT /api/items/:id/quantity</b> — Изменить количество</summary>

**Тело запроса:**
```json
{
  "delta": 50,
  "note": "Поступление от поставщика ООО Метизы"
}
```

**Запрос:**
```bash
curl -X PUT -H "X-Telegram-Init-Data: ..." \
  -H "Content-Type: application/json" \
  -d '{"delta":50,"note":"Поступление"}' \
  http://localhost:8080/api/items/1/quantity
```

**Ответ:**
```json
{
  "id": 1,
  "current_quantity": 1550,
  "transaction_id": 15
}
```
</details>

---

## 🧪 Тестирование

```bash
cd backend
go test ./... -v
```

**Ожидаемый вывод:**
```
=== RUN   TestItemRepository
--- PASS: TestItemRepository (0.15s)
=== RUN   TestInventoryService
--- PASS: TestInventoryService (0.23s)
PASS
ok      storageaccounting/internal/repository    0.456s
```

---

## 📦 Деплой в Production

### 🐳 Docker Deployment

```bash
cd deployments
docker-compose -f docker-compose.yml up -d --build
```

### ⚙️ Production Checklist

- [ ] Настроить HTTPS с валидными SSL сертификатами
- [ ] Обновить `nginx.conf` — включить HTTPS server block
- [ ] Настроить автоматические бэкапы PostgreSQL
- [ ] Сконфигурировать S3 хранилище для production
- [ ] Включить Telegram webhook для бота
- [ ] Настроить мониторинг и логирование
- [ ] Установить rate limiting на API

---

## 🤝 Контрибуция

Проект находится в активной разработке. Pull requests приветствуются!

---

## 📄 Лицензия

Proprietary

---

<div align="center">

**Сделано с ❤️ для производственных складов**

⭐ Поставьте звезду, если проект был полезен!

</div>
