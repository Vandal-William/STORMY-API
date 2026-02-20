# Documentation API Message Service

## 📋 Table des matières

1. [Architecture Générale](#architecture-générale)
2. [Flux d'une Requête](#flux-dune-requête)
3. [Composants](#composants)
4. [Routes Conversations](#routes-conversations)
5. [Routes Messages](#routes-messages)
6. [Modèles de Données](#modèles-de-données)
7. [Gestion des Erreurs](#gestion-des-erreurs)
8. [Exemples d'Utilisation](#exemples-dutilisation)

---

## Architecture Générale

Le service message suit le **Repository Pattern** avec une architecture en couches :

```
┌─────────────────────────────────────────────────────────┐
│                  HTTP Requests (Clients)                 │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              HTTP Handlers (router.go)                   │
│  Responsabilités:                                        │
│  - Extraire les paramètres des requêtes HTTP             │
│  - Parser les IDs string en UUID                         │
│  - Valider le Content-Type JSON                          │
│  - Appeler le service métier approprié                   │
│  - Retourner les réponses HTTP formatées                 │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│           Business Logic Services (*.go)                │
│  Responsabilités:                                        │
│  - Implémenter la logique métier                         │
│  - Valider les données métier                            │
│  - Orchestrer les opérations multi-étapes               │
│  - Appeler les repositories pour l'accès données        │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│      Data Access Layer - Repositories (message.go)      │
│  Responsabilités:                                        │
│  - Abstraction de la source de données                   │
│  - CRUD operations (Create, Read, Update, Delete)        │
│  - Pagination et filtrage                               │
│  - Interactions avec Cassandra ou RAM                    │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│            Data Storage Layer                            │
│  Implementations:                                         │
│  - InMemoryMessageRepository (développement/tests)      │
│  - InMemoryConversationRepository (développement/tests)  │
│  - CassandraMessageRepository (production)              │
│  - CassandraConversationRepository (production)          │
└────────────────────────────────────────────────────────┘
```

---

## Flux d'une Requête

### Exemple: Créer une Conversation

```
1. HTTP REQUEST (Client)
   POST /conversations
   Content-Type: application/json
   {
     "name": "Team Meeting",
     "type": "group",
     "created_by": "550e8400-e29b-41d4-a716-446655440000",
     "member_ids": ["550e8400-e29b-41d4-a716-446655440001", ...]
   }

2. HANDLER (handler/conversation.go:CreateConversation)
   ├─ Extraire la requête JSON
   ├─ Valider le format JSON
   ├─ Parser les UUIDs pour validation
   └─ Appeler conversationService.CreateConversation()

3. SERVICE (service/conversation.go:CreateConversation)
   ├─ Valider: au moins 2 membres et un créateur
   ├─ Créer un objet domain.Conversation
   ├─ Appeler conversationRepo.Create()
   ├─ Pour chaque membre:
   │  └─ Appeler conversationRepo.AddMember()
   └─ Retourner la conversation créée

4. REPOSITORY (repository/message.go:Create)
   ├─ Implém. InMemory: Stocker dans map[string]*Conversation
   ├─ Implém. Cassandra: INSERT INTO conversations VALUES(...)
   └─ Retourner la conversation avec ID généré

5. RESPONSE
   HTTP 201 Created
   Content-Type: application/json
   {
     "id": "550e8400-...",
     "name": "Team Meeting",
     "type": "group",
     "created_by": "550e8400-...",
     "created_at": "2026-02-20T17:10:40Z"
   }
```

---

## Composants

### 1. Domain (internal/domain/models.go)

Les modèles de données représentent l'état métier. Tous les IDs utilisent `gocql.UUID` pour compatibilité Cassandra.

**Conversation:**
```go
type Conversation struct {
    ID          gocql.UUID  // Identifiant unique
    Type        string      // 'private' ou 'group'
    Name        string      // Nom de la conversation
    Description string      // Description optionnelle
    AvatarURL   string      // Avatar optionnel
    CreatedBy   gocql.UUID  // Créateur de la conversation
    CreatedAt   time.Time   // Quand créée
    UpdatedAt   *time.Time  // Dernière modification
}
```

**Message:**
```go
type Message struct {
    ID             gocql.UUID              // Identifiant unique
    ConversationID gocql.UUID              // Conversation parente
    SenderID       gocql.UUID              // Auteur du message
    Content        string                  // Contenu
    Type           string                  // 'text', 'image', etc.
    ReplyToID      *gocql.UUID             // Message parent (reply)
    IsForwarded    bool                    // Message transféré
    IsEdited       bool                    // Modifié après création
    IsDeleted      bool                    // Soft-deleted
    CreatedAt      time.Time               // Quand créé
    UpdatedAt      *time.Time              // Dernière modification
    Attachments    []MessageAttachment     // Pièces jointes
}
```

### 2. Handler (internal/handler/*.go)

Interface HTTP. Responsable de:
- Recevoir les requêtes HTTP
- Parser les UUIDs depuis les URL parameters (string)
- Valider le Content-Type et les données
- Appeler le service
- Retourner les réponses HTTP avec codes de statut

**Exemple:**
```go
// GetConversation récupère une conversation par ID
func (h *ConversationHandler) GetConversation(c *gin.Context) {
    // 1. Extraire l'ID depuis l'URL parameter
    idStr := c.Param("id")
    
    // 2. Parser le string en gocql.UUID
    id, err := gocql.ParseUUID(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID format"})
        return
    }
    
    // 3. Appeler le service avec l'UUID typé
    conversation, err := h.conversationService.GetConversation(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 4. Retourner la réponse
    c.JSON(http.StatusOK, conversation)
}
```

### 3. Service (internal/service/*.go)

Logique métier. Responsable de:
- Valider les règles métier
- Orchestrer les opérations multi-étapes
- Gérer les transactions logiques
- Appeler le repository pour les données

**Exemple:**
```go
// CreateConversation crée une nouvelle conversation avec membres
func (s *ConversationService) CreateConversation(ctx context.Context, 
    req *domain.CreateConversationRequest) (*domain.Conversation, error) {
    
    // Validation métier
    if len(req.MemberIDs) < 2 || len(req.CreatedBy) == 0 {
        return nil, fmt.Errorf("conversation must have at least 2 members")
    }
    
    // Créer l'objet conversation
    conversation := &domain.Conversation{
        ID:        gocql.TimeUUID(),  // Générer un UUID basé sur le temps
        CreatedBy: req.CreatedBy,
        Name:      req.Name,
        Type:      req.Type,
    }
    
    // Persister la conversation
    createdConv, err := s.conversationRepo.Create(ctx, conversation)
    if err != nil {
        return nil, err
    }
    
    // Ajouter chaque membre (opération multi-étapes)
    for _, memberID := range req.MemberIDs {
        member := &domain.ConversationMember{
            ConversationID: createdConv.ID,
            UserID:         memberID,
            JoinedAt:       createdConv.CreatedAt,
        }
        if err := s.conversationRepo.AddMember(ctx, member); err != nil {
            return nil, err
        }
    }
    
    return createdConv, nil
}
```

### 4. Repository (internal/repository/message.go)

Abstraction de données. Responsable de:
- Isoler la logique de persistance
- Implémenter CRUD operations
- Supporter plusieurs backends (In-Memory, Cassandra)
- Gérer les requêtes complexes

**Interface:**
```go
type ConversationRepository interface {
    Create(ctx context.Context, conversation *domain.Conversation) 
        (*domain.Conversation, error)
    
    GetByID(ctx context.Context, id gocql.UUID) 
        (*domain.Conversation, error)
    
    GetByUserID(ctx context.Context, userID gocql.UUID) 
        ([]domain.Conversation, error)
    
    Update(ctx context.Context, id gocql.UUID, conversation *domain.Conversation) 
        (*domain.Conversation, error)
    
    Delete(ctx context.Context, id gocql.UUID) error
    
    AddMember(ctx context.Context, member *domain.ConversationMember) error
    RemoveMember(ctx context.Context, conversationID, userID gocql.UUID) error
    GetMembers(ctx context.Context, conversationID gocql.UUID) 
        ([]domain.ConversationMember, error)
}
```

**Implémentations:**

**InMemory (Développement/Test):**
```go
type InMemoryConversationRepository struct {
    conversations map[string]*domain.Conversation
    members       map[string]*domain.ConversationMember
}

func (r *InMemoryConversationRepository) Create(ctx context.Context, 
    conversation *domain.Conversation) (*domain.Conversation, error) {
    id := conversation.ID.String()
    conversation.CreatedAt = time.Now()
    r.conversations[id] = conversation
    return conversation, nil
}
```

**Cassandra (Production):**
```go
type CassandraConversationRepository struct {
    session *gocql.Session
}

func (r *CassandraConversationRepository) Create(ctx context.Context, 
    conversation *domain.Conversation) (*domain.Conversation, error) {
    query := `INSERT INTO conversations 
        (id, type, name, created_by, created_at) 
        VALUES (?, ?, ?, ?, ?)`
    
    return conversation, r.session.Query(query,
        conversation.ID,
        conversation.Type,
        conversation.Name,
        conversation.CreatedBy,
        conversation.CreatedAt,
    ).Exec()
}
```

---

## Routes Conversations

### 1. POST `/conversations` - Créer une Conversation

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Extraire JSON, valider Content-Type |
| **Service** | Valider les règles métier (2+ membres, créateur obligatoire) |
| **Repository** | Insérer dans conversations + ajouter les membres |

**Requête:**
```json
POST /conversations
Content-Type: application/json

{
  "name": "Team Meeting",
  "type": "group",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "member_ids": [
    "550e8400-e29b-41d4-a716-446655440001",
    "550e8400-e29b-41d4-a716-446655440002"
  ]
}
```

**Réponse (201):**
```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "name": "Team Meeting",
  "type": "group",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-02-20T17:10:40Z"
}
```

**Codes de retour:**
- `201 Created` - Conversation créée avec succès
- `400 Bad Request` - Données invalides (JSON malformé)
- `500 Internal Server Error` - Erreur serveur

---

### 2. GET `/conversations/:id` - Récupérer une Conversation

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID string en UUID, valider le format |
| **Service** | Aucune logique (simple lookup) |
| **Repository** | Récupérer par ID depuis la source de données |

**Requête:**
```
GET /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8
```

**Réponse (200):**
```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "name": "Team Meeting",
  "type": "group",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-02-20T17:10:40Z"
}
```

**Codes de retour:**
- `200 OK` - Conversation trouvée
- `400 Bad Request` - ID format invalide
- `404 Not Found` - Conversation n'existe pas
- `500 Internal Server Error` - Erreur serveur

---

### 3. GET `/users/:user_id/conversations` - Lister les Conversations d'un User

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'user_id string en UUID, contrôler les query params |
| **Service** | Aucune logique |
| **Repository** | Récupérer toutes les conversations de l'utilisateur (index secondaire) |

**Requête:**
```
GET /users/550e8400-e29b-41d4-a716-446655440000/conversations
```

**Réponse (200):**
```json
{
  "conversations": [
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "name": "Team Meeting",
      "type": "group",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-02-20T17:10:40Z"
    }
  ]
}
```

---

### 4. PUT `/conversations/:id` - Modifier une Conversation

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID, extraire JSON de modification |
| **Service** | Valider l'existence, valider les règles métier |
| **Repository** | Mettre à jour dans la source de données |

**Requête:**
```json
PUT /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8
Content-Type: application/json

{
  "name": "Updated Team Meeting"
}
```

**Réponse (200):**
```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "name": "Updated Team Meeting",
  "type": "group",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-02-20T17:10:40Z",
  "updated_at": "2026-02-20T18:20:30Z"
}
```

---

### 5. DELETE `/conversations/:id` - Supprimer une Conversation

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID |
| **Service** | Valider l'existence |
| **Repository** | Supprimer de la source de données (soft-delete possible) |

**Requête:**
```
DELETE /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8
```

**Réponse (200):**
```json
{
  "message": "conversation deleted successfully"
}
```

---

### 6. GET `/conversations/:id/members` - Lister les Membres

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID de conversation |
| **Service** | Aucune logique |
| **Repository** | Récupérer tous les membres (avec joint + left) |

**Requête:**
```
GET /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8/members
```

**Réponse (200):**
```json
{
  "members": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "role": "member",
      "is_muted": false,
      "joined_at": "2026-02-20T17:10:40Z"
    }
  ]
}
```

---

### 7. POST `/conversations/:id/members` - Ajouter un Membre

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID conversation, extraire l'user_id JSON |
| **Service** | Valider l'existence de la conversation |
| **Repository** | Insérer le member-user dans conversation_members |

**Requête:**
```json
POST /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8/members
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

**Réponse (200):**
```json
{
  "message": "member added successfully"
}
```

---

### 8. DELETE `/conversations/:id/members/:user_id` - Retirer un Membre

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser les IDs conversation et user |
| **Service** | Valider l'existence |
| **Repository** | Supprimer ou soft-delete le member |

**Requête:**
```
DELETE /conversations/6ba7b810-9dad-11d1-80b4-00c04fd430c8/members/550e8400-e29b-41d4-a716-446655440003
```

**Réponse (200):**
```json
{
  "message": "member removed successfully"
}
```

---

## Routes Messages

### 1. POST `/messages` - Créer un Message

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Extraire JSON, valider Content-Type |
| **Service** | Valider les données (conversation existe, sender existe) |
| **Repository** | Insérer le message + entrée de status pour chaque member |

**Requête:**
```json
POST /messages
Content-Type: application/json

{
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "sender_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello everyone!",
  "type": "text",
  "reply_to_id": null
}
```

**Réponse (201):**
```json
{
  "id": "7ba7b811-9dad-11d1-80b4-00c04fd430c9",
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "sender_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello everyone!",
  "type": "text",
  "is_forwarded": false,
  "is_edited": false,
  "is_deleted": false,
  "created_at": "2026-02-20T17:10:40Z"
}
```

---

### 2. GET `/messages/:id` - Récupérer un Message

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID string en UUID |
| **Service** | Aucune logique |
| **Repository** | Récupérer par ID avec attachments et status |

**Requête:**
```
GET /messages/7ba7b811-9dad-11d1-80b4-00c04fd430c9
```

**Réponse (200):**
```json
{
  "id": "7ba7b811-9dad-11d1-80b4-00c04fd430c9",
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "sender_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Hello everyone!",
  "type": "text",
  "created_at": "2026-02-20T17:10:40Z"
}
```

---

### 3. GET `/users/:user_id/messages` - Lister les Messages d'un User

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser user_id, gérer la pagination (limit) |
| **Service** | Aucune logique |
| **Repository** | Récupérer les messages de l'utilisateur (par sender) avec limite |

**Requête:**
```
GET /users/550e8400-e29b-41d4-a716-446655440001/messages?limit=50
```

**Réponse (200):**
```json
{
  "messages": [
    {
      "id": "7ba7b811-9dad-11d1-80b4-00c04fd430c9",
      "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "sender_id": "550e8400-e29b-41d4-a716-446655440001",
      "content": "Hello everyone!",
      "created_at": "2026-02-20T17:10:40Z"
    }
  ]
}
```

---

### 4. PUT `/messages/:id` - Modifier un Message

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID, extraire JSON de modification |
| **Service** | Valider l'existence, marquer comme édité |
| **Repository** | Mettre à jour le message |

**Requête:**
```json
PUT /messages/7ba7b811-9dad-11d1-80b4-00c04fd430c9
Content-Type: application/json

{
  "content": "Updated message content"
}
```

**Réponse (200):**
```json
{
  "id": "7ba7b811-9dad-11d1-80b4-00c04fd430c9",
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "sender_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "Updated message content",
  "is_edited": true,
  "updated_at": "2026-02-20T18:20:30Z"
}
```

---

### 5. DELETE `/messages/:id` - Supprimer un Message

**Responsabilités dans le flux:**

| Couche | Action |
|-------|--------|
| **Handler** | Parser l'ID |
| **Service** | Valider l'existence |
| **Repository** | Soft-delete: marquer is_deleted=true |

**Requête:**
```
DELETE /messages/7ba7b811-9dad-11d1-80b4-00c04fd430c9
```

**Réponse (200):**
```json
{
  "message": "message deleted successfully"
}
```

---

## Modèles de Données

### Conversation

| Champ | Type | Description |
|-------|------|-------------|
| ID | UUID | Identifiant unique |
| Type | String | 'private' ou 'group' |
| Name | String | Nom de la conversation |
| Description | String | Description détaillée |
| AvatarURL | String | URL de l'avatar |
| CreatedBy | UUID | Utilisateur qui a créé la conversation |
| CreatedAt | DateTime | Quand créée |
| UpdatedAt | DateTime* | Dernière modification |

### Message

| Champ | Type | Description |
|-------|------|-------------|
| ID | UUID | Identifiant unique |
| ConversationID | UUID | Conversation parente |
| SenderID | UUID | Auteur du message |
| Content | String | Contenu du message |
| Type | String | 'text', 'image', 'video', etc. |
| ReplyToID | UUID* | Message auquel on répond |
| IsForwarded | Boolean | Message transféré |
| IsEdited | Boolean | Message modifié |
| IsDeleted | Boolean | Supprimé logiquement (soft-delete) |
| CreatedAt | DateTime | Quand créé |
| UpdatedAt | DateTime* | Dernière modification |

### ConversationMember

| Champ | Type | Description |
|-------|------|-------------|
| ID | UUID | Identifiant unique |
| ConversationID | UUID | Conversation |
| UserID | UUID | Utilisateur membre |
| Role | String | 'owner', 'admin', 'member' |
| IsMuted | Boolean | Notifications silencieuses |
| JoinedAt | DateTime | Quand rejoint |
| LeftAt | DateTime* | Quand parti (null = encore membre) |

---

## Gestion des Erreurs

### Codes HTTP standards

| Code | Sens | Quand |
|------|------|-------|
| 200 OK | Succès | Requête réussie (GET, PUT, DELETE) |
| 201 Created | Création | Ressource créée (POST) |
| 400 Bad Request | Erreur client | Données invalides, ID format mauvais |
| 404 Not Found | Introuvable | Ressource n'existe pas |
| 500 Internal Server Error | Erreur serveur | Erreur non gérée dans le code |

### Formats d'erreur

**Erreur simple:**
```json
{
  "error": "invalid conversation ID format"
}
```

**Erreur avec détails (à implémenter):**
```json
{
  "error": "validation_error",
  "details": {
    "member_ids": "must have at least 2 members"
  }
}
```

---

## Exemples d'Utilisation

### Scénario complet: Créer une conversation et envoyer un message

```bash
# 1. Créer une conversation
curl -X POST http://localhost:3001/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Project Discussion",
    "type": "group",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "member_ids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"]
  }'

# Réponse: 
# {
#   "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
#   "name": "Project Discussion",
#   ...
# }

# 2. Envoyer un message
CONV_ID="6ba7b810-9dad-11d1-80b4-00c04fd430c8"
curl -X POST http://localhost:3001/messages \
  -H "Content-Type: application/json" \
  -d "{
    \"conversation_id\": \"$CONV_ID\",
    \"sender_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"content\": \"Let's discuss the project\",
    \"type\": \"text\"
  }"

# Réponse:
# {
#   "id": "7ba7b811-9dad-11d1-80b4-00c04fd430c9",
#   "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
#   ...
# }

# 3. Récupérer les messages d'un utilisateur
curl -X GET http://localhost:3001/users/550e8400-e29b-41d4-a716-446655440000/messages

# 4. Modifier le message
MSG_ID="7ba7b811-9dad-11d1-80b4-00c04fd430c9"
curl -X PUT http://localhost:3001/messages/$MSG_ID \
  -H "Content-Type: application/json" \
  -d '{"content": "Updated: Let'\''s discuss the project in detail"}'

# 5. Ajouter un nouveau membre
curl -X POST http://localhost:3001/conversations/$CONV_ID/members \
  -H "Content-Type: application/json" \
  -d '{"user_id": "550e8400-e29b-41d4-a716-446655440003"}'

# 6. Lister les membres
curl -X GET http://localhost:3001/conversations/$CONV_ID/members

# 7. Supprimer le message
curl -X DELETE http://localhost:3001/messages/$MSG_ID

# 8. Supprimer la conversation
curl -X DELETE http://localhost:3001/conversations/$CONV_ID
```

---

## Architecture du Projet

```
services/message/
├── cmd/
│   └── api/
│       └── main.go              # Point d'entrée (initialise services, handlers, routes)
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration (Cassandra, Redis, NATS)
│   ├── domain/
│   │   └── models.go            # Modèles métier (Conversation, Message, etc.)
│   ├── handler/
│   │   ├── conversation.go       # HTTP handlers pour conversations
│   │   ├── conversation_test.go  # Tests des handlers
│   │   ├── message.go            # HTTP handlers pour messages
│   │   ├── health.go             # Health check endpoint
│   │   └── message_test.go       # Tests des handlers messages
│   ├── service/
│   │   ├── conversation.go       # Logique métier conversations
│   │   ├── conversation_test.go  # Tests du service
│   │   └── message.go            # Logique métier messages
│   ├── repository/
│   │   ├── message.go            # Interfaces + In-Memory implementations
│   │   ├── cassandra_*.go        # Implémentations Cassandra
│   │   └── *_test.go             # Tests des repositories
│   ├── router/
│   │   └── router.go             # Définition des routes Gin
│   ├── middleware/
│   │   └── logger.go             # Middleware de logging
│   └── infrastructure/
│       └── cassandra/            # Client Cassandra, connexions
├── db/
│   └── cassandra/
│       └── schema.cql            # Schéma Cassandra (7 tables)
├── k8s/
│   ├── message.yaml              # Deployment + Service K8s
│   └── kustomization.yaml
├── Dockerfile                    # Image Docker multi-stage
├── go.mod / go.sum               # Dependencies Go
└── README.md / Documentation
```

---

## Pattern architectural: Repository Pattern

Le Repository Pattern fournit une abstraction entre la logique métier et la persistance:

**Avantages:**
- ✅ Interchangeabilité: In-Memory pour dev, Cassandra pour prod
- ✅ Testabilité: Mocker facilement les données
- ✅ Maintenance: Changements de BD sans affecter la logique métier
- ✅ Séparation des responsabilités: Chaque couche a un seul rôle

**Flux:**
```
Service ← Interface Repository ← Implémentation (In-Memory ou Cassandra)
```

---

## Type UUID et Security

Tous les IDs utilisent le type `gocql.UUID` pour:
- ✅ Compatibilité Cassandra native
- ✅ Sécurité: UUIDs non-séquentiels (vs. auto-increment prédictibles)
- ✅ Distribution: Pas besoin d'ID centralisé
- ✅ TimeUUID: Permet de trier par création

**Conversion string ↔ UUID:**
```go
// String → UUID (HTTP → Service)
idStr := c.Param("id")  // "550e8400-e29b-41d4-a716-446655440000"
id, err := gocql.ParseUUID(idStr)

// UUID → String (Service → JSON)
jsonID := message.ID.String()  // "550e8400-e29b-41d4-a716-446655440000"
```

---

## Prochaines étapes d'intégration

- [ ] Remplacer In-Memory par Cassandra repositories
- [ ] Ajouter authentification (JWT) aux handlers
- [ ] Implémenter les webhooks pour les notifications (NATS)
- [ ] Ajouter Redis pour le caching
- [ ] Implémenter la pagination avancée
- [ ] Ajouter full-text search pour les messages
- [ ] Intégrer avec le service user pour validation des IDs
