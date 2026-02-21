# Guide Rapide - Architecture Message Service

## 🎯 Concept clé: Handler → Service → Repository

Chaque requête HTTP passe par exactement 3 couches:

```
Client HTTP
    ↓
[Handler] Extrait les données HTTP, valide, appelle le Service
    ↓
[Service] Logique métier, validation, orchestration
    ↓
[Repository] Persiste/récupère les données (Cassandra ou RAM)
    ↓
Storage (Base de données ou Mémoire)
```

---

## 📍 Chaque couche a une responsabilité unique

### Handler (Couche HTTP)
**Fichier:** `internal/handler/message.go` et `conversation.go`

```go
// ❌ NE FAIT PAS:
- Logique métier
- Appels multiples à la repos

// ✅ FAIT:
- Extraire les paramètres HTTP
- Parser les strings en UUID
- Valider le Content-Type
- Appeler 1x le service
- Retourner la réponse JSON
```

**Exemple:**
```go
func (h *ConversationHandler) GetConversation(c *gin.Context) {
    // 1. Parser l'ID
    id, err := gocql.ParseUUID(c.Param("id"))
    if err != nil {
        c.JSON(400, gin.H{"error": "bad ID"})
        return
    }
    
    // 2. Appeler le service UNE FOIS
    conv, err := h.service.GetConversation(ctx, id)
    
    // 3. Retourner la réponse
    c.JSON(200, conv)
}
```

### Service (Couche Métier)
**Fichier:** `internal/service/message.go` et `conversation.go`

```go
// ❌ NE FAIT PAS:
- Connaître HTTP (Gin, status codes)
- Requêtes SQL/CQL directes

// ✅ FAIT:
- Valider les règles métier
- Orchestrer les opérations multi-étapes
- Appeler le repository
- Logique "if member count < 2 { error }"
```

**Exemple:**
```go
func (s *ConversationService) CreateConversation(ctx, req) {
    // Validation métier
    if len(req.MemberIDs) < 2 {
        return nil, fmt.Errorf("need 2+ members")
    }
    
    // Orchestration multi-étape
    conv := s.repo.Create(ctx, conversation)
    for member := range req.Members {
        s.repo.AddMember(ctx, member)  // Étape 2
    }
    
    return conv
}
```

### Repository (Couche Données)
**Fichier:** `internal/repository/message.go`

```go
// ❌ NE FAIT PAS:
- Validation métier
- Logique métier

// ✅ FAIT:
- CRUD (Create, Read, Update, Delete)
- Requêtes CQL pour Cassandra
- Opérations sur les maps (In-Memory)
- Index secundaries (queries complexes)
```

**Exemple - In-Memory (Développement):**
```go
func (r *InMemoryConversationRepository) Create(ctx, conv) {
    r.conversations[conv.ID.String()] = conv
    return conv
}
```

**Exemple - Cassandra (Production):**
```go
func (r *CassandraConversationRepository) Create(ctx, conv) {
    query := "INSERT INTO conversations (id, name, ...) VALUES (...)"
    return conv
}
```

---

## 🔄 Types de Routes

### CRUD Complet

#### CREATE: POST /resource
```
HTTP Request
    ↓
Handler: Extraire JSON
    ↓ 
Service: Valider, générer ID, orchestrer
    ↓
Repository: INSERT
    → Retour: 201 Created + nouvelle ressource
```

#### READ: GET /resource/:id
```
HTTP Request
    ↓
Handler: Parser l'ID
    ↓
Service: Aucune logique (simple lookup)
    ↓
Repository: SELECT by ID
    → Retour: 200 OK + ressource
```

#### UPDATE: PUT /resource/:id
```
HTTP Request
    ↓
Handler: Parser l'ID + extraire JSON
    ↓
Service: Valider existence, modifier objet
    ↓
Repository: UPDATE
    → Retour: 200 OK + ressource modifiée
```

#### DELETE: DELETE /resource/:id
```
HTTP Request
    ↓
Handler: Parser l'ID
    ↓
Service: Valider existence
    ↓
Repository: DELETE (ou soft-delete)
    → Retour: 200 OK + message
```

---

## 🔗 Flux de Mapping des Types

Les IDs traversent le système en changeant de type:

```
HTTP Client
    ↓ (JSON avec string UUIDs)
HTTP Handler
    ↓ gocql.ParseUUID(string) → gocql.UUID
Service
    ↓ (reçoit gocql.UUID)
Repository
    ↓ Persiste avec UUID
Database (Cassandra)
    ↓ (stocke UUID, requête par UUID)
    ↓ SELECT ... WHERE id = ?
Repository
    ↓ Retour gocql.UUID
Service
    ↓ Retour *domain.Conversation
Handler
    ↓ JSON marshaling → string UUID
HTTP Response
    ↓ (JSON avec string UUIDs)
Client
```

**Important:** À chaque étape, le type doit correspondre!
- ❌ Handler ne peut pas passer string au Service
- ❌ Service ne peut pas passer UUID au Handler sans conversion
- ✅ Conversion = responsibility du Handler uniquement

---

## 📚 Routes par Catégorie

### Conversations (8 routes)

| Méthode | Route | Quoi |
|---------|-------|------|
| POST | `/conversations` | Créer |
| GET | `/conversations/:id` | Récupérer 1 |
| PUT | `/conversations/:id` | Modifier |
| DELETE | `/conversations/:id` | Supprimer |
| GET | `/users/:user_id/conversations` | Lister de l'user |
| GET | `/conversations/:id/members` | Lister les membres |
| POST | `/conversations/:id/members` | Ajouter un membre |
| DELETE | `/conversations/:id/members/:user_id` | Retirer un membre |

### Messages (5 routes)

| Méthode | Route | Quoi |
|---------|-------|------|
| POST | `/messages` | Créer |
| GET | `/messages/:id` | Récupérer 1 |
| PUT | `/messages/:id` | Modifier |
| DELETE | `/messages/:id` | Supprimer |
| GET | `/users/:user_id/messages` | Lister de l'user |

---

## 🧪 Test d'une Route Complete

### Créer une Conversation (manière guidée)

```bash
# 1. Identifier les données nécessaires
CREATOR_ID="550e8400-e29b-41d4-a716-446655440000"
MEMBER1_ID="550e8400-e29b-41d4-a716-446655440001"
MEMBER2_ID="550e8400-e29b-41d4-a716-446655440002"

# 2. Faire la requête
curl -X POST http://localhost:3001/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dev Team",
    "type": "group",
    "created_by": "'$CREATOR_ID'",
    "member_ids": ["'$MEMBER1_ID'", "'$MEMBER2_ID'"]
  }'

# 3. Résultat attendu (201 Created)
# {
#   "id": "GENERATED_UUID",
#   "name": "Dev Team",
#   "type": "group",
#   "created_by": "550e8400-...",
#   "created_at": "2026-02-20T17:10:40Z"
# }

# 4. Récupérer la conversation créée
CONV_ID="GENERATED_UUID_FROM_ABOVE"
curl -X GET http://localhost:3001/conversations/$CONV_ID

# 5. Modifier la conversation
curl -X PUT http://localhost:3001/conversations/$CONV_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Dev Team"}'

# 6. Ajouter un membre
curl -X POST http://localhost:3001/conversations/$CONV_ID/members \
  -H "Content-Type: application/json" \
  -d '{"user_id": "'$MEMBER1_ID'"}'

# 7. Lister les membres
curl -X GET http://localhost:3001/conversations/$CONV_ID/members

# 8. Supprimer un membre
curl -X DELETE http://localhost:3001/conversations/$CONV_ID/members/$MEMBER1_ID

# 9. Supprimer la conversation
curl -X DELETE http://localhost:3001/conversations/$CONV_ID

# 10. Vérifier suppression (404)
curl -X GET http://localhost:3001/conversations/$CONV_ID
```

---

## 🛠️ Ajouter une Nouvelle Route

### Checklist (10 points)

1. **Décider le HTTP method** (POST/GET/PUT/DELETE)
2. **Décider la route** (`/resource/:id/action`)
3. **Ajouter au handler:**
   ```go
   func (h *Handler) ActionName(c *gin.Context) {
       // Parse params → UUID
       // Extraire JSON
       // Appeler 1x service
       // Retourner réponse
   }
   ```
4. **Ajouter au service:**
   ```go
   func (s *Service) ActionName(ctx, id UUID) (..., error) {
       // Validation métier
       // Appeler repo
       // Orchestrer multi-étapes si besoin
   }
   ```
5. **Ajouter au repository (interface):**
   ```go
   type Repository interface {
       ActionName(ctx, id UUID) (..., error)
   }
   ```
6. **Implémenter dans In-Memory repo:**
   ```go
   func (r *InMemory...) ActionName(ctx, id UUID) (..., error) {
       // Map operations
   }
   ```
7. **Implémenter dans Cassandra repo:**
   ```go
   func (r *Cassandra...) ActionName(ctx, id UUID) (..., error) {
       // CQL queries
   }
   ```
8. **Ajouter dans router:**
   ```go
   group.POST("/path", handler.ActionName)
   ```
9. **Ajouter des tests handler:**
   ```go
   func TestActionName(t *testing.T) { ... }
   ```
10. **Ajouter des tests service:**
   ```go
   func TestActionName(t *testing.T) { ... }
   ```

---

## 💡 Points Clés à Retenir

1. **Handler ≠ Service ≠ Repository**
   - Handler: HTTP plumbing
   - Service: Business logic
   - Repository: Data access

2. **UUID parsing se fait UNE FOIS**
   - Dans le Handler
   - Tous les autres reçoivent gocql.UUID

3. **Chaque layer appelle le suivant UNE FOIS** (ou orchestré sans boucles)
   - Handler → Service (1x)
   - Service → Repository (1x par opération logique)

4. **Erreurs remontent:**
   ```go
   err := repository.Create()
   if err != nil {
       return nil, err  // Service retourne l'erreur
   }
   // Handler reçoit l'erreur et choisit le code HTTP
   ```

5. **Data flows:**
   - Donnée entre les couches = structures Go typées
   - JSON ↔ HTTP handler = conversion JSON automatique (Gin)
   - UUID string ↔ UUID = conversion explicite (ParseUUID)

---

## 📖 Lectures complémentaires

- `API_DOCUMENTATION.md` - Documentation complète (routes, exemples)
- `internal/domain/models.go` - Structures de données
- `internal/handler/*.go` - Implémentations HTTP
- `internal/service/*.go` - Logique métier
- `internal/repository/message.go` - Interfaces + In-Memory
