# Architecture de la Gateway Scalable

## Vue d'ensemble

La gateway a été complètement refactorisée pour devenir un **proxy universel et modulaire** capable de:
- ✅ Supporter **40+ APIs** différentes
- ✅ Ajouter une nouvelle API en **3 lignes de configuration**
- ✅ Copier intelligemment tous les headers, cookies, body et query parameters
- ✅ Gérer l'authentification JWT de manière centralisée
- ✅ Router les requêtes vers les services cibles de façon transparente

## Architecture modulaire

```
gateway/
├── cmd/api/
│   └── main.go                  # Point d'entrée simplifié
├── internal/
│   ├── config/
│   │   └── config.go           # Charge depuis YAML + env
│   ├── proxy/
│   │   └── universal.go        # Proxy universel intelligент
│   ├── registry/
│   │   └── service_registry.go # Registre dynamique des services
│   ├── router/
│   │   └── router.go           # Routes dynamiques générées
│   ├── handler/                # DÉPRÉCIÉ - handlers spécifiques
│   ├── service/                # DÉPRÉCIÉ - services spécifiques
│   └── middleware/             # Middleware partagés
├── config/
│   └── services.yaml           # Configuration des 40+ APIs
└── ...
```

## Flux de traitement d'une requête

```
Requête HTTP
    ↓
[Logging Middleware]
    ↓
[Route Matching]
    ├─ /info → Health Check
    └─ /{service}/* → [Auth Middleware?] → UniversalProxy
    ↓
[UniversalProxy]
    ├─ Copie headers
    ├─ Copie cookies
    ├─ Copie body
    ├─ Copie query params
    └─ Forward vers Service
    ↓
Service Cible
    ↓
Réponse
    ├─ Code status
    ├─ Headers
    ├─ Cookies
    └─ Body
    ↓
Client
```

## Configuration des services

### Fichier `config/services.yaml`

Chaque service est défini avec ces propriétés:

```yaml
services:
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    description: Service de gestion des messages
    auth_required: true
```

**Propriétés:**
- `name`: Identifiant unique du service
- `url`: URL de base du service (ex: `http://message-service:3001`)
- `prefix`: Préfixe de route pour accéder au service (ex: `/messages`)
- `description`: Description du rôle du service (optionnel)
- `auth_required`: Si `true`, nécessite un JWT valide

### Exemple complet avec plusieurs services

```yaml
services:
  # Service 1: Messages
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    description: Gestion des messages et conversations
    auth_required: true

  # Service 2: Utilisateurs
  - name: users
    url: http://user-service:3000
    prefix: /users
    description: Gestion des utilisateurs
    auth_required: false

  # Service 3: Authentification
  - name: auth
    url: http://user-service:3000
    prefix: /auth
    description: Authentification (login/register)
    auth_required: false

  # Service 4: Présence
  - name: presence
    url: http://presence-service:3002
    prefix: /presence
    description: Détection de présence
    auth_required: true

  # ... 36 autres services ...
```

## Comment ajouter une nouvelle API en 3 lignes

### Étape 1: Ajouter la configuration (1 ligne)

Éditer `config/services.yaml` et ajouter:

```yaml
  - name: mon-api
    url: http://mon-api:3000
    prefix: /mon-api
    description: Ma nouvelle API
    auth_required: false  # ou true
```

### Étape 2: Aucun autre changement nécessaire!

C'est tout! La gateway:
1. ✅ Charge automatiquement la configuration
2. ✅ Enregistre le service dans la registry
3. ✅ Crée les routes dynamiques
4. ✅ Configure l'authentification si nécessaire

### Exemple: Ajouter une API de recherche

```yaml
# Avant
services:
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    auth_required: true

# Après (ajoutez ces 4 lignes)
  - name: search
    url: http://search-service:3005
    prefix: /search
    description: Service de recherche avec elasticsearch
    auth_required: true
  # C'est tout!
```

Dès le redémarrage, toutes ces routes fonctionneront automatiquement:
- `GET /search/*` → proxifiée vers `http://search-service:3005/*`
- `POST /search/*` → proxifiée vers `http://search-service:3005/*`
- `PUT /search/*` → proxifiée vers `http://search-service:3005/*`
- `DELETE /search/*` → proxifiée vers `http://search-service:3005/*`

## Composants clés

### 1. UniversalProxyHandler (`internal/proxy/universal.go`)

**Responsabilité**: Proxifier intelligemment les requêtes HTTP

**Fonctionnalités**:
- Copie **TOUS** les headers HTTP (sauf hop-by-hop)
- Copie **TOUS** les cookies
- Copie le **body** entièrement
- Copie les **query parameters**
- Supporte tous les verbes HTTP (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)

**Utilisation**:
```go
proxyHandler := proxy.NewUniversalProxyHandler(httpClient)
proxyHandler.ProxyRequest(c, "http://target-service:3000/path")
```

### 2. ServiceRegistry (`internal/registry/service_registry.go`)

**Responsabilité**: Maintenir un registre à jour de tous les services

**Opérations**:
- `Register(service)` - Enregistrer un service
- `FindByPrefix(prefix)` - Trouver un service par son préfixe
- `FindByName(name)` - Trouver un service par son nom
- `GetAll()` - Lister tous les services
- `Unregister(prefix)` - Supprimer un service

**Thread-safe**: Utilise des verrous RWMutex pour accès concurrent

**Exemple**:
```go
registry := registry.NewServiceRegistry()
registry.Register(&Service{
    Name: "messages",
    URL: "http://message-service:3001",
    Prefix: "/messages",
    AuthRequired: true,
})

if service, found := registry.FindByPrefix("/messages"); found {
    fmt.Println("Service trouvé:", service.URL)
}
```

### 3. Router dynamique (`internal/router/router.go`)

**Responsabilité**: Créer dynamiquement toutes les routes

**Processus**:
1. Charge la configuration via `config.Load()`
2. Crée un proxy universel
3. Pour chaque service:
   - Crée un groupe de routes avec le préfixe du service
   - Ajoute le middleware JWT si `auth_required=true`
   - Enregistre les handlers pour tous les verbes HTTP
   - Connecte chaque route au proxy universel

**Routes spéciales**:
- `GET /` → Redirection vers `/swagger/index.html`
- `GET /info` → Health check et liste des services
- `GET /swagger/*` → Documentation Swagger (si présente)

### 4. Configuration (`internal/config/config.go`)

**Responsabilité**: Charger et valider la configuration

**Ordre de priorité**:
1. **Primaire**: Fichier `config/services.yaml`
2. **Fallback**: Variables d'environnement
3. **Défaut**: Services vides

**Support**:
- Chargement YAML avec validation
- Variable d'environnement pour paramètres du serveur
- Registry automatique des services

## Variables d'environnement

```bash
# Serveur HTTP
PORT=8080              # Port d'écoute (défaut: 8080)
HOST=0.0.0.0          # Interface d'écoute (défaut: 0.0.0.0)

# JWT
JWT_SECRET=ma-clé-secrète  # Clé pour valider les tokens JWT
```

## Avantages de cette architecture

### 1. Scalabilité

- ✅ Ajouter une 40ème API en 1 minute
- ✅ Éviter la duplication de code pour chaque service
- ✅ Configuration centralisée et lisible

### 2. Modularité

- ✅ Chaque composant a une responsabilité unique
- ✅ Proxy universel réutilisable
- ✅ Registry thread-safe et découplée

### 3. Maintenabilité

- ✅ Code bien documenté avec commentaires détaillés
- ✅ Pas de handlers spécifiques à maintenir
- ✅ Configuration YAML simple et explicite

### 4. Robustesse

- ✅ Gestion intelligente des headers HTTP
- ✅ Support complet des cookies
- ✅ Copie intégrale du body et des query params
- ✅ Gestion des erreurs explicite

### 5. Performance

- ✅ Proxy universel optimisé
- ✅ Registry avec cache
- ✅ Pas de reflection ou d'introspection coûteuse
- ✅ Client HTTP avec timeout configurable

## Exemples d'utilisation

### Ajouter une API de gestion d'événements

```yaml
  - name: events
    url: http://events-service:3007
    prefix: /events
    description: Gestion des événements et calendrier
    auth_required: true
```

Routes créées automatiquement:
- `POST /events` → `http://events-service:3007/`
- `GET /events/123` → `http://events-service:3007/123`
- `PUT /events/123` → `http://events-service:3007/123`
- `DELETE /events/123` → `http://events-service:3007/123`

### Ajouter une API publique sans authentification

```yaml
  - name: blog
    url: http://blog-service:3010
    prefix: /blog
    description: Service de blog public
    auth_required: false  # ← Pas d'authentification requise
```

Tous les clients peuvent accéder à `/blog/*` sans token JWT.

### Ajouter une API protégée

```yaml
  - name: admin
    url: http://admin-service:3020
    prefix: /admin
    description: Panneau d'administration (Admin seulement)
    auth_required: true  # ← Authentification requise
```

Seuls les clients avec un JWT valide peuvent accéder à `/admin/*`.

## Debugging et maintenance

### Vérifier les services enregistrés

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
      "name": "messages",
      "url": "http://message-service:3001",
      "prefix": "/messages",
      "description": "Service de gestion des messages",
      "auth_required": true
    },
    ...
  ]
}
```

### Tester une requête proxifiée

```bash
# Requête authentifiée
curl -H "Cookie: ACCESS_TOKEN=VOTRE_TOKEN" \
  http://localhost:8080/messages/123

# Requête non authentifiée (si auth_required=false)
curl http://localhost:8080/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass"}'
```

## Migration depuis l'ancienne architecture

### Ancien code (déprecié)
```go
messageHandler := handler.NewMessageHandler(httpClient, cfg.Services.MessageURL)
router.SetupRoutes(r, healthHandler, messageHandler, authHandler, cfg.JWT.Secret)
```

### Nouveau code (simplifié)
```go
cfg := config.Load()
router.SetupRoutes(r, cfg)
```

Les anciens handlers (`handler/auth.go`, `handler/message.go`) sont dépréciés mais peuvent être conservés pour compatibilité temporaire.
