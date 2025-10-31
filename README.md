# Kanban Simple

A lightweight kanban board application with a REST API for bot integration and automation. Built with Go, SQLite, and vanilla JavaScript.

## Features

- **Drag-and-Drop Interface**: Intuitive card management
- **REST API**: Complete API for automation and bot integration
- **SQLite Database**: Embedded database with zero configuration
- **Archive System**: Archive completed cards with easy restoration
- **Search & Filter**: Search across boards and cards
- **Comments**: Track progress with card comments
- **Labels**: Organize cards with colored labels
- **Lightweight**: Docker image < 15MB (scratch-based)
- **No Authentication**: Simple, open board (authentication can be added via reverse proxy)

## Quick Start

### Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/kanban-simple.git
cd kanban-simple

# Run directly (Go 1.22+ required)
go run cmd/server/main.go

# Access at http://localhost:8080
```

### Building

```bash
# Build binary
go build -o server cmd/server/main.go

# Run
./server
```

### Docker

```bash
# Build image
docker build -t kanban .

# Run container
docker run -p 8080:8080 kanban

# Or use docker-compose
docker-compose up -d
```

## Configuration

Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_PATH` | `./data/kanban.db` | SQLite database file path |
| `MIGRATIONS_PATH` | `./migrations` | Database migration files |
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | Gin mode (debug/release) |

## API Documentation

### OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:
```
http://localhost:8080/openapi.yaml
```

You can use this spec with API clients like:
- **Postman**: Import OpenAPI spec
- **Bruno**: Import OpenAPI spec
- **Insomnia**: Import OpenAPI spec
- **Swagger Editor**: View at https://editor.swagger.io/ (paste the spec)
- **Redoc**: Generate documentation with `npx @redocly/cli preview-docs openapi.yaml`

### API Endpoints

**Base URL**: `http://localhost:8080/api`

#### Boards
- `GET /api/boards` - List all boards
- `POST /api/boards` - Create board
- `GET /api/boards/{id}` - Get board
- `PUT /api/boards/{id}` - Update board
- `DELETE /api/boards/{id}` - Delete board
- `GET /api/boards/{id}/lists` - Get board lists

#### Lists (Columns)
- `POST /api/boards/{board_id}/lists` - Create list
- `GET /api/lists/{id}` - Get list
- `PUT /api/lists/{id}` - Update list
- `PATCH /api/lists/{id}/move` - Move list (reorder)
- `DELETE /api/lists/{id}` - Delete list
- `GET /api/lists/{id}/cards` - Get list cards

#### Cards (Tasks)
- `POST /api/lists/{list_id}/cards` - Create card
- `POST /api/cards/quick` - Quick create (minimal fields)
- `GET /api/cards/{id}` - Get card
- `PUT /api/cards/{id}` - Update card
- `PATCH /api/cards/{id}/move` - Move card (list/position)
- `POST /api/cards/{id}/archive` - Archive card
- `POST /api/cards/{id}/unarchive` - Unarchive card
- `DELETE /api/cards/{id}` - Delete card
- `GET /api/cards?query=...` - Search cards

#### Comments
- `GET /api/cards/{id}/comments` - Get card comments
- `POST /api/cards/{id}/comments` - Add comment

#### Labels
- `GET /api/labels` - List all labels
- `POST /api/labels` - Create label
- `GET /api/labels/{id}` - Get label
- `PUT /api/labels/{id}` - Update label
- `DELETE /api/labels/{id}` - Delete label
- `POST /api/cards/{id}/labels/{label_id}` - Assign label to card
- `DELETE /api/cards/{id}/labels/{label_id}` - Remove label from card
- `GET /api/cards/{id}/labels` - Get card labels

### Example API Usage

**Create a card**:
```bash
curl -X POST http://localhost:8080/api/lists/1/cards \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Fix bug in authentication",
    "description": "Users unable to log in",
    "color": "#ef4444",
    "due_date": "2024-12-31T23:59:59Z"
  }'
```

**Move a card**:
```bash
curl -X PATCH http://localhost:8080/api/cards/1/move \
  -H "Content-Type: application/json" \
  -d '{
    "list_id": 2,
    "position": 1.5
  }'
```

**Add a comment**:
```bash
curl -X POST http://localhost:8080/api/cards/1/comments \
  -H "Content-Type: application/json" \
  -d '{"content": "Working on this now"}'
```

**Search cards**:
```bash
curl "http://localhost:8080/api/cards?query=bug&board_id=1&archived=false"
```

## Database Schema

The application uses SQLite with the following tables:

**boards**
- `id` (INTEGER PRIMARY KEY)
- `name` (TEXT)
- `description` (TEXT)
- `created_at`, `updated_at` (DATETIME)

**lists**
- `id` (INTEGER PRIMARY KEY)
- `board_id` (INTEGER, FK → boards)
- `name` (TEXT)
- `color` (TEXT)
- `position` (REAL) - for ordering
- `created_at`, `updated_at` (DATETIME)

**cards**
- `id` (INTEGER PRIMARY KEY)
- `list_id` (INTEGER, FK → lists)
- `title` (TEXT)
- `description` (TEXT)
- `color` (TEXT)
- `position` (REAL) - for ordering
- `archived` (BOOLEAN)
- `due_date` (DATETIME)
- `created_at`, `updated_at` (DATETIME)

**comments**
- `id` (INTEGER PRIMARY KEY)
- `card_id` (INTEGER, FK → cards)
- `content` (TEXT)
- `created_at` (DATETIME)

**labels**
- `id` (INTEGER PRIMARY KEY)
- `name` (TEXT)
- `color` (TEXT)
- `created_at` (DATETIME)

**card_labels** (many-to-many)
- `card_id` (INTEGER, FK → cards)
- `label_id` (INTEGER, FK → labels)

### Database Features
- **WAL Mode**: Write-Ahead Logging for better concurrency
- **Foreign Keys**: Enforced with CASCADE deletes
- **Indexes**: Optimized queries on foreign keys
- **Migrations**: Automatic schema setup on first run

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── middleware/          # Middleware (error handling)
│   │   └── router.go            # Route definitions
│   ├── database/
│   │   └── db.go                # Database connection
│   ├── models/                  # Data models
│   └── repository/              # Database queries
├── migrations/                  # SQL migration files
├── web/
│   └── static/                  # Frontend (HTML/CSS/JS)
├── openapi.yaml                 # OpenAPI specification
├── Dockerfile                   # Scratch-based image
├── docker-compose.yml           # Docker Compose config
└── README.md                    # This file
```

## Development

### Database Migrations

Migrations run automatically on startup. To add new migrations:

1. Create file in `migrations/` with format `00X_description.sql`
2. Migrations run in alphabetical order
3. Use transactions for safety

### Testing the API

Use the OpenAPI spec with your favorite API client, or test with curl:

```bash
# Health check
curl http://localhost:8080/api/health

# List boards
curl http://localhost:8080/api/boards

# Get OpenAPI spec
curl http://localhost:8080/openapi.yaml
```

## License

MIT License - see LICENSE file for details
