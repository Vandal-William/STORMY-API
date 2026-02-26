# Documentation - Service Moderation

## Vue d'ensemble

Le service **Moderation** (port 3004) gère les signalements d'utilisateurs/messages et les bannissements. C'est un service interne appelé par le gateway ou d'autres services (pas d'authentification JWT directe).

---

## Architecture des fichiers

| Fichier | Rôle |
|---------|------|
| `report/report.module.ts` | Module des signalements |
| `report/report.controller.ts` | Endpoints REST signalements |
| `report/report.service.ts` | Logique métier signalements |
| `report/dto/create-report.dto.ts` | DTO création de signalement |
| `report/dto/update-report.dto.ts` | DTO mise à jour du statut |
| `user-ban/user-ban.module.ts` | Module des bannissements |
| `user-ban/user-ban.controller.ts` | Endpoints REST bannissements |
| `user-ban/user-ban.service.ts` | Logique métier bannissements |
| `user-ban/dto/create-ban.dto.ts` | DTO création de ban |

---

## Routes - Signalements (`/reports`)

### `POST /reports`

Crée un nouveau signalement.

**Body :**

```json
{
  "reporterId": "uuid-du-reporter",
  "reportedUserId": "uuid-signale (optionnel)",
  "reportedMessageId": "uuid-message (optionnel)",
  "conversationId": "uuid-conversation (optionnel)",
  "reason": "spam",
  "description": "Description optionnelle (max 1000 chars)"
}
```

**Valeurs possibles pour `reason` :** `spam`, `harassment`, `inappropriate`, `fake_account`, `other`

**Réponse 201 :**

```json
{
  "id": "uuid",
  "reporterId": "...",
  "reportedUserId": "...",
  "reportedMessageId": null,
  "conversationId": null,
  "reason": "spam",
  "description": "...",
  "status": "pending",
  "createdAt": "2026-02-26T09:52:06.311Z"
}
```

---

### `GET /reports`

Liste paginée des signalements avec filtres optionnels.

**Query parameters :**

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `page` | number | non | 1 | Numéro de page |
| `limit` | number | non | 20 | Résultats par page (max 100) |
| `status` | string | non | - | Filtre par statut (`pending`, `reviewed`, `resolved`, `dismissed`) |
| `reason` | string | non | - | Filtre par raison (`spam`, `harassment`, etc.) |

**Réponse 200 :**

```json
{
  "data": [...],
  "total": 1,
  "page": 1,
  "limit": 20,
  "totalPages": 1
}
```

---

### `GET /reports/:id`

Détail d'un signalement.

**Erreurs :** `404 Not Found` si inexistant

---

### `PATCH /reports/:id`

Met à jour le statut d'un signalement.

**Body :**

```json
{
  "status": "reviewed"
}
```

**Valeurs possibles :** `pending`, `reviewed`, `resolved`, `dismissed`

**Erreurs :** `404 Not Found` si inexistant

---

### `DELETE /reports/:id`

Supprime un signalement.

**Réponse 200 :** `{ "deleted": true }`

**Erreurs :** `404 Not Found` si inexistant

---

## Routes - Bannissements (`/bans`)

### `POST /bans`

Bannit un utilisateur. Si `expiresAt` est omis, le ban est permanent.

**Body :**

```json
{
  "userId": "uuid-a-bannir",
  "reason": "Raison du ban (optionnel, max 500 chars)",
  "expiresAt": "2026-03-26T00:00:00.000Z (optionnel, ISO 8601)"
}
```

**Réponse 201 :**

```json
{
  "id": "uuid-ban",
  "userId": "...",
  "reason": "...",
  "expiresAt": "2026-03-26T00:00:00.000Z",
  "createdAt": "..."
}
```

**Erreurs :**

- `409 Conflict` : utilisateur déjà banni (ban actif)

---

### `GET /bans`

Liste paginée des bannissements.

**Query parameters :**

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `page` | number | non | 1 | Numéro de page |
| `limit` | number | non | 20 | Résultats par page (max 100) |
| `active` | string | non | - | `true` = bans actifs uniquement, `false` = bans expirés uniquement |

---

### `GET /bans/check/:userId`

Vérifie si un utilisateur est actuellement banni (ban non expiré ou permanent).

**Réponse 200 (non banni) :**

```json
{ "banned": false }
```

**Réponse 200 (banni) :**

```json
{
  "banned": true,
  "ban": {
    "id": "uuid-ban",
    "userId": "...",
    "reason": "...",
    "expiresAt": "2026-03-26T00:00:00.000Z",
    "createdAt": "..."
  }
}
```

---

### `GET /bans/:id`

Détail d'un bannissement.

**Erreurs :** `404 Not Found` si inexistant

---

### `DELETE /bans/:id`

Lève un bannissement (supprime l'entrée).

**Réponse 200 :** `{ "deleted": true }`

**Erreurs :** `404 Not Found` si inexistant

---

## Règles métier

| Règle | Comportement |
|-------|-------------|
| Ban doublon | 409 Conflict si un ban actif existe déjà pour cet utilisateur |
| Ban permanent | `expiresAt` = null → le ban n'expire jamais |
| Ban temporaire | `expiresAt` défini → le ban expire automatiquement (vérifié au check) |
| Vérification ban | Seuls les bans non expirés ou permanents comptent |

---

## Tester avec curl

```bash
# 1. Créer un signalement
curl -X POST http://localhost:3004/reports \
  -H "Content-Type: application/json" \
  -d '{"reporterId":"uuid-reporter","reportedUserId":"uuid-signale","reason":"spam","description":"Spam dans le groupe"}'

# 2. Lister les signalements (filtre par statut)
curl "http://localhost:3004/reports?status=pending&page=1&limit=10"

# 3. Mettre à jour le statut
curl -X PATCH http://localhost:3004/reports/uuid-report \
  -H "Content-Type: application/json" \
  -d '{"status":"resolved"}'

# 4. Supprimer un signalement
curl -X DELETE http://localhost:3004/reports/uuid-report

# 5. Bannir un utilisateur (temporaire)
curl -X POST http://localhost:3004/bans \
  -H "Content-Type: application/json" \
  -d '{"userId":"uuid-user","reason":"Spam","expiresAt":"2026-03-26T00:00:00.000Z"}'

# 6. Bannir un utilisateur (permanent)
curl -X POST http://localhost:3004/bans \
  -H "Content-Type: application/json" \
  -d '{"userId":"uuid-user","reason":"Compte fake"}'

# 7. Vérifier si un utilisateur est banni
curl http://localhost:3004/bans/check/uuid-user

# 8. Lister les bans actifs
curl "http://localhost:3004/bans?active=true"

# 9. Lever un ban
curl -X DELETE http://localhost:3004/bans/uuid-ban
```
