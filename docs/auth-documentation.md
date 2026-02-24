# Documentation - Authentification (Service User)

## Vue d'ensemble

L'authentification est gérée par le **service user** (NestJS). Elle utilise **JWT stocké dans un cookie httpOnly** pour sécuriser les sessions utilisateurs.

---

## Architecture des fichiers

| Fichier | Rôle |
|---------|------|
| `auth.module.ts` | Configure le module (Passport, JWT, providers) |
| `auth.service.ts` | Logique métier (register, login, getProfile, generateToken) |
| `auth.controller.ts` | Routes HTTP + pose/supprime le cookie |
| `jwt.strategy.ts` | Stratégie Passport : extrait le JWT depuis le cookie |
| `jwt-auth.guard.ts` | Guard qui protège les routes authentifiées |
| `dto/register.dto.ts` | Validation des données d'inscription |
| `dto/login.dto.ts` | Validation des données de connexion |

---

## Routes disponibles

### `POST /auth/register`

Crée un nouvel utilisateur et pose le cookie JWT.

**Body :**
```json
{
  "phone": "0612345678",   // min 6 caractères
  "username": "monuser",   // min 3 caractères
  "password": "monmotdepasse", // min 8 caractères
  "email": "optionnel@mail.com" // optionnel
}
```

**Réponse 201 :**
```json
{ "message": "registered" }
```

**Header Set-Cookie :**
```
ACCESS_TOKEN=eyJ...; Max-Age=86400; Path=/; HttpOnly; SameSite=Lax
```

**Erreurs :**
- `409 Conflict` : phone ou username déjà pris

---

### `POST /auth/login`

Authentifie un utilisateur existant et pose le cookie JWT.

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

---

### `GET /auth/me` (protégée)

Retourne le profil de l'utilisateur connecté. Nécessite le cookie `ACCESS_TOKEN`.

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

Supprime le cookie `ACCESS_TOKEN`.

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
Controller → valide le body via RegisterDto (ValidationPipe)
  │
  ▼
AuthService.register()
  ├── Vérifie que phone/username ne sont pas pris (sinon 409)
  ├── Hash le mot de passe avec bcrypt (salt: 10)
  ├── Crée le user en base via Prisma
  └── Génère un JWT { sub: userId, username }
  │
  ▼
Controller → pose le cookie ACCESS_TOKEN (httpOnly, 24h)
  │
  ▼
Client reçoit { message: "registered" } + cookie
```

### Connexion

```
Client POST /auth/login { username, password }
  │
  ▼
AuthService.login()
  ├── Cherche le user par username (sinon 401)
  ├── Compare le password avec bcrypt (sinon 401)
  └── Génère un JWT { sub: userId, username }
  │
  ▼
Controller → pose le cookie ACCESS_TOKEN (httpOnly, 24h)
  │
  ▼
Client reçoit { message: "logged in" } + cookie
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
  ├── Vérifie l'expiration
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
Controller → efface le cookie ACCESS_TOKEN (Expires: 1970)
  │
  ▼
Client reçoit { message: "logged out" }
Les prochaines requêtes n'auront plus de token valide
```

---

## Sécurité

| Mesure | Détail |
|--------|--------|
| **httpOnly** | Le cookie n'est pas accessible via JavaScript (protection XSS) |
| **SameSite: Lax** | Protection contre les attaques CSRF |
| **bcrypt (salt: 10)** | Hash du mot de passe coûteux en calcul |
| **Expiration 24h** | Le JWT expire après 24 heures |
| **passwordHash exclu** | Le hash n'est jamais renvoyé dans les réponses |
| **ValidationPipe** | Whitelist active : les champs inconnus sont ignorés |
| **secure: false** | En dev uniquement — à passer à `true` en production (HTTPS) |

---

## Tester avec curl

```bash
# 1. Register
curl -c cookies.txt -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"0612345678","username":"monuser","password":"motdepasse8"}'

# 2. Login
curl -c cookies.txt -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"monuser","password":"motdepasse8"}'

# 3. Profil (route protégée)
curl -b cookies.txt http://localhost:3000/auth/me

# 4. Logout
curl -b cookies.txt -c cookies.txt -X POST http://localhost:3000/auth/logout
```

> `-c cookies.txt` sauvegarde le cookie, `-b cookies.txt` le renvoie.

---

## Variables d'environnement requises

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | Clé secrète pour signer/vérifier les JWT |
| `DATABASE_URL` | URL de connexion PostgreSQL |
| `PORT` | Port du service (défaut: 3000) |
