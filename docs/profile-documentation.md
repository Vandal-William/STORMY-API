# Documentation - Profil Utilisateur (Service User)

## Vue d'ensemble

Le module **Profile** gère le CRUD du profil utilisateur. Toutes les routes sont **protégées par JWT** (cookie `ACCESS_TOKEN` requis).

La création de profil est gérée par `POST /auth/register` (voir [auth-documentation.md](auth-documentation.md)).

---

## Architecture des fichiers

| Fichier | Rôle |
|---------|------|
| `profile.module.ts` | Déclaration du module Profile |
| `profile.service.ts` | Logique métier (lecture, mise à jour, recherche, suppression) |
| `profile.controller.ts` | Routes HTTP protégées par JwtAuthGuard |
| `dto/update-profile.dto.ts` | Validation des données de mise à jour |

---

## Routes disponibles

Toutes les routes nécessitent le cookie `ACCESS_TOKEN` valide.

### `GET /profile/me` (protégée)

Retourne le profil complet de l'utilisateur connecté.

**Réponse 200 :**

```json
{
  "id": "uuid",
  "phone": "0612345678",
  "username": "monuser",
  "email": "user@mail.com",
  "avatarUrl": "https://example.com/avatar.jpg",
  "about": "Hello world",
  "lastSeen": "2026-02-25T10:00:00.000Z",
  "createdAt": "2026-02-24T08:34:54.297Z"
}
```

**Erreurs :**

- `401 Unauthorized` : cookie absent ou token invalide/expiré
- `404 Not Found` : utilisateur non trouvé

---

### `GET /profile/search?username=xxx` (protégée)

Recherche d'utilisateurs par nom d'utilisateur (correspondance partielle, insensible à la casse). Résultats paginés.

**Query parameters :**

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `username` | string | oui | - | Terme de recherche |
| `page` | number | non | 1 | Numéro de page |
| `limit` | number | non | 20 | Résultats par page (max 100) |

**Réponse 200 :**

```json
{
  "data": [
    {
      "id": "uuid",
      "username": "monuser",
      "avatarUrl": "https://example.com/avatar.jpg",
      "about": "Hello world",
      "lastSeen": "2026-02-25T10:00:00.000Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "totalPages": 1
}
```

**Erreurs :**

- `400 Bad Request` : paramètre `username` manquant ou vide
- `401 Unauthorized` : cookie absent ou token invalide/expiré

---

### `GET /profile/:id` (protégée)

Retourne le profil public d'un autre utilisateur (informations limitées : pas d'email ni de téléphone).

**Paramètre :** `id` (UUID)

**Réponse 200 :**

```json
{
  "id": "uuid",
  "username": "autreuser",
  "avatarUrl": "https://example.com/avatar.jpg",
  "about": "Hello world",
  "lastSeen": "2026-02-25T10:00:00.000Z"
}
```

**Erreurs :**

- `400 Bad Request` : id n'est pas un UUID valide
- `401 Unauthorized` : cookie absent ou token invalide/expiré
- `404 Not Found` : utilisateur non trouvé

---

### `PATCH /profile/me` (protégée)

Met à jour le profil de l'utilisateur connecté. Tous les champs sont optionnels.

**Body :**

```json
{
  "username": "nouveau_username",
  "email": "nouveau@mail.com",
  "avatarUrl": "https://example.com/new-avatar.jpg",
  "about": "Nouvelle bio"
}
```

**Validation :**

| Champ | Règles |
|-------|--------|
| `username` | Optionnel, string, 3-50 caractères |
| `email` | Optionnel, format email valide, max 255 caractères |
| `avatarUrl` | Optionnel, string, max 500 caractères |
| `about` | Optionnel, string, max 500 caractères |

**Réponse 200 :**

```json
{
  "id": "uuid",
  "phone": "0612345678",
  "username": "nouveau_username",
  "email": "nouveau@mail.com",
  "avatarUrl": "https://example.com/new-avatar.jpg",
  "about": "Nouvelle bio",
  "lastSeen": "2026-02-25T10:00:00.000Z",
  "createdAt": "2026-02-24T08:34:54.297Z"
}
```

**Erreurs :**

- `400 Bad Request` : données invalides
- `401 Unauthorized` : cookie absent ou token invalide/expiré
- `404 Not Found` : utilisateur non trouvé
- `409 Conflict` : nom d'utilisateur déjà pris

---

### `DELETE /profile/me` (protégée)

Supprime définitivement le compte de l'utilisateur connecté, ainsi que ses contacts, utilisateurs bloqués et refresh tokens.

**Réponse 200 :**

```json
{ "deleted": true }
```

**Erreurs :**

- `401 Unauthorized` : cookie absent ou token invalide/expiré
- `404 Not Found` : utilisateur non trouvé

---

## Flow détaillé

### Consultation du profil

```
Client GET /profile/me (cookie ACCESS_TOKEN envoyé automatiquement)
  │
  ▼
JwtAuthGuard → vérifie le JWT (sinon 401)
  │
  ▼
ProfileController.getOwnProfile()
  │
  ▼
ProfileService.getOwnProfile(userId)
  └── findUnique avec select (exclut passwordHash)
  │
  ▼
Client reçoit le profil complet
```

### Recherche d'utilisateurs

```
Client GET /profile/search?username=mon&page=1&limit=10
  │
  ▼
JwtAuthGuard → vérifie le JWT (sinon 401)
  │
  ▼
ProfileController.searchUsers()
  ├── Valide que username est présent (sinon 400)
  └── Parse page/limit avec bornes (1-100)
  │
  ▼
ProfileService.searchByUsername()
  ├── findMany (contains, insensible à la casse) + count en parallèle
  └── Retourne { data, total, page, limit, totalPages }
  │
  ▼
Client reçoit les résultats paginés
```

### Mise à jour du profil

```
Client PATCH /profile/me { username: "nouveau" }
  │
  ▼
JwtAuthGuard → vérifie le JWT (sinon 401)
  │
  ▼
ValidationPipe → valide le body via UpdateProfileDto
  │
  ▼
ProfileService.updateProfile(userId, dto)
  ├── Vérifie que le user existe (sinon 404)
  ├── Met à jour via Prisma (try/catch P2002)
  └── Si username déjà pris → 409 Conflict
  │
  ▼
Client reçoit le profil mis à jour
```

### Suppression de compte

```
Client DELETE /profile/me
  │
  ▼
JwtAuthGuard → vérifie le JWT (sinon 401)
  │
  ▼
ProfileService.deleteAccount(userId)
  ├── Vérifie que le user existe (sinon 404)
  └── Transaction Prisma :
      ├── Supprime les contacts (userId OU contactUserId)
      ├── Supprime les blocked users (userId OU blockedUserId)
      └── Supprime le user (cascade: refresh tokens)
  │
  ▼
Client reçoit { deleted: true }
```

---

## Tester avec curl

```bash
# Prérequis : être connecté (cookies.txt contient ACCESS_TOKEN)

# 1. Mon profil
curl -b cookies.txt http://localhost:3000/profile/me

# 2. Rechercher un utilisateur
curl -b cookies.txt "http://localhost:3000/profile/search?username=mon&page=1&limit=10"

# 3. Profil public d'un autre utilisateur
curl -b cookies.txt http://localhost:3000/profile/550e8400-e29b-41d4-a716-446655440000

# 4. Mettre à jour mon profil
curl -b cookies.txt -X PATCH http://localhost:3000/profile/me \
  -H "Content-Type: application/json" \
  -d '{"username":"nouveau_username","about":"Ma nouvelle bio"}'

# 5. Supprimer mon compte
curl -b cookies.txt -X DELETE http://localhost:3000/profile/me
```

> `-b cookies.txt` envoie les cookies sauvegardés lors du login.
