# Quick Start - Gateway Scalable

## Démarrage rapide

### 1. Configuration

Créez ou éditez un fichier `.env`:

```bash
PORT=8080
HOST=0.0.0.0
JWT_SECRET=votre-clé-secrète-super-forte
```

### 2. Configurer vos services

Éditez `config/services.yaml`:

```yaml
services:
  # Service d'authentification (exmple)
  - name: auth
    url: http://localhost:3000
    prefix: /auth
    description: Service d'authentification
    auth_required: false

  # Ajoutez vos services ici
```

### 3. Compiler et lancer

```bash
# Installer les dépendances
go mod download

# Compiler
make build

# Lancer la gateway
make run
```

La gateway démarre sur `http://localhost:8080`

## Ajouter une nouvelle API

### Option A: Via `config/services.yaml` (RECOMMANDÉ)

```yaml
services:
  - name: ma-api
    url: http://ma-api:port
    prefix: /ma-api
    description: Ma nouvelle API
    auth_required: true  # ou false
```

**Redémarrez la gateway** et c'est prêt!

### Routes créées automatiquement:
- `GET /ma-api/*` 
- `POST /ma-api/*`
- `PUT /ma-api/*`
- `DELETE /ma-api/*`
- `PATCH /ma-api/*`

### Option B: Via variables d'environnement (Fallback)

```bash
USER_SERVICE_URL=http://user-service:3000
MESSAGE_SERVICE_URL=http://message-service:3001
PRESENCE_SERVICE_URL=http://presence-service:3002
NOTIFICATION_SERVICE_URL=http://notification-service:3003
MODERATION_SERVICE_URL=http://moderation-service:3004
```

## Tester

### 1. Health Check

```bash
curl http://localhost:8080/info
```

Réponse:
```json
{
  "gateway": "healthy",
  "timestamp": 1709976000,
  "services": [
    {
      "name": "auth",
      "url": "http://localhost:3000",
      "prefix": "/auth",
      "auth_required": false
    }
  ]
}
```

### 2. Requête authentifiée

```bash
# 1. Login pour obtenir un token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass"}' \
  -c cookies.txt

# 2. Utiliser le cookie dans une requête protégée
curl http://localhost:8080/protected-api/resource \
  -b cookies.txt
```

### 3. Requête non authentifiée

```bash
curl http://localhost:8080/auth/register \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"new@example.com","password":"secure"}'
```

## Structure des requêtes proxifiées

### Avant (recopié)
```
Client → Gateway → Service
- Headers: ✅ copiés
- Cookies: ✅ copiés
- Body: ✅ copié
- Query params: ✅ copiés
- Method: ✅ préservée
```

### Réponse
```
Service → Gateway → Client
- Status code: ✅ préservé
- Headers: ✅ copiés
- Cookies: ✅ copiés (Set-Cookie)
- Body: ✅ copié intégralement
```

## Exemples d'intégration

### Ajouter un service Express.js

```yaml
  - name: nodejs-api
    url: http://nodejs-app:3005
    prefix: /node
    description: API Node.js
    auth_required: false
```

Testez:
```bash
curl http://localhost:8080/node/health
```

### Ajouter un service Python/FastAPI

```yaml
  - name: python-api
    url: http://python-app:8000
    prefix: /python
    description: API Python FastAPI
    auth_required: true
```

Testez (avec token JWT):
```bash
curl -b cookies.txt http://localhost:8080/python/api/data
```

### Ajouter un service Java/Spring

```yaml
  - name: java-api
    url: http://java-app:8080
    prefix: /java
    description: API Java Spring Boot
    auth_required: true
```

### Ajouter un service Go

```yaml
  - name: go-api
    url: http://go-microservice:9090
    prefix: /go
    description: Microservice Go
    auth_required: false
```

## Authentification JWT

### Flux de sécurité

```
1. Client demande /auth/login
   ↓ (pas d'auth requise)
2. Gateway proxifie vers le service auth
3. Service retourne JWT dans Set-Cookie
4. Client accède à une route protégée
   ↓ avec le cookie contenant le token
5. Gateway valide le JWT via le middleware
6. Si valide: proxifie la requête
7. Si invalide: retourne 401 Unauthorized
```

### Configuration par service

```yaml
# Service public (pas d'authentification)
  - name: blog
    auth_required: false

# Service privé (authentification requise)
  - name: admin
    auth_required: true
```

## Troubleshooting

### La gateway ne démarre pas

```bash
# Vérifier les logs
go run ./cmd/api/main.go
```

Causes possibles:
- ❌ Port déjà utilisé → Changez `PORT`
- ❌ Fichier `config/services.yaml` invalide → Vérifiez la syntaxe YAML
- ❌ Service introuvable → Vérifiez les URLs dans la config

### Une requête retourne 502 Bad Gateway

Cela signifie que le service cible est injoignable:
- Vérifiez que l'URL est correcte dans `config/services.yaml`
- Vérifiez que le service est bien lancé
- Vérifiez les logs du service cible

```bash
# Test direct au service
curl http://service-url:port/path
```

### Une requête retourne 401 Unauthorized

Cela signifie que le JWT est invalide ou manquant:
- Vérifiez que vous incluez le cookie avec `-b cookies.txt`
- Vérifiez que `auth_required: true` dans la config
- Vérifiez que le token n'a pas expiré

## Fichiers clés

```
gateway/
├── config/
│   └── services.yaml      ← Ajouter vos services ICI
├── cmd/api/
│   └── main.go           ← Point d'entrée
├── internal/
│   ├── config/           ← Chargement de la config
│   ├── proxy/            ← Proxy universel
│   ├── registry/         ← Registre des services
│   └── router/           ← Routes dynamiques
└── ARCHITECTURE.md       ← Documentation complète
```

## Support des verbes HTTP

Tous ces verbes sont automatiquement proxifiés:

✅ GET     - Récupérer une ressource
✅ POST    - Créer une ressource
✅ PUT     - Remplacer une ressource
✅ DELETE  - Supprimer une ressource
✅ PATCH   - Modifier partiellement
✅ HEAD    - Comme GET sans body
✅ OPTIONS - Infos CORS

## Prochaines étapes

Voir [ARCHITECTURE.md](./ARCHITECTURE.md) pour:
- Explication détaillée de l'architecture
- Composants avancés
- Patterns de sécurité
- Optimisations de performance
