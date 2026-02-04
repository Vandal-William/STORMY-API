# Configuration Kubernetes - Améliorations

## Résumé des changements

Votre configuration Kubernetes a été entièrement revue et optimisée pour la production. Voici les principales améliorations apportées:

### 1. **Health Checks (Liveness & Readiness Probes)**
- Ajoutés pour tous les services
- Permettent à Kubernetes de détecter les pods défaillants et de les redémarrer automatiquement
- Assurent que le trafic n'est envoyé que vers les pods prêts

### 2. **Gestion des ressources**
- **Requests**: Garantissent les ressources minimales
  - Services légers (gateway, presence): 100m CPU, 128-256Mi RAM
  - Services de données (postgres, cassandra): 250m-500m CPU, 512Mi-1Gi RAM

- **Limits**: Préventent les débordements de ressources
  - Postgres: Limité à 1Gi RAM, 1000m CPU
  - Cassandra: Limité à 2Gi RAM, 2000m CPU

### 3. **Ports en conflit - RÉSOLUS**
- **Moderation**: 3000 → **3004** (évite conflict avec user-service)
- **Notification**: 3000 → **3003** (évite conflicts)
- Gateway communicates maintenant correctement avec les bons ports

### 4. **Secrets Kubernetes (au lieu de plaintext)**
- `jwt-secret`: Secret pour user-service
- `postgres-secret`: Secret pour Postgres
- `grafana-secret`: Secret pour Grafana
- **⚠️ À changer en production!**

### 5. **Volumes persistants**
- **Postgres**: Sauvegarde des données avec `/var/lib/postgresql/data`
- **Redis**: Persistence mode (`--appendonly yes`)
- **Cassandra**: Stockage `/var/lib/cassandra`
- Utilise `emptyDir {}` pour développement (changez en `PersistentVolumeClaim` pour prod)

### 6. **Réplication & Haute disponibilité**
- `gateway`: 2 replicas
- `user-service`: 2 replicas
- `message-service`: 2 replicas
- `presence-service`: 2 replicas
- Services de base de données: 1 replica (pour dev)

### 7. **ImagePullPolicy**
- Défini à `IfNotPresent` pour tous les services
- Évite les pulls inutiles depuis le registry

### 8. **Labels et namespaces**
- Tous les ressources avec labels appropriés
- Namespace: `default` (changez en `production` pour prod)

### 9. **Variables d'environnement complètes**
- Ports Redis et Cassandra explicitement définis
- Configuration NATS JetStream avec ressources appropriées
- Configuration Grafana avec plugins

### 10. **Ports nommés**
- Chaque port a un nom (`http`, `redis`, `postgres`, etc.)
- Facilite le monitoring et le troubleshooting

## Configuration NATS - Nouveau**
```yaml
jetstream {
  store_dir: /data
  max_memory: 128M
  max_file: 256M
}
```
- Port monitoring: 8222 (accessible pour metrics)

## Déploiement

```bash
# Déployer tous les services
kubectl apply -f k8s/

# Vérifier le déploiement
kubectl get pods
kubectl get services
kubectl describe pod <pod-name>

# Logs
kubectl logs <pod-name>

# Port-forwarding local
kubectl port-forward svc/gateway 8080:8080
```

## ⚠️ IMPORTANT - À faire avant production

1. **Changer les secrets**:
   ```bash
   kubectl create secret generic jwt-secret --from-literal=secret=YOUR-SECURE-SECRET
   kubectl create secret generic postgres-secret --from-literal=password=YOUR-DB-PASSWORD
   kubectl create secret generic grafana-secret --from-literal=password=YOUR-GRAFANA-PASSWORD
   ```

2. **Utiliser PersistentVolumes**:
   - Remplacer `emptyDir {}` par des `PersistentVolumeClaim`
   - Configurer le stockage approprié

3. **Namespace**:
   - Créer un namespace `production`
   - Utiliser des resource quotas

4. **Image Registry**:
   - Utilisez votre registry privé au lieu de `dev`
   - Configurez `imagePullSecrets`

5. **Network Policies**:
   - Ajouter des policies de firewall K8s
   - Restreindre le trafic inter-pods

6. **Ingress**:
   - Ajouter une Ingress pour exposer gateway
   - Remplacer `NodePort` par `ClusterIP`

7. **Monitoring**:
   - Ajouter Prometheus
   - Configurer les AlertRules

## Vérification de la configuration

```bash
# Valider les YAML
kubectl apply -f k8s/ --dry-run=client

# Voir les ressources
kubectl get all
kubectl get secrets
kubectl get configmaps
```

Votre infrastructure Kubernetes est maintenant prête pour fonctionner correctement! 🚀
