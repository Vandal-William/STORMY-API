# Message Service - Documentation

Bienvenue dans le service de messaging de Stormy!

## 📚 Documentation

Ce service expose une API REST complète pour gérer les conversations et les messages.

### Pour Commencer Rapidement
Lire: **[ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md)** - Guide simple et rapide de l'architecture

### Documentation Complète
Lire: **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** - Référence exhaustive de toutes les routes

---

## 🚀 Quick Start

### Déploiement

Le service est actuellement déployé sur Kubernetes:

```bash
# Voir les pods
kubectl get pods -l app=message-service

# Voir les logs
kubectl logs -l app=message-service

# Redéployer (après rebuild)
kubectl rollout restart deployment/message-service
```

### Build Docker

```bash
cd services/message
docker build -t message-service:dev .

# Tester localement
docker run -p 3001:3001 message-service:dev
```

### Test Rapide

```bash
# Health check
curl http://localhost:3001/info

# Créer une conversation
curl -X POST http://localhost:3001/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test",
    "type": "group",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "member_ids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"]
  }'
```

---

## 📋 API Routes

### Conversations (8 endpoints)

```
POST   /conversations                          Créer
GET    /conversations/:id                      Récupérer
PUT    /conversations/:id                      Modifier
DELETE /conversations/:id                      Supprimer
GET    /users/:user_id/conversations           Lister (user)
GET    /conversations/:id/members              Lister membres
POST   /conversations/:id/members              Ajouter membre
DELETE /conversations/:id/members/:user_id    Retirer membre
```

### Messages (5 endpoints)

```
POST   /messages                     Créer
GET    /messages/:id                 Récupérer
PUT    /messages/:id                 Modifier
DELETE /messages/:id                 Supprimer
GET    /users/:user_id/messages      Lister (user)
```

---

## 🏗️ Architecture

Le service utilise le **Repository Pattern** avec 3 couches:

```
HTTP Requests
    ↓
Handler (handler/*.go) - HTTP plumbing
    ↓
Service (service/*.go) - Business logic
    ↓
Repository (repository/message.go) - Data access
    ↓
Storage (In-Memory ou Cassandra)
```

**Chaque couche a une responsabilité unique.**

Pour comprendre comment ça marche:
1. Commencez par [ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md)
2. Consultez [API_DOCUMENTATION.md](API_DOCUMENTATION.md) pour les détails

---

## 📁 Structure du Projet

```
services/message/
├── cmd/api/main.go                    # Point d'entrée
├── internal/
│   ├── handler/                       # HTTP Handlers
│   │   ├── conversation.go
│   │   ├── message.go
│   │   └── health.go
│   ├── service/                       # Business Logic
│   │   ├── conversation.go
│   │   └── message.go
│   ├── repository/                    # Data Access
│   │   ├── message.go (interfaces)
│   │   └── cassandra_*.go (implémentations)
│   ├── domain/models.go               # Structures métier
│   ├── config/config.go               # Configuration
│   ├── router/router.go               # Routes Gin
│   └── middleware/logger.go           # Middleware
├── k8s/message.yaml                   # Kubernetes manifests
├── db/cassandra/schema.cql            # Schéma Cassandra
├── Dockerfile                         # Build image
├── API_DOCUMENTATION.md               # Documentation complète
├── ARCHITECTURE_GUIDE.md              # Guide rapide
└── README.md                          # Ce fichier
```

---

## 🔧 Configuration

Variables d'environnement:

```bash
# Cassandra
CASSANDRA_HOSTS=cassandra:9042
CASSANDRA_KEYSPACE=message_service
CASSANDRA_CONSISTENCY=LOCAL_ONE

# Redis (cache)
REDIS_HOST=redis-message
REDIS_PORT=6379

# NATS (message queue)
NATS_URL=nats://nats:4222

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=3001
```

---

## 🧪 Tests

```bash
# Lancer les tests du service
go test ./internal/service/...

# Lancer les tests des handlers
go test ./internal/handler/...

# Lancer tous les tests avec coverage
go test -v -cover ./...
```

---

## 📖 En savoir plus

- **Architecture en détail**: [ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md)
- **Routes et exemples**: [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
- **Code du service**: `internal/`
- **Schéma DB**: `db/cassandra/schema.cql`
- **Kubernetes**: `k8s/message.yaml`

---

## 🚨 Troubleshooting

### Le service ne démarre pas

```bash
# Vérifier les logs
kubectl logs -l app=message-service

# Vérifier la configuration
kubectl get configmap -n default
```

### Les endpoints retournent 404

Vérifier que les routes sont bien définies dans `router/router.go`

### UUID format invalide

Les IDs doivent être des UUIDs valides:
```
550e8400-e29b-41d4-a716-446655440000  ✅ Bon
not-a-uuid  ❌ Mauvais
550e8400e29b41d4a716446655440000  ❌ Mauvais (pas de tirets)
```

---

## 📝 Ajouter une nouvelle route

Voir la checklist complète dans [ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md#-ajouter-une-nouvelle-route)

TL;DR:
1. Ajouter au Handler (HTTP)
2. Ajouter au Service (logique métier)
3. Ajouter à la Repository interface
4. Implémenter en In-Memory
5. Implémenter en Cassandra
6. Ajouter dans router.go
7. Tester

---

## 🔮 Roadmap

- [ ] Intégration complète Cassandra
- [ ] Authentification JWT
- [ ] Webhooks (NATS)
- [ ] Redis caching
- [ ] Full-text search
- [ ] Advanced pagination
- [ ] Integration tests

---

## 📞 Support

Questions? Consultez les docs:
1. [ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) - Concepts et patterns
2. [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Routes et exemples
3. Code source - `internal/`

Bonne programmation! 🚀
