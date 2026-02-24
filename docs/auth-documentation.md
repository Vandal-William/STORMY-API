# Documentation - Authentification (Service User)

## Vue d'ensemble

L'authentification est gérée par le **service user** (NestJS). Elle utilise un système **double token** (access + refresh) stockés dans des **cookies httpOnly** :

- **ACCESS_TOKEN** : JWT courte durée (15 min) pour accéder aux routes protégées
- **REFRESH_TOKEN** : token opaque longue durée (7 jours) pour renouveler l'access token

---

## Architecture des fichiers

| Fichier | Rôle |
|---------|------|
| `auth.module.ts` | Configure le module (Passport, JWT 15min, providers) |
| `auth.service.ts` | Logique métier (register, login, refreshAccessToken, logout, getProfile) |
| `auth.controller.ts` | Routes HTTP + gestion des 2 cookies + rate limiting |
| `jwt.strategy.ts` | Stratégie Passport : extrait le JWT depuis le cookie |
| `jwt-auth.guard.ts` | Guard qui protège les routes authentifiées |
| `dto/register.dto.ts` | Validation des données d'inscription |
| `dto/login.dto.ts` | Validation des données de connexion |

---

## Routes disponibles

### `POST /auth/register`

Crée un nouvel utilisateur et pose les 2 cookies (access + refresh).

**Rate limit : 3 requêtes / minute**

**Body :**
```json
{
  "phone": "0612345678",
  "username": "monuser",
  "password": "monmotdepasse",
  "email": "optionnel@mail.com"
}
```

**Réponse 201 :**
```json
{ "message": "registered" }
```

**Headers Set-Cookie :**
```
ACCESS_TOKEN=eyJ...; Max-Age=900; Path=/; HttpOnly; SameSite=Lax
REFRESH_TOKEN=5d262a...; Max-Age=604800; Path=/; HttpOnly; SameSite=Lax
```

**Erreurs :**
- `409 Conflict` : "Phone number already taken" / "Username already taken"
- `429 Too Many Requests` : rate limit dépassé

---

### `POST /auth/login`

Authentifie un utilisateur existant et pose les 2 cookies.

**Rate limit : 5 requêtes / minute**

**Body :**
```json
{
  "username": "monuser",
  "password": "monmotdepasse"
}
```

**Réponse 201 :**
```json
{ "message": "logged in" }
```

**Erreurs :**
- `401 Unauthorized` : identifiants invalides
- `429 Too Many Requests` : rate limit dépassé

---

### `POST /auth/refresh`

Renouvelle l'access token en utilisant le refresh token. Appelé quand l'access token expire (15 min).

**Rate limit : 60 requêtes / minute (global)**

**Prérequis :** cookie `REFRESH_TOKEN` présent

**Réponse 201 :**
```json
{ "message": "token refreshed" }
```

**Header Set-Cookie :**
```
ACCESS_TOKEN=eyJ...; Max-Age=900; Path=/; HttpOnly; SameSite=Lax
```

**Erreurs :**
- `401 Unauthorized` : refresh token absent, invalide ou expiré

---

### `GET /auth/me` (protégée)

Retourne le profil de l'utilisateur connecté. Nécessite le cookie `ACCESS_TOKEN` valide.

**Réponse 200 :**
```json
{
  "id": "uuid",
  "phone": "0612345678",
  "username": "monuser",
  "email": null,
  "avatarUrl": null,
  "about": null,
  "lastSeen": null,
  "createdAt": "2026-02-24T08:34:54.297Z"
}
```

**Erreurs :**
- `401 Unauthorized` : cookie absent ou token invalide/expiré

---

### `POST /auth/logout`

Supprime les 2 cookies et invalide le refresh token en base de données.

**Réponse 201 :**
```json
{ "message": "logged out" }
```

---

## Flow détaillé

### Inscription

```
Client POST /auth/register { phone, username, password }
  │
  ▼
Rate Limiter → vérifie 3 req/min (sinon 429)
  │
  ▼
Controller → valide le body via RegisterDto (ValidationPipe)
  │
  ▼
AuthService.register()
  ├── Hash le mot de passe avec bcrypt (salt: 10)
  ├── Crée le user en base via Prisma (try/catch P2002 si doublon)
  ├── Génère un access token JWT (15 min)
  └── Génère un refresh token opaque + stocke en base (7 jours)
  │
  ▼
Controller → pose ACCESS_TOKEN (15 min) + REFRESH_TOKEN (7 jours)
  │
  ▼
Client reçoit { message: "registered" } + 2 cookies
```

### Connexion

```
Client POST /auth/login { username, password }
  │
  ▼
Rate Limiter → vérifie 5 req/min (sinon 429)
  │
  ▼
AuthService.login()
  ├── Cherche le user par username (sinon 401)
  ├── Compare le password avec bcrypt (sinon 401)
  ├── Génère un access token JWT (15 min)
  └── Génère un refresh token opaque + stocke en base (7 jours)
  │
  ▼
Controller → pose ACCESS_TOKEN (15 min) + REFRESH_TOKEN (7 jours)
  │
  ▼
Client reçoit { message: "logged in" } + 2 cookies
```

### Renouvellement du token

```
Client POST /auth/refresh (cookie REFRESH_TOKEN envoyé automatiquement)
  │
  ▼
Controller → extrait REFRESH_TOKEN du cookie (sinon 401)
  │
  ▼
AuthService.refreshAccessToken()
  ├── Cherche le refresh token en base (sinon 401)
  ├── Vérifie qu'il n'est pas expiré (sinon supprime + 401)
  └── Génère un nouveau access token JWT (15 min)
  │
  ▼
Controller → repose ACCESS_TOKEN (15 min)
  │
  ▼
Client reçoit { message: "token refreshed" } + nouveau cookie ACCESS_TOKEN
```

### Accès route protégée (GET /auth/me)

```
Client GET /auth/me (cookie ACCESS_TOKEN envoyé automatiquement)
  │
  ▼
JwtAuthGuard intercepte la requête
  │
  ▼
JwtStrategy
  ├── Extrait le JWT depuis le cookie ACCESS_TOKEN
  ├── Vérifie la signature avec JWT_SECRET
  ├── Vérifie l'expiration (15 min)
  └── Retourne { userId, username } dans req.user
  │
  ▼
Controller → appelle AuthService.getProfile(userId)
  │
  ▼
AuthService → récupère le user en base (sans le passwordHash)
  │
  ▼
Client reçoit le profil utilisateur
```

### Déconnexion

```
Client POST /auth/logout
  │
  ▼
Controller → extrait REFRESH_TOKEN du cookie
  │
  ▼
AuthService.logout()
  └── Supprime le refresh token en base de données
  │
  ▼
Controller → efface ACCESS_TOKEN + REFRESH_TOKEN (Expires: 1970)
  │
  ▼
Client reçoit { message: "logged out" }
Le refresh token est invalidé : impossible de regénérer un access token
```

---

## Sécurité

| Mesure | Détail |
|--------|--------|
| **httpOnly** | Les 2 cookies ne sont pas accessibles via JavaScript (protection XSS) |
| **SameSite: Lax** | Protection contre les attaques CSRF |
| **bcrypt (salt: 10)** | Hash du mot de passe coûteux en calcul |
| **Access token 15 min** | Fenêtre d'exposition réduite si le token est compromis |
| **Refresh token 7 jours** | Stocké en base, révocable au logout |
| **Rate limiting** | Register: 3/min, Login: 5/min, Global: 60/min (anti brute force) |
| **Race condition** | Contrainte UNIQUE en base + catch Prisma P2002 (pas de doublon possible) |
| **passwordHash exclu** | Le hash n'est jamais renvoyé dans les réponses |
| **ValidationPipe** | Whitelist active : les champs inconnus sont ignorés |
| **secure: false** | En dev uniquement — a passer a `true` en production (HTTPS) |

---

## Schema base de donnees

### Table `refresh_tokens`

| Colonne | Type | Description |
|---------|------|-------------|
| `id` | UUID | Clé primaire |
| `token` | VARCHAR(500) | Token opaque unique (128 hex chars) |
| `user_id` | UUID | FK vers `users.id` (CASCADE on delete) |
| `expires_at` | DATETIME | Date d'expiration (J+7) |
| `created_at` | DATETIME | Date de création |

Index : `user_id`, `expires_at`

---

## Tester avec curl

```bash
# 1. Register (pose ACCESS_TOKEN + REFRESH_TOKEN)
curl -c cookies.txt -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"0612345678","username":"monuser","password":"motdepasse8"}'

# 2. Login
curl -c cookies.txt -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"monuser","password":"motdepasse8"}'

# 3. Profil (route protégée)
curl -b cookies.txt http://localhost:3000/auth/me

# 4. Refresh (renouvelle ACCESS_TOKEN via REFRESH_TOKEN)
curl -b cookies.txt -c cookies.txt -X POST http://localhost:3000/auth/refresh

# 5. Logout (supprime les 2 cookies + invalide le refresh token en base)
curl -b cookies.txt -c cookies.txt -X POST http://localhost:3000/auth/logout

# 6. Vérifier le rate limiting (le 6eme doit renvoyer 429)
for i in $(seq 1 6); do curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"wrong"}'; echo; done
```

> `-c cookies.txt` sauvegarde les cookies, `-b cookies.txt` les renvoie.

---

## Variables d'environnement requises

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | Clé secrète pour signer/vérifier les JWT |
| `DATABASE_URL` | URL de connexion PostgreSQL |
| `PORT` | Port du service (défaut: 3000) |
