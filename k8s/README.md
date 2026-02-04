# Kubernetes Deployment Guide

## Architecture

```
                    ┌─────────────────┐
                    │    GATEWAY      │
                    │   (port 8080)   │
                    │   NodePort:30080│
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ user-service  │   │message-service│   │presence-service│
│  (port 3000)  │   │  (port 3001)  │   │  (port 3002)   │
│    NestJS     │   │      Go       │   │      Go        │
└───────┬───────┘   └───────┬───────┘   └───────┬────────┘
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   postgres    │   │   cassandra   │   │  redis-user   │
│  (port 5432)  │   │  (port 9042)  │   │  (port 6379)  │
└───────────────┘   └───────────────┘   └───────────────┘

        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│notification-  │   │moderation-    │   │     nats      │
│   service     │   │   service     │   │  (port 4222)  │
│ (port 3003)   │   │  (port 3004)  │   │   Message Bus │
│    NestJS     │   │    NestJS     │   └───────────────┘
└───────────────┘   └───────────────┘
```

## Prérequis

- Docker Desktop
- Minikube
- kubectl

### Installation (macOS)

```bash
# Installer Minikube
brew install minikube

# Installer kubectl (si pas déjà installé)
brew install kubectl
```

### Installation (Windows)

```powershell
# Avec Chocolatey
choco install minikube
choco install kubernetes-cli

# Ou télécharger depuis :
# https://minikube.sigs.k8s.io/docs/start/
# https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/
```

## Démarrage rapide

### 1. Démarrer Minikube

```bash
minikube start
```

### 2. Configurer Docker pour utiliser Minikube

```bash
# macOS/Linux
eval $(minikube docker-env)

# Windows PowerShell
& minikube -p minikube docker-env --shell powershell | Invoke-Expression

# Windows CMD
@FOR /f "tokens=*" %i IN ('minikube -p minikube docker-env --shell cmd') DO @%i
```

### 3. Builder les images Docker

```bash
# Depuis la racine du projet
docker build -t gateway:dev ./services/gateway
docker build -t user-service:dev ./services/user
docker build -t message-service:dev ./services/message
docker build -t presence-service:dev ./services/presence
docker build -t notification-service:dev ./services/notification
docker build -t moderation-service:dev ./services/moderation
```

### 4. Déployer sur Kubernetes

```bash
kubectl apply -f k8s/
```

### 5. Vérifier le déploiement

```bash
# Voir tous les pods
kubectl get pods

# Voir tous les services
kubectl get svc

# Voir les logs d'un service
kubectl logs -f deployment/gateway
```

### 6. Accéder à l'application

```bash
# Option 1 : Port-forward
kubectl port-forward svc/gateway 8080:8080
# Puis accéder à http://localhost:8080/info

# Option 2 : URL Minikube
minikube service gateway --url
```

## Structure des fichiers K8s

| Fichier               | Description                               |
| --------------------- | ----------------------------------------- |
| `gateway.yaml`      | API Gateway (Go) - Point d'entrée unique |
| `user.yaml`         | Service utilisateur (NestJS)              |
| `message.yaml`      | Service messages (Go)                     |
| `presence.yaml`     | Service présence (Go)                    |
| `notification.yaml` | Service notifications (NestJS)            |
| `moderation.yaml`   | Service modération (NestJS)              |
| `postgres.yaml`     | Base de données PostgreSQL               |
| `cassandra.yaml`    | Base de données Cassandra                |
| `redis.yaml`        | Redis pour cache/sessions                 |
| `nats.yaml`         | Message broker NATS                       |
| `grafana.yaml`      | Monitoring avec Grafana                   |

## Services et Ports

| Service              | Port interne | Type             | Technologie   |
| -------------------- | ------------ | ---------------- | ------------- |
| gateway              | 8080         | NodePort (30080) | Go/Gin        |
| user-service         | 3000         | ClusterIP        | NestJS        |
| message-service      | 3001         | ClusterIP        | Go/Gin        |
| presence-service     | 3002         | ClusterIP        | Go/Gin        |
| notification-service | 3003         | ClusterIP        | NestJS        |
| moderation-service   | 3004         | ClusterIP        | NestJS        |
| postgres             | 5432         | ClusterIP        | PostgreSQL 16 |
| cassandra            | 9042         | ClusterIP        | Cassandra 4.1 |
| redis-user           | 6379         | ClusterIP        | Redis 7       |
| redis-message        | 6379         | ClusterIP        | Redis 7       |
| nats                 | 4222         | ClusterIP        | NATS 2.10     |
| grafana              | 3000         | NodePort (30090) | Grafana       |

## Health Checks

Tous les services exposent un endpoint `/info` pour les health checks :

```bash
curl http://localhost:8080/info
```

Réponse :

```json
{
  "gateway": "ok",
  "message": "200 OK",
  "moderation": "200 OK",
  "notification": "200 OK",
  "presence": "200 OK",
  "user": "200 OK"
}
```

## Commandes utiles

### Gestion des pods

```bash
# Redémarrer un deployment
kubectl rollout restart deployment/gateway

# Scaler un deployment
kubectl scale deployment/user-service --replicas=3

# Voir les événements
kubectl get events --sort-by=.metadata.creationTimestamp
```

### Debug

```bash
# Voir les logs d'un pod spécifique
kubectl logs -f <pod-name>

# Exécuter un shell dans un pod
kubectl exec -it <pod-name> -- /bin/sh

# Décrire un pod (voir les erreurs)
kubectl describe pod <pod-name>
```

### Nettoyage

```bash
# Supprimer tous les déploiements
kubectl delete -f k8s/

# Arrêter Minikube
minikube stop

# Supprimer le cluster Minikube
minikube delete
```

## Troubleshooting

### Erreur "exec format error"

Les images ont été buildées pour une architecture différente. Rebuilder les images :

```bash
eval $(minikube docker-env)
docker build -t <service>:dev ./services/<service>
kubectl rollout restart deployment/<service>
```

### Pod en CrashLoopBackOff

1. Vérifier les logs : `kubectl logs <pod-name>`
2. Décrire le pod : `kubectl describe pod <pod-name>`
3. Vérifier que l'image existe : `docker images | grep <service>`

### Impossible de se connecter au service

1. Vérifier que le pod est Running : `kubectl get pods`
2. Vérifier le service : `kubectl get svc`
3. Utiliser port-forward : `kubectl port-forward svc/<service> <port>:<port>`

## Variables d'environnement

Les services utilisent les variables d'environnement suivantes (définies dans les fichiers YAML) :

| Variable     | Service                           | Description                 |
| ------------ | --------------------------------- | --------------------------- |
| DATABASE_URL | user, notification, moderation    | URL PostgreSQL              |
| REDIS_HOST   | user, message, presence           | Hôte Redis                 |
| NATS_URL     | message, notification, moderation | URL NATS                    |
| JWT_SECRET   | user                              | Secret JWT (via Secret K8s) |

## Secrets

Les secrets sont définis dans les fichiers YAML correspondants :

- `jwt-secret` : Secret JWT pour l'authentification
- `postgres-secret` : Credentials PostgreSQL
- `grafana-secret` : Credentials Grafana admin

**Important** : En production, utilisez des secrets externes (Vault, AWS Secrets Manager, etc.)

## Docker Compose vs Kubernetes

Ce projet peut être exécuté soit avec **Docker Compose** (développement simple) soit avec **Kubernetes/Minikube** (proche de la production).

### Comparaison

| Critère | Docker Compose | Kubernetes (Minikube) |
|---------|----------------|----------------------|
| Complexité | Simple | Plus complexe |
| Cas d'usage | Développement local | Staging/Production |
| Scaling | Manuel | Automatique |
| Health checks | Basique | Avancé (liveness/readiness) |
| Visualisation | Docker Desktop | Minikube Dashboard |

### Vérifier ce qui tourne

```bash
# Voir les conteneurs Docker Compose
docker ps

# Voir les pods Kubernetes
kubectl get pods
```

### URLs d'accès

| Source | URL | Commande pour démarrer |
|--------|-----|------------------------|
| Docker Compose | `http://localhost:8080` | `docker compose up -d` |
| Kubernetes (port-forward) | `http://localhost:9090` | `kubectl port-forward svc/gateway 9090:8080` |
| Kubernetes (NodePort) | `http://<minikube-ip>:30080` | `minikube service gateway --url` |

### Choisir l'environnement

```bash
# Utiliser uniquement Docker Compose
minikube stop
docker compose up -d

# Utiliser uniquement Kubernetes
docker compose down
minikube start
kubectl apply -f k8s/
```

## Minikube Dashboard

Le dashboard Kubernetes permet de visualiser graphiquement l'état du cluster (comme Docker Desktop pour les conteneurs).

### Lancer le dashboard

```bash
# Ouvre automatiquement dans le navigateur
minikube dashboard

# Ou obtenir uniquement l'URL
minikube dashboard --url
```

### Ce que montre le dashboard

| Section | Description |
|---------|-------------|
| **Workloads > Deployments** | État des 6 microservices déployés |
| **Workloads > Pods** | Chaque instance de service (Running, Error, etc.) |
| **Service > Services** | Les endpoints et ports exposés |
| **Config > ConfigMaps** | Variables de configuration |
| **Config > Secrets** | Secrets (JWT, credentials DB) |

### Comment fonctionne le dashboard

1. Le dashboard est un **addon Minikube** pré-installé
2. `minikube dashboard` lance un **proxy sécurisé** entre ta machine et le cluster
3. L'URL générée (ex: `http://127.0.0.1:XXXXX/...`) est temporaire et change à chaque lancement
4. Le port est attribué aléatoirement par Minikube

### Arrêter le dashboard

```bash
# Le dashboard s'arrête quand tu fermes le terminal
# Ou trouver et tuer le processus
ps aux | grep "minikube dashboard"
kill <PID>
```

## Docker Desktop Kubernetes

Alternative à Minikube, Docker Desktop inclut Kubernetes intégré.

### Activer K8s dans Docker Desktop

1. Ouvre **Docker Desktop**
2. Clique sur l'icône **engrenage** (Settings) en haut à droite
3. Dans le menu de gauche, clique sur **Kubernetes**
4. Coche **Enable Kubernetes**
5. Clique sur **Apply & Restart**

### Gérer les contextes K8s

```bash
# Voir les contextes disponibles
kubectl config get-contexts

# Utiliser Docker Desktop
kubectl config use-context docker-desktop

# Utiliser Minikube
kubectl config use-context minikube

# Vérifier le contexte actuel
kubectl config current-context
```

### Déployer sur Docker Desktop K8s

```bash
# S'assurer d'être sur le bon contexte
kubectl config use-context docker-desktop

# Déployer
kubectl apply -f k8s/

# Vérifier
kubectl get pods
```

### Différences Docker Desktop vs Minikube

| Critère | Docker Desktop K8s | Minikube |
|---------|-------------------|----------|
| Installation | Intégré à Docker | Séparé |
| Dashboard | Non inclus (manuel) | Inclus |
| NodePort | `localhost:30080` direct | `minikube service --url` |
| Ressources | Partagées avec Docker | Dédiées |

### URLs d'accès selon l'environnement

| Environnement | URL | Commande |
|---------------|-----|----------|
| Docker Compose | `http://localhost:8080` | `docker compose up -d` |
| Docker Desktop K8s (NodePort) | `http://localhost:30080` | `kubectl apply -f k8s/` |
| Docker Desktop K8s (port-forward) | `http://localhost:9090` | `kubectl port-forward svc/gateway 9090:8080` |
| Minikube (NodePort) | `http://<minikube-ip>:30080` | `minikube service gateway --url` |

### Éviter les conflits de ports

Docker Compose et Kubernetes peuvent avoir des conflits sur le port 8080.

```bash
# Option 1 : Utiliser uniquement Docker Compose
docker compose up -d
# Accès : http://localhost:8080

# Option 2 : Utiliser uniquement Kubernetes
docker compose down
kubectl apply -f k8s/
# Accès : http://localhost:30080

# Option 3 : Les deux en parallèle (ports différents)
# Docker Compose : http://localhost:8080
# Kubernetes : http://localhost:30080 (NodePort)
```

### Installer le Dashboard sur Docker Desktop K8s

Docker Desktop K8s n'inclut pas de dashboard par défaut. Pour l'installer :

```bash
# Installer le dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml

# Lancer le proxy
kubectl proxy

# Accéder au dashboard
# http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
```

**Alternative** : Utiliser [Lens](https://k8slens.dev/) (application gratuite) pour une interface graphique complète.