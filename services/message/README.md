# Message Service

Service responsible for handling message operations with Cassandra storage and Redis caching.

## Architecture

```
cmd/api/           - Application entry point
internal/
  ├── config/      - Configuration management
  ├── domain/      - Domain models
  ├── handler/     - HTTP request handlers
  ├── middleware/  - HTTP middleware
  ├── repository/  - Data access layer
  ├── service/     - Business logic
  └── router/      - Route definitions
pkg/
  └── errors/      - Custom error types
tests/
  ├── unit/        - Unit tests
  └── integration/ - Integration tests
```

## Configuration

Set the following environment variables:

```bash
PORT=3001
HOST=0.0.0.0
CASSANDRA_HOST=cassandra
CASSANDRA_PORT=9042
REDIS_HOST=redis-message
REDIS_PORT=6379
NATS_URL=nats://nats:4222
```

See `.env.example` for more details.

## Development

```bash
# Build the project
make build

# Run locally
make run

# Run tests
make test

# Run linter
make lint
```

## API Endpoints

### Health Check

```
GET /info
```

Returns the health status of the service.

**Response:**
```json
{
  "service": "message",
  "status": "ok"
}
```

### Messages

#### Create Message

```
POST /messages
```

**Request:**
```json
{
  "sender_id": "user1",
  "receiver_id": "user2",
  "content": "Hello, World!"
}
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "sender_id": "user1",
    "receiver_id": "user2",
    "content": "Hello, World!",
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  },
  "message": "Message created successfully"
}
```

#### Get Message

```
GET /messages/:id
```

#### Get User Messages

```
GET /messages/user/:user_id
```

#### Update Message

```
PUT /messages/:id
```

**Request:**
```json
{
  "content": "Updated content"
}
```

#### Delete Message

```
DELETE /messages/:id
```

## Scalability Features

- **Repository pattern** for flexible data storage (in-memory, Cassandra, etc.)
- **Service layer** for business logic isolation
- **Dependency injection** for testability
- **Configuration management** via environment variables
- **Middleware support** for cross-cutting concerns
- **Unit tests** with mock implementations
- **CQRS-ready** architecture for future optimization
- **Event-driven** ready (NATS integration prepared)
- **Clear separation of concerns** (handler, service, repository, domain)

## Testing

```bash
# Run unit tests
make test

# Run with verbose output
make test-v
```

## Future Enhancements

- [ ] Cassandra integration for persistence
- [ ] Redis caching layer
- [ ] NATS event publishing
- [ ] Message pagination
- [ ] Message search/filtering
- [ ] Message archival
- [ ] Rate limiting
- [ ] Authentication/Authorization
