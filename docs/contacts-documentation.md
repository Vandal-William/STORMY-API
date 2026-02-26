# Documentation - Contacts & Utilisateurs Bloqués (Service User)

## Vue d'ensemble

Le module **Contact** gère les contacts de l'utilisateur et le blocage d'utilisateurs. Toutes les routes sont **protégées par JWT** (cookie `ACCESS_TOKEN` requis).

---

## Architecture des fichiers

| Fichier | Rôle |
|---------|------|
| `contact.module.ts` | Déclaration du module Contact |
| `contact.service.ts` | Logique métier (contacts + blocage) |
| `contact.controller.ts` | Routes HTTP protégées par JwtAuthGuard |
| `dto/contact.dto.ts` | DTOs de validation (AddContact, UpdateContact, BlockUser) |

---

## Routes - Contacts

### `POST /contacts` (protégée)

Ajoute un utilisateur en contact.

**Body :**

```json
{
  "contactUserId": "uuid-du-contact",
  "nickname": "Mon pote (optionnel)"
}
```

**Réponse 201 :**

```json
{
  "id": "uuid-du-contact-entry",
  "userId": "mon-uuid",
  "contactUserId": "uuid-du-contact",
  "nickname": "Mon pote",
  "createdAt": "2026-02-25T13:52:26.953Z",
  "contactUser": {
    "id": "uuid-du-contact",
    "username": "user2",
    "avatarUrl": null,
    "about": null,
    "lastSeen": null
  }
}
```

**Erreurs :**

- `400 Bad Request` : tentative de s'ajouter soi-même ou contact bloqué
- `404 Not Found` : utilisateur cible inexistant
- `409 Conflict` : contact déjà existant

---

### `GET /contacts` (protégée)

Liste paginée de mes contacts avec les infos du contact.

**Query parameters :**

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `page` | number | non | 1 | Numéro de page |
| `limit` | number | non | 20 | Résultats par page (max 100) |

**Réponse 200 :**

```json
{
  "data": [
    {
      "id": "uuid-entry",
      "userId": "mon-uuid",
      "contactUserId": "uuid-contact",
      "nickname": "Mon pote",
      "createdAt": "...",
      "contactUser": {
        "id": "uuid-contact",
        "username": "user2",
        "avatarUrl": null,
        "about": null,
        "lastSeen": null
      }
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "totalPages": 1
}
```

---

### `PATCH /contacts/:id` (protégée)

Modifie le nickname d'un contact.

**Paramètre :** `id` (UUID de l'entrée contact)

**Body :**

```json
{
  "nickname": "Best friend"
}
```

**Réponse 200 :** Contact mis à jour avec les infos du contactUser.

**Erreurs :**

- `404 Not Found` : contact non trouvé ou n'appartient pas à l'utilisateur

---

### `DELETE /contacts/:id` (protégée)

Supprime un contact.

**Paramètre :** `id` (UUID de l'entrée contact)

**Réponse 200 :**

```json
{ "deleted": true }
```

**Erreurs :**

- `404 Not Found` : contact non trouvé ou n'appartient pas à l'utilisateur

---

## Routes - Utilisateurs Bloqués

### `POST /contacts/blocked` (protégée)

Bloque un utilisateur. **Supprime automatiquement le contact des deux côtés** si existant.

**Body :**

```json
{
  "blockedUserId": "uuid-a-bloquer"
}
```

**Réponse 201 :**

```json
{
  "id": "uuid-blocage",
  "userId": "mon-uuid",
  "blockedUserId": "uuid-a-bloquer",
  "createdAt": "...",
  "blockedUser": {
    "id": "uuid-a-bloquer",
    "username": "user2",
    "avatarUrl": null,
    "about": null,
    "lastSeen": null
  }
}
```

**Erreurs :**

- `400 Bad Request` : tentative de se bloquer soi-même
- `404 Not Found` : utilisateur cible inexistant
- `409 Conflict` : utilisateur déjà bloqué

---

### `GET /contacts/blocked` (protégée)

Liste paginée des utilisateurs bloqués.

**Query parameters :** `page`, `limit` (mêmes règles que GET /contacts)

**Réponse 200 :**

```json
{
  "data": [
    {
      "id": "uuid-blocage",
      "userId": "mon-uuid",
      "blockedUserId": "uuid-bloque",
      "createdAt": "...",
      "blockedUser": {
        "id": "uuid-bloque",
        "username": "user2",
        "avatarUrl": null,
        "about": null,
        "lastSeen": null
      }
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "totalPages": 1
}
```

---

### `DELETE /contacts/blocked/:id` (protégée)

Débloque un utilisateur.

**Paramètre :** `id` (UUID de l'entrée de blocage)

**Réponse 200 :**

```json
{ "deleted": true }
```

**Erreurs :**

- `404 Not Found` : blocage non trouvé ou n'appartient pas à l'utilisateur

---

## Règles métier

| Règle | Comportement |
|-------|-------------|
| S'ajouter soi-même | 400 Bad Request |
| Ajouter un contact déjà existant | 409 Conflict |
| Ajouter un utilisateur bloqué en contact | 400 Bad Request |
| Se bloquer soi-même | 400 Bad Request |
| Bloquer un utilisateur déjà bloqué | 409 Conflict |
| Bloquer un contact | Le contact est supprimé des deux côtés automatiquement |
| Modifier/supprimer le contact d'un autre | 404 Not Found |

---

## Tester avec curl

```bash
# Prérequis : être connecté (cookies.txt contient ACCESS_TOKEN)
# Avoir l'UUID d'un autre utilisateur (via GET /profile/search)

# 1. Ajouter un contact
curl -b cookies.txt -X POST http://localhost:3000/contacts \
  -H "Content-Type: application/json" \
  -d '{"contactUserId":"uuid-du-contact","nickname":"Mon pote"}'

# 2. Lister mes contacts
curl -b cookies.txt "http://localhost:3000/contacts?page=1&limit=10"

# 3. Modifier le nickname
curl -b cookies.txt -X PATCH http://localhost:3000/contacts/uuid-entry \
  -H "Content-Type: application/json" \
  -d '{"nickname":"Best friend"}'

# 4. Supprimer un contact
curl -b cookies.txt -X DELETE http://localhost:3000/contacts/uuid-entry

# 5. Bloquer un utilisateur
curl -b cookies.txt -X POST http://localhost:3000/contacts/blocked \
  -H "Content-Type: application/json" \
  -d '{"blockedUserId":"uuid-a-bloquer"}'

# 6. Lister les bloqués
curl -b cookies.txt "http://localhost:3000/contacts/blocked"

# 7. Débloquer
curl -b cookies.txt -X DELETE http://localhost:3000/contacts/blocked/uuid-blocage
```
