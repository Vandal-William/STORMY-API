# 🚀 Gateway Scalable - Vue d'ensemble

## ✨ Structure du projet

```
gateway/
├── cmd/api/main.go
├── config/services.yaml
└── internal/
    ├── config/config.go
    ├── middleware/
    │   ├── auth.go
    │   └── logger.go
    ├── proxy/
    │   └── universal.go
    ├── registry/
    │   └── service_registry.go
    └── router/
        └── router.go
```

---

## 📝 Qu'est-ce que chaque fichier?

### **cmd/api/main.go**
Point d'entrée. Charge la config, crée l'engine Gin, setup les routes, lance le serveur.
```bash
go run ./cmd/api/main.go
```

### **config/services.yaml** ⭐
Configuration centralisée de TOUS les services. Ajouter une API = ajouter 4 lignes ici.
```yaml
services:
  - name: messages
    url: http://message-service:3001
    prefix: /messages
    auth_required: true
```

### **internal/config/config.go**
Charge le YAML et les variables d'environnement. Crée la ServiceRegistry. Expose `Config` et `Load()`.

### **internal/middleware/auth.go**
Valide les JWT depuis les cookies. Extrait l'user ID et le met dans le contexte pour les routes protégées.

### **internal/middleware/logger.go**
Log chaque requête HTTP avec la méthode, le chemin, l'IP client et la latence.

### **internal/proxy/universal.go** ✨ NOUVEAU
Proxy universel unique! Clone copies: headers, cookies, body, query params. Forward vers le service cible. Gère tous les verbes HTTP.

### **internal/registry/service_registry.go** ✨ NOUVEAU
Registre dynamique thread-safe des services. `Register()`, `FindByPrefix()`, `FindByName()`, `GetAll()`. Lookup O(1).

### **internal/router/router.go**
Génère AUTOMATIQUEMENT les routes pour CHAQUE service. Applique JWT middleware sélectivement. Setup handler pour `/info`.

---

## 🎯 Flux d'une requête

```
Client Request
    ↓
[Router] Middleware (Logging)
    ↓
[Router] Match /{service}/*
    ↓
[Middleware] JWT (si auth_required=true)
    ↓
[UniversalProxy] Clone tout + Forward
    ↓
Service Cible
    ↓
Response → Clone tout → Client
```

---

## ⚡ Démarrer en 3 étapes

### 1️⃣ Configurer services.yaml
```yaml
services:
  - name: mon-api
    url: http://mon-api:3000
    prefix: /mon-api
    auth_required: false
```

### 2️⃣ Compiler
```bash
go mod tidy
go build ./cmd/api/main.go
```

### 3️⃣ Lancer
```bash
./main
```

---

## 🔑 Points clés

| Concept | C'est quoi? |
|---------|-----------|
| **UniversalProxy** | Proxy unique qui gère TOUS les services. Zéro duplication. |
| **ServiceRegistry** | Registre des services. Lookup rapide par prefix/nom. |
| **Routes dynamiques** | Toutes les routes sont générées automatiquement depuis la config. |
| **auth_required** | Contrôle si JWT middleware s'applique ou pas. |
| **config/services.yaml** | Seul endroit où ajouter des services! Ajouter une API = 4 lignes. |

---

## 🧪 Tester

```bash
# Health check + liste des services
curl http://localhost:8080/info

# Tester une requête (ajustez l'API)
curl http://localhost:8080/mon-api/resource

# Avec authentification
curl -b cookies.txt http://localhost:8080/protected/resource
```

---

## 🎓 Comment ajouter une nouvelle API?

1. Ouvrir `config/services.yaml`
2. Ajouter ces 4 lignes:
```yaml
  - name: api-nom
    url: http://api-nom:port
    prefix: /api-nom
    auth_required: false
```
3. Redémarrer la gateway
4. C'EST TOUT! ✨

Routes créées automatiquement:
- `GET /api-nom/*`
- `POST /api-nom/*`
- `PUT /api-nom/*`
- `DELETE /api-nom/*`
- `PATCH /api-nom/*`

---

## ⚙️ Variables d'environnement

```bash
PORT=8080              # Port du serveur
HOST=0.0.0.0          # Interface d'écoute
JWT_SECRET=votre-clé  # Clé JWT pour valider tokens
```

---

## ✅ Résumé

- ✨ **Proxy universel** - Gère tous les services sans duplication
- 🔧 **Registry dynamique** - Lookup O(1), thread-safe
- 📝 **Routes auto** - Générées depuis la config
- 🚀 **Scalable** - Ajouter une API = 4 lignes YAML
- 🔒 **Sécurisé** - JWT middleware intégré
- 📦 **Propre** - Code SOLID, bien documenté

**Bon développement! 🎉**
