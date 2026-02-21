# Architecture - Services Gateway & Message

## Vue d'ensemble

Les services Gateway et Message ont été restructurés pour suivre une architecture scalable, modulaire et testable en Go.

## Structure Générale

```
services/
├── gateway/
│   ├── cmd/api/                 # Point d'entrée
│   │   └── main.go
│   ├── internal/                # Code interne (non exportable)
│   │   ├── config/              # Gestion configuration
│   │   ├── domain/              # Modèles métier
│   │   ├── handler/             # Handlers HTTP
│   │   ├── middleware/          # Middlewares
│   │   ├── router/              # Définition routes
│   │   └── service/             # Logique métier
│   ├── pkg/                     # Code réutilisable (exportable)
│   │   └── client/              # Client HTTP
│   ├── tests/
│   │   ├── unit/                # Tests unitaires
│   │   └── integration/         # Tests intégration
│   ├── Dockerfile
│   ├── Makefile
│   ├── go.mod
│   └── README.md
│
└── message/
    ├── cmd/api/                 # Point d'entrée
    │   └── main.go
    ├── internal/                # Code interne (non exportable)
    │   ├── config/              # Gestion configuration
    │   ├── domain/              # Modèles métier
    │   ├── handler/             # Handlers HTTP
    │   ├── middleware/          # Middlewares
    │   ├── repository/          # Couche données
    │   ├── router/              # Définition routes
    │   └── service/             # Logique métier
    ├── pkg/
    │   └── errors/              # Types d'erreurs
    ├── tests/
    │   ├── unit/                # Tests unitaires
    │   └── integration/         # Tests intégration
    ├── Dockerfile
    ├── Makefile
    ├── go.mod
    └── README.md
```

## Principes d'Architecture

### 1. **Séparation des Responsabilités**

- **Handler (HTTP Layer)**: Gère les requêtes/réponses HTTP, validation des entrées
- **Service (Business Layer)**: Contient la logique métier, interaction entre modules
- **Repository (Data Layer)**: Gère l'accès aux données (Cassandra, Redis, etc.)
- **Domain**: Modèles de données et constants métier

### 3. **Dépendance d'Injection** 

Chaque composant reçoit ses dépendances via constructeur:

```go
// Service reçoit le repository
service := service.NewMessageService(repo)

// Handler reçoit le service
handler := handler.NewMessageHandler(service)
```

Avantages:
- ✅ Testabilité (facile à mocker)
- ✅ Flexibilité (changer l'implémentation)
- ✅ Pas de couplage fort

### 4. **Configuration Centralisée**

Via `internal/config/config.go`:
- Variables d'environnement
- Timeouts et limites
- Adresses des services dépendants

### 5. **Testabilité**

Architecture favorisant les tests:
- Repository interface pour mocker les données
- Mocks HTTP client facilement
- Pas de dépendances globales
- Contexte passé explicitement

## Principaux Patterns

### Pattern: Repository

Le service Message utilise le pattern Repository pour l'accès aux données:

```go
// Interface pour flexibilité
type MessageRepository interface {
    Create(ctx context.Context, message *Message) error
    GetByID(ctx context.Context, id string) (*Message, error)
    // ...
}

// Implémentations possibles:
- InMemoryMessageRepository  // Dev/Test
- CassandraRepository        // Production
- RedisRepository           // Cache
```

### Pattern: Service Locator (léger)

Configuration et services initialisés dans `main.go`:

```go
// Configuration
cfg := config.Load()

// Repositories
repo := repository.NewInMemoryMessageRepository()

// Services
svc := service.NewMessageService(repo)

// Handlers
handler := handler.NewMessageHandler(svc)
```

### Pattern: Middleware

Middlewares pour les cross-cutting concerns:

```go
r.Use(middleware.LoggerMiddleware())
r.Use(middleware.AuthMiddleware())  // Futur
```

## Couches de l'Application

### Gateway Service

```
HTTP Request
    ↓
Router (route matching)
    ↓
HealthHandler (HTTP layer)
    ↓
HealthService (business layer)
    ↓
HTTPClient (external services)
    ↓
HTTP Response
```

### Message Service

```
HTTP Request
    ↓
Router (route matching)
    ↓
MessageHandler (HTTP layer)
    ↓
MessageService (business layer)
    ↓
MessageRepository (data layer)
    ↓
Database/Cache
    ↓
HTTP Response
```

## Configuration Management

### Exemple: Gateway

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Services ServicesConfig
}

// Usage dans main.go
cfg := config.Load()
r.Run(cfg.Server.GetAddr())
```

### Avantages
- ✅ Single source of truth
- ✅ Validation centralisée
- ✅ Facile à tester
- ✅ Pas de magic strings partout

## Error Handling

### Package errors

Message service exporte ses erreurs depuis `pkg/errors/`:

```go
const (
    ErrNotFound       = 404
    ErrInvalidInput   = 400
    ErrUnauthorized   = 401
)

// Utilisation
return NewAppError(ErrNotFound, "Message not found", details)
```

## Tests

### Approche

1. **Mocks via interfaces**: Les services dépendent d'interfaces
2. **Mocks simples**: Pas de framework de mock lourd
3. **Tests séparés**: `tests/unit/` et `tests/integration/`

### Exemple: Message Service

```go
// Mock pour repository
mockRepo := &MockMessageRepository{
    messages: make(map[string]*Message),
}

// Service avec mock
svc := service.NewMessageService(mockRepo)

// Test
msg, err := svc.CreateMessage(ctx, req)
```

## Scalabilité

### Horizontal Scaling

- **Stateless services**: Pas de session locale
- **Configuration externalisée**: Via variables d'env
- **Logs structurés**: Facilite la centralisation

### Vertical Scaling

- **Timeouts configurables**: Pas de blocage infini
- **Context cancellation**: Gère les timeouts
- **Connection pooling**: Via HTTP client

## Évolution Future

### Améliorations possibles

**Gateway:**
- [ ] Circuit breaker pour services down
- [ ] Request/Response logging complet
- [ ] Rate limiting
- [ ] Caching responses
- [ ] Agrégation d'erreurs

**Message:**
- [ ] Cassandra integration
- [ ] Redis caching
- [ ] NATS event publishing
- [ ] Pagination
- [ ] Full-text search
- [ ] Message archival

## Build & Deployment

### Local Development

```bash
cd services/gateway
make build
make run
make test
```

### Docker (existant)

```bash
docker build -t gateway:dev .
docker run -e PORT=8080 gateway:dev
```

### Kubernetes (existant)

K8s manifests in `k8s/gateway.yaml` et `k8s/message.yaml`

## Conventions de Code

### Naming

- Interfaces: `interface_name.go`
- Implementation: Dans même package
- Test files: `*_test.go`
- Structs: CamelCase sans préfixe I
- Functions: CamelCase, verbes en tête (GetMessage, CreateUser)

### Imports

```go
// Standard library
import (
    "context"
    "fmt"
    "net/http"
)

// External libraries
import (
    "github.com/gin-gonic/gin"
)

// Internal packages
import (
    "message-service/internal/domain"
)
```

### Error Handling

```go
// Préféré
msg, err := service.GetMessage(ctx, id)
if err != nil {
    return fmt.Errorf("failed to get message: %w", err)
}

// Éviter les panics en API
// Éviter silent failures
```

## Ressources

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
