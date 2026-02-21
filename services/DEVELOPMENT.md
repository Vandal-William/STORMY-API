# Development Guidelines

## Quick Start

### Gateway Service

```bash
cd services/gateway

# Installer dépendances
go mod download

# Compiler
make build

# Lancer
make run

# Tests
make test

# Lancer en dev mode (hot reload avec air)
make dev
```

### Message Service

```bash
cd services/message

# Installer dépendances
go mod download

# Compiler
make build

# Lancer
make run

# Tests
make test
```

## Structure des Fichiers

### Règles

1. **Un fichier par concept**: `health.go`, `message.go`, non `handlers.go`
2. **Package organization**:
   - Un package = un concept
   - Nom du fichier décrit le contenu
   - Pas de fichier `utils.go` vague

3. **Internal vs Pkg**:
   - `internal/`: Code privé au service (non importable)
   - `pkg/`: Code réutilisable (exportable)

### Exemple correcte Structure

```
internal/
├── config/
│   └── config.go          # Juste la config
├── domain/
│   └── models.go          # Modèles métier
├── handler/
│   ├── health.go          # Health check
│   └── message.go         # Message endpoints
├── service/
│   └── message.go         # Logique métier
└── router/
    └── router.go          # Setup routes
```

## Code Style

### Exemple Gateway Handler

```go
package handler

import (
    "gateway-service/internal/service"
    "net/http"
    "github.com/gin-gonic/gin"
)

// HealthHandler handles health-related requests
type HealthHandler struct {
    healthService *service.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
    return &HealthHandler{
        healthService: healthService,
    }
}

// GetInfo returns gateway health info
func (h *HealthHandler) GetInfo(c *gin.Context) {
    info := h.healthService.GetGatewayInfo()
    c.JSON(http.StatusOK, info)
}
```

### Points clés

✅ Commentaires pour exports publics
✅ Receiver court et cohérent (`h`, `s`)
✅ Constructeurs avec préfixe `New`
✅ Dépendances explicites via paramètres
✅ Pas de package init() global
✅ Erreurs gérées explicitement

## Dépendances Injection

### Anti-pattern

```go
// ❌ Mauvais: global state
var repo MessageRepository = initRepo()

func GetMessage(id string) {
    return repo.Get(id)  // Difficile à tester
}
```

### Pattern

```go
// ✅ Bon: dépendances explicites
type Service struct {
    repo MessageRepository
}

func NewService(repo MessageRepository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetMessage(id string) {
    return s.repo.Get(id)  // Facile à tester et à modifier
}
```

## Testing

### Structure Test

```go
package service

import "testing"

func TestCreateMessage(t *testing.T) {
    // Arrange: Setup
    repo := NewMockRepository()
    svc := NewMessageService(repo)
    
    // Act: Execute
    msg, err := svc.CreateMessage(ctx, req)
    
    // Assert: Verify
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
    if msg.ID == "" {
        t.Error("expected ID, got empty")
    }
}
```

### Nommer les Tests

```go
// ✅ Bon: TestFunctionBehavior
TestCreateMessageWithValidPayload(t *testing.T)
TestCreateMessageWithEmptyContent(t *testing.T)
TestGetMessageNotFound(t *testing.T)

// ❌ Mauvais
Test1(t *testing.T)
TestStuff(t *testing.T)
```

## Configuration

### Loading Config

```go
// internal/config/config.go
func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port: getEnv("PORT", "8080"),
        },
    }
}

func getEnv(key, defaultVal string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultVal
}
```

### Variables d'Env

- Nommer en UPPER_SNAKE_CASE
- Chaque service a le plus petit jeu possible
- Préfixer par SERVICE_NAME_ pour multi-service configs
- Documenter dans `.env.example`

## Error Handling

### Pattern pour APIs

```go
// ✅ Bon: Logger + Return appropriate status
func (h *MessageHandler) GetMessage(c *gin.Context) {
    msg, err := h.service.GetMessage(ctx, id)
    if err != nil {
        log.Errorf("failed to get message: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Failed to retrieve message",
        })
        return
    }
    c.JSON(http.StatusOK, msg)
}
```

### Pattern pour Services

```go
// ✅ Bon: Retourner erreur explicite
func (s *Service) GetMessage(ctx context.Context, id string) (*Message, error) {
    msg, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("repository error: %w", err)
    }
    if msg == nil {
        return nil, fmt.Errorf("message not found")
    }
    return msg, nil
}
```

## Logging

### Utilisation

```go
import "log"

// ✅ Simple et efficace
log.Printf("Starting service on %s\n", addr)
log.Printf("[ERROR] Failed to connect: %v\n", err)

// Futur: structured logging
// logger.WithFields(log.Fields{
//     "service": "message",
//     "user_id": userID,
// }).Infof("Creating message")
```

## Git Workflow

### Avant de commit

```bash
# Tests
make test

# Linter (si configuré)
make lint

# Build
make build
```

### Commit Message

```
feat(gateway): Add circuit breaker for service health checks
fix(message): Handle nil messages in update endpoint
docs(architecture): Update scalability guidelines
refactor(service): Extract validation to separate function
test(message): Add edge cases for message creation
```

## Useful Commands

```bash
# Compile
go build ./cmd/api/main.go

# Format
go fmt ./...

# Lint (si golangci-lint installé)
golangci-lint run ./...

# Vendoring
go mod tidy
go mod vendor

# Debug avec dlv (si installé)
dlv debug ./cmd/api

# Benchmark (préfixe Benchmark)
go test -bench=. -benchmem

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance Tips

### Timeouts

```go
// ✅ Toujours définir les timeouts
client := &http.Client{
    Timeout: 5 * time.Second,
}

// Avec context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### Buffers

```go
// Pre-allocate si size connu
items := make([]Message, 0, 100)  // Capacity 100

// ❌ Éviter
items := make([]Message, 0)  // Reallocation à chaque append
```

### Goroutines

```go
// ✅ Limiter avec worker pool pattern
// ❌ Éviter: goroutine par request
for msg := range messages {
    go processMessage(msg)  // Chaotique!
}

// ✅ Bon: Canal avec workers
workers := 10
for i := 0; i < workers; i++ {
    go worker(messages)
}
```

## Debugging

### Logs bien structurés

```go
log.Printf("[%s] %d | %s %s | %v\n", 
    time.Now().Format("15:04:05"),
    statusCode,
    method,
    path,
    latency)
```

### Breakpoints dans IDE

- Placer breakpoint dans code
- Run debug configuration
- VS Code: Installation de l'extension Go automatique

## Checklist Before PR

- [ ] Tests passent: `make test`
- [ ] Code compile: `make build`
- [ ] Lint OK: `make lint` (si dispo)
- [ ] Nouveaux endpoints documentés dans README
- [ ] Configuration exemple mise à jour: `.env.example`
- [ ] Imports triés et nettoyés
- [ ] Pas de TODO/FIXME sans issue associée
- [ ] Commit messages explicites
