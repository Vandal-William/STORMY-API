# Gateway Service

API Gateway that provides a single entry point and health monitoring for all downstream services.

## Architecture

```
cmd/api/           - Application entry point
internal/
  ├── config/      - Configuration management
  ├── domain/      - Domain models
  ├── handler/     - HTTP request handlers
  ├── middleware/  - HTTP middleware
  ├── service/     - Business logic
  └── router/      - Route definitions
pkg/
  └── client/      - HTTP client for service calls
tests/
  ├── unit/        - Unit tests
  └── integration/ - Integration tests
```

## Configuration

Set the following environment variables:

```bash
PORT=8080
HOST=0.0.0.0
USER_SERVICE_URL=http://user-service:3000
MESSAGE_SERVICE_URL=http://message-service:3001
PRESENCE_SERVICE_URL=http://presence-service:3002
NOTIFICATION_SERVICE_URL=http://notification-service:3003
MODERATION_SERVICE_URL=http://moderation-service:3004
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

Returns the health status of the gateway and all downstream services.

**Response:**
```json
{
  "gateway": "ok",
  "message": "200 OK",
  "presence": "200 OK",
  "user": "200 OK",
  "notification": "200 OK",
  "moderation": "200 OK"
}
```

## Scalability Features

- **Configuration management** via environment variables
- **Dependency injection** for easy testing and modularity
- **Middleware support** for cross-cutting concerns
- **Abstract service layer** for business logic
- **HTTP client wrapper** for service communication
- **Unit tests** with mock implementations
- **Clear separation of concerns** (handler, service, domain, client)

## Testing

The service is designed to be testable:

- Services depend on interfaces, not concrete implementations
- Handlers depend on service interfaces
- Mock clients can be easily created for testing

```bash
# Run unit tests
make test

# Run with verbose output
make test-v
```
