# Storage Accounting

Warehouse inventory management system via Telegram Mini App.

## Features

- View inventory items with photos and quantities
- Track quantity changes (additions and reductions)
- Complete transaction audit trail
- Mobile-optimized interface
- Photo compression and storage

## Tech Stack

- **Backend**: Go 1.25
- **Frontend**: Vanilla JavaScript with Telegram WebApp SDK
- **Database**: PostgreSQL
- **Storage**: Telegram File API + S3-compatible fallback
- **Deployment**: Docker + nginx

## Project Structure

```
.
├── backend/              # Go backend service
│   ├── cmd/server/       # Application entry point
│   ├── internal/         # Private application code
│   │   ├── api/          # HTTP handlers and middleware
│   │   ├── models/       # Data models
│   │   ├── repository/   # Data access layer
│   │   ├── service/      # Business logic
│   │   └── telegram/     # Telegram Bot API integration
│   └── migrations/       # Database migrations
├── frontend/             # Telegram Mini App
│   ├── css/              # Styles
│   ├── js/               # JavaScript modules
│   └── assets/           # Static assets
├── deployments/          # Docker and deployment configs
└── tests/                # Integration and unit tests
```

## Development Setup

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- PostgreSQL 16+
- Telegram Bot Token

### Local Development

1. Clone the repository:
```bash
git clone <repository-url>
cd StorageAccounting
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure your `.env` file with your Telegram bot token and other credentials.

4. Start services with Docker Compose:
```bash
cd deployments
docker-compose up -d
```

5. Access the application:
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### Running Without Docker

1. Start PostgreSQL locally.

2. Run database migrations:
```bash
# Migration commands to be added
```

3. Run the backend:
```bash
cd backend
go run cmd/server/main.go
```

4. Serve the frontend via nginx or any static file server.

## Configuration

Configure the application using environment variables (see `.env.example`):

- `PORT` - HTTP server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `TELEGRAM_BOT_TOKEN` - Your Telegram bot token
- `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` - S3-compatible storage config

## Deployment

### Docker Deployment

Build and run with Docker Compose:

```bash
cd deployments
docker-compose up -d
```

### Production Considerations

- Configure HTTPS with valid SSL certificates
- Update nginx.conf to enable HTTPS server block
- Set up automated database backups
- Configure proper S3 storage for production
- Enable Telegram webhook for bot integration

## API Documentation

### Endpoints

- `GET /health` - Health check
- `GET /api/items` - List inventory items
- `GET /api/items/:id` - Get item details
- `POST /api/items` - Create new item
- `PUT /api/items/:id/quantity` - Update item quantity
- `GET /api/items/:id/transactions` - Get transaction history

All API endpoints require Telegram WebApp authentication.

## Testing

Run tests:
```bash
cd backend
go test ./...
```

## License

Proprietary
