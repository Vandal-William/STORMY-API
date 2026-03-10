# Guide d'intégration des services existants

Ce guide explique comment intégrer vos services existants avec la nouvelle gateway scalable.

## 📋 Checklist avant l'intégration

- ✅ La gateway est compilée et testée
- ✅ Votre service a une URL de base (ex: `http://service:port`)
- ✅ Votre service expose des endpoints HTTP
- ✅ Vous connaissez les endpoints de votre service

## 🔌 Étapes d'intégration

### Étape 1: Identifier les paramètres du service

Avant d'ajouter à la configuration, collectez:

```
Service Name:    [exemple: "messages"]
Service URL:     [exemple: "http://message-service:3001"]
Route Prefix:    [exemple: "/messages"]
Description:     [exemple: "Service de gestion des messages"]
Auth Required:   [true/false]
```

### Étape 2: Ajouter à `config/services.yaml`

```yaml
services:
  - name: messages           # ← Identifier unique
    url: http://message-service:3001    # ← URL complète
    prefix: /messages        # ← Préfixe pour accéder
    description: Service de gestion des messages
    auth_required: true      # ← JWT requiert?
```

### Étape 3: Tester la connexion

```bash
# Vérifier que la gateway reconnait le service
curl http://localhost:8080/info

# Tester une requête simple
curl http://localhost:8080/messages/123
```

## 📚 Intégration des services existants

### Service `users` (User Service)

Ajouter à `config/services.yaml`:

```yaml
  - name: users
    url: http://user-service:3000
    prefix: /users
    description: Gestion des utilisateurs
    auth_required: false
```

Routes créées:
```
GET    /users/:id     → http://user-service:3000/:id
POST   /users         → http://user-service:3000
PUT    /users/:id     → http://user-service:3000/:id
DELETE /users/:id     → http://user-service:3000/:id
```

### Service `messages` (Message Service)

Ajouter à `config/services.yaml`:

```yaml
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    description: Gestion des messages et conversations
    auth_required: true
```

Routes créées:
```
GET    /messages/:id                → http://message-service:3001/:id
POST   /messages                    → http://message-service:3001
PUT    /messages/:id                → http://message-service:3001/:id
DELETE /messages/:id                → http://message-service:3001/:id
GET    /messages/:id/members        → http://message-service:3001/:id/members
POST   /messages/:id/members        → http://message-service:3001/:id/members
```

### Service `presence` (Presence Service)

Ajouter à `config/services.yaml`:

```yaml
  - name: presence
    url: http://presence-service:3002
    prefix: /presence
    description: Détection de présence et statut
    auth_required: true
```

### Service `notification` (Notification Service)

Ajouter à `config/services.yaml`:

```yaml
  - name: notification
    url: http://notification-service:3003
    prefix: /notification
    description: Service de notifications
    auth_required: true
```

### Service `moderation` (Moderation Service)

Ajouter à `config/services.yaml`:

```yaml
  - name: moderation
    url: http://moderation-service:3004
    prefix: /moderation
    description: Modération du contenu
    auth_required: true
```

### Routes d'authentification (Auth)

Les routes d'authentification sont souvent sur un service existant:

```yaml
  - name: auth
    url: http://user-service:3000
    prefix: /auth
    description: Routes d'authentification
    auth_required: false
```

Routes disponibles:
```
POST   /auth/login     → http://user-service:3000/auth/login
POST   /auth/register  → http://user-service:3000/auth/register
POST   /auth/logout    → http://user-service:3000/auth/logout
```

## 🔐 Authentification et services

### Services publics (sans authentification)

Pour les services accessibles sans token JWT:

```yaml
  - name: blog
    url: http://blog-service:3010
    auth_required: false    # ← N'importe qui peut accéder
```

### Services protégés (avec authentification)

Pour les services nécessitant un JWT:

```yaml
  - name: admin
    url: http://admin-service:3020
    auth_required: true     # ← Token JWT requiert
```

Flux:
1. Client se connecte via `/auth/login`
2. Reçoit un JWT dans les cookies
3. Envoie les requêtes vers `/admin/*` avec le cookie
4. Gateway valide le JWT
5. Proxifie à`admin-service`

## 🧪 Tests d'intégration

### Test 1: Vérifier l'enregistrement

```bash
# Affiche tous les services enregistrés
curl http://localhost:8080/info
```

Réponse attendue:
```json
{
  "gateway": "healthy",
  "timestamp": 1709976000,
  "services": [
    {
      "name": "messages",
      "url": "http://message-service:3001",
      "prefix": "/messages",
      "auth_required": true
    },
    ...
  ]
}
```

### Test 2: Requête simple (GET)

```bash
curl http://localhost:8080/messages/123
```

Cela proxifie vers: `http://message-service:3001/123`

### Test 3: Requête avec données (POST)

```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"title":"Mon message","content":"Coucou"}'
```

Cela proxifie vers:
```
POST http://message-service:3001
Body: {"title":"Mon message","content":"Coucou"}
```

### Test 4: Requête avec authentification

```bash
# 1. Se connecter
curl -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# 2. Accéder à une ressource protégée
curl -b cookies.txt http://localhost:8080/messages/123
```

### Test 5: Tester les verbes HTTP

```bash
# GET
curl http://localhost:8080/messages/123

# POST
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"data":"test"}'

# PUT
curl -X PUT http://localhost:8080/messages/123 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated"}'

# DELETE
curl -X DELETE http://localhost:8080/messages/123

# PATCH
curl -X PATCH http://localhost:8080/messages/123 \
  -H "Content-Type: application/json" \
  -d '{"status":"archived"}'
```

## 🚀 Migration depuis une gateway monolithique

Si vous aviez une gateway avec des handlers spécifiques:

### Avant (ancien code)
```go
// handler/message.go
func (h *MessageHandler) CreateMessage(c *gin.Context) {
    // Code spécifique à ce service
}

// router/router.go
messageGroup := protectedAPI.Group("/messages")
messageGroup.POST("", messageHandler.CreateMessage)
```

### Après (nouveau code)
```yaml
# config/services.yaml
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    auth_required: true
```

**C'est tout!** Les handlers spécifiques ne sont plus nécessaires.

## 📝 Exemples de configurations complètes

### Configuration minimale (1 service)

```yaml
services:
  - name: api
    url: http://api:3000
    prefix: /api
    auth_required: false
```

### Configuration complète (5 services)

```yaml
services:
  # Publics
  - name: auth
    url: http://user-service:3000
    prefix: /auth
    description: Authentification
    auth_required: false

  - name: blog
    url: http://blog-service:3010
    prefix: /blog
    description: Blog public
    auth_required: false

  # Protégés
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    description: Messaging privé
    auth_required: true

  - name: profile
    url: http://user-service:3000
    prefix: /profile
    description: Profil utilisateur
    auth_required: true

  - name: admin
    url: http://admin-service:3020
    prefix: /admin
    description: Panneau d'administration
    auth_required: true
```

### Configuration pour microservices

```yaml
services:
  # Authentification & Utilisateurs
  - name: users
    url: http://user-service:3000
    prefix: /users
    auth_required: false

  - name: auth
    url: http://user-service:3000
    prefix: /auth
    auth_required: false

  # Messaging
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    auth_required: true

  - name: conversations
    url: http://message-service:3001
    prefix: /conversations
    auth_required: true

  # Présence & Notifications
  - name: presence
    url: http://presence-service:3002
    prefix: /presence
    auth_required: true

  - name: notifications
    url: http://notification-service:3003
    prefix: /notifications
    auth_required: true

  # Modération
  - name: moderation
    url: http://moderation-service:3004
    prefix: /moderation
    auth_required: true

  # Services additionnels
  - name: search
    url: http://search-service:3005
    prefix: /search
    auth_required: false

  - name: analytics
    url: http://analytics-service:3006
    prefix: /analytics
    auth_required: true
```

## 🔍 Debugging

### La requête retourne 502 Bad Gateway

```bash
# Vérifier que l'URL du service est correcte
curl http://message-service:3001/health

# Vérifier les logs de la gateway
docker logs gateway
```

### La requête retourne 401 Unauthorized

```bash
# Vérifier que:
# 1. auth_required: true dans la config
# 2. Vous envoyez le token JWT
curl -b cookies.txt http://localhost:8080/messages/123
```

### La requête retourne 404 Not Found

```bash
# Vérifier que le service et le préfixe existent
curl http://localhost:8080/info
```

## 📚 Ressources

- [Architecture.md](./ARCHITECTURE.md) - Explication détaillée
- [Quick Start](./QUICK_START.md) - Démarrage rapide
- [RFC 7231 HTTP Methods](https://tools.ietf.org/html/rfc7231)

## ✅ Checklist de vérification

- [ ] Service identifié et paramètres collectés
- [ ] Configuration ajoutée à `config/services.yaml`
- [ ] Gateway redémarrée
- [ ] Service visible dans `GET /info`
- [ ] Tests de connexion réussis
- [ ] Authentification correcte (pour services protégés)
- [ ] Headers/cookies proxifiés correctement
- [ ] Body et query params proxifiés correctement
