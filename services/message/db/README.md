# Message Service - Cassandra Database Setup

## Architecture Overview

La service message utilise **Apache Cassandra** comme base de données avec une architecture modulaire et scalable. Le schéma est conçu pour :

- **Haute disponibilité** : Duplication des données pour la résilience
- **Performance** : Tables dénormalisées pour les requêtes courantes
- **Conformité RGPD** : Soft deletes avec timestamps pour l'audit
- **Evolved design** : Support des pièces jointes, statuts de messages, conversations de groupe

## Schema Overview

### Core Tables

#### `conversations`
Stocke les métadonnées des conversations (privées ou groupes).

```
id (UUID PK)
type (enum: 'private', 'group')
name (varchar) - Null pour conversations privées
description (text)
avatar_url (varchar)
created_by (UUID) - Référence vers users.id
created_at (timestamp)
updated_at (timestamp)
```

#### `conversation_members`
Membres d'une conversation avec rôles et permissions.

```
id (UUID PK)
conversation_id (UUID) - Référence
user_id (UUID) - Référence
role (enum: 'owner', 'admin', 'member')
is_muted (boolean)
joined_at (timestamp)
left_at (timestamp, nullable) - Pour soft delete
```

#### `messages`
Messages dans les conversations.

```
id (UUID PK)
conversation_id (UUID) - Référence
sender_id (UUID) - Référence
content (text)
type (enum: 'text', 'image', 'video', 'audio', 'document', 'location', 'contact', 'system')
reply_to_id (UUID, nullable) - Référence récursive
is_forwarded (boolean)
is_edited (boolean)
is_deleted (boolean) - Soft delete
created_at (timestamp)
updated_at (timestamp)
```

#### `message_attachments`
Pièces jointes des messages (images, vidéos, documents).

```
id (UUID PK)
message_id (UUID) - Référence
file_url (varchar)
file_name (varchar)
file_type (varchar)
file_size (int)
thumbnail_url (varchar, nullable)
```

#### `message_status`
Statut de livraison et de lecture par destinataire.

```
id (UUID PK)
message_id (UUID) - Référence
user_id (UUID) - Référence
status (enum: 'sent', 'delivered', 'read')
delivered_at (timestamp, nullable)
read_at (timestamp, nullable)
```

### Denormalized Tables (for Performance)

#### `conversation_messages`
Index dénormalisé pour la pagination rapide des messages d'une conversation.

```
conversation_id (UUID) - Clustering key 1
created_at (timestamp DESC) - Clustering key 2
id (UUID) - Clustering key 3
sender_id (UUID)
content (text)
type (varchar)
```

#### `user_conversations`
Index dénormalisé pour trouver rapidement les conversations d'un utilisateur.

```
user_id (UUID) - Partition key
last_activity (timestamp DESC) - Clustering key 1
conversation_id (UUID) - Clustering key 2
conversation_type (varchar)
```

## Installation

### Option 1: Docker Compose (Recommended for Development)

```bash
# Lancer tous les services (Cassandra + Redis + NATS + Message Service)
docker-compose up -d

# Vérifier que Cassandra est prêt (attend ~40 secondes)
docker-compose logs cassandra

# Initialiser le schéma
docker-compose exec cassandra cqlsh -f /db/cassandra/schema.cql
```

### Option 2: Local Cassandra

#### Installation
```bash
# macOS
brew install cassandra

# Ubuntu/Debian
sudo apt-get install cassandra cassandra-tools

# Windows
# Télécharger et installer depuis https://archive.apache.org/dist/cassandra/
```

#### Démarrage
```bash
# Linux/macOS
cassandra

# Ou en tant que service
sudo service cassandra start
```

#### Initialisation du schéma
```bash
# Avec cqlsh directement
cqlsh -f db/cassandra/schema.cql

# Ou avec le script Python
python3 db/cassandra/setup.py localhost 9042

# Ou avec le script bash
bash db/cassandra/setup.sh localhost 9042
```

### Option 3: Kubernetes (Production)

```bash
# Installer Cassandra via Helm (cassandra-operator chart)
helm repo add cassandra https://cassandra-operator.io/
helm repo update
helm install cassandra cassandra/cassandra -f k8s/cassandra-values.yaml

# Vérifier les pods
kubectl get pods -l app=cassandra

# Initialiser le schéma
kubectl exec -it cassandra-0 -- cqlsh -f /db/cassandra/schema.cql

# Ou en tant que job
kubectl apply -f k8s/cassandra-init-job.yaml
```

## Configuration

### Environment Variables

```bash
# Cassandra
CASSANDRA_HOSTS=localhost              # Point de contact
CASSANDRA_PORT=9042                    # Port Cassandra
CASSANDRA_KEYSPACE=message_service     # Keyspace à utiliser
CASSANDRA_USERNAME=                    # Authentification (optionnel)
CASSANDRA_PASSWORD=                    # Authentification (optionnel)
CASSANDRA_CONSISTENCY=LOCAL_ONE        # Niveau de cohérence
CASSANDRA_TIMEOUT=10                   # Timeout (secondes)
CASSANDRA_CONNECT_TIMEOUT=10           # Timeout de connexion (secondes)

# Redis (pour caching)
REDIS_HOST=localhost
REDIS_PORT=6379

# NATS (pour événements)
NATS_URL=nats://localhost:4222
```

### Consistency Levels

- **ONE** : Plus rapide, moins de copies
- **LOCAL_ONE** : Recommandé, cohérence locale avec réplication
- **QUORUM** : Plus sûr, attend majorité
- **LOCAL_QUORUM** : Production, cohérence locale avec quorum

Pour production : `LOCAL_QUORUM` avec `replication_factor=3`

## Repository Implementation

### MessageRepository

L'interface `MessageRepository` supporte deux implémentations :

#### 1. InMemoryMessageRepository (Development)
```go
repo := repository.NewInMemoryMessageRepository()
```

#### 2. CassandraMessageRepository (Production)
```go
client, _ := cassandra.NewClient(cfg.Cassandra)
repo := repository.NewCassandraMessageRepository(client)
```

### ConversationRepository

Similairement :

```go
// Development
convRepo := repository.NewInMemoryConversationRepository()

// Production
convRepo := repository.NewCassandraConversationRepository(client)
```

## API Endpoints

### Messages

```
POST   /messages                    # Créer un message
GET    /messages/:id                # Récupérer un message
GET    /messages/conversation/:id   # Messages d'une conversation
GET    /messages/user/:user_id      # Messages d'un utilisateur
PUT    /messages/:id                # Modifier un message
DELETE /messages/:id                # Supprimer un message
```

### Conversations

```
POST   /conversations               # Créer une conversation
GET    /conversations/:id           # Récupérer une conversation
GET    /conversations/user/:user_id # Conversations d'un utilisateur
PUT    /conversations/:id           # Modifier une conversation
DELETE /conversations/:id           # Supprimer une conversation
```

### Conversation Members

```
POST   /conversations/:id/members   # Ajouter un membre
DELETE /conversations/:id/members/:user_id  # Retirer un membre
GET    /conversations/:id/members   # Lister les membres
```

## Scalability Features

### 1. Partitioning

Les tables sont partitionnées intelligemment :
- `messages` : Par `conversation_id` pour les recherches rapides
- `conversation_members` : Par `conversation_id` pour la cohérence
- `message_status` : Par `message_id` pour les performances

### 2. Clustering

Clustering Order (DESC) pour les requêtes récentes :
```
conversation_messages:
  PARTITION BY conversation_id
  CLUSTER BY created_at DESC, id

user_conversations:
  PARTITION BY user_id
  CLUSTER BY last_activity DESC, conversation_id
```

### 3. Denormalization

Tables dénormalisées (`conversation_messages`, `user_conversations`) pour :
- Éviter les JOIN coûteux
- Pagination rapide
- Requêtes à faible latence

### 4. TTL (Time To Live)

Optional pour message temporaires :
```go
query := session.Query(
    "INSERT INTO messages (...) USING TTL 3600",  // 1 heure
    ...
)
```

## Monitoring & Operations

### Health Check

```bash
# Via cqlsh
cqlsh -e "SELECT now() FROM system.local LIMIT 1"

# Via API
curl http://localhost:3001/info
```

### Backup

```bash
# Créer un backup
nodetool snapshot cassandra message_service

# Restaurer depuis backup
nodetool restore
```

### Compaction

```bash
# Forcer une compaction
nodetool compact message_service
```

### Ring Status

```bash
nodetool status message_service
```

## Best Practices

### 1. Query Patterns

✅ **Recommandé :**
```go
// Récupérer messages d'une conversation
"SELECT * FROM conversation_messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT ?"

// Récupérer conversations d'un utilisateur
"SELECT * FROM user_conversations WHERE user_id = ? ORDER BY last_activity DESC LIMIT ?"
```

❌ **À éviter :**
```go
// Requête sans clé de partition
"SELECT * FROM messages WHERE sender_id = ?"

// Requête sans limite
"SELECT * FROM messages WHERE conversation_id = ?"
```

### 2. Batch Operations

```go
batch := session.NewBatch(gocql.LoggedBatch)
batch.Query("INSERT INTO messages (...)", ...)
batch.Query("INSERT INTO conversation_messages (...)", ...)
session.ExecuteBatch(batch)
```

### 3. Connection Pooling

```go
cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(
    gocql.RoundRobinHostPolicy(),
)
```

## Testing

### Unit Tests (In-Memory)

```bash
go test ./internal/repository -v
```

### Integration Tests (Cassandra)

```bash
# Lancer Cassandra
docker-compose up cassandra

# Attendre que ce soit prêt
sleep 40

# Exécuter les tests
go test ./tests/integration -v
```

### Load Testing

```bash
# Avec k6
k6 run tests/load/messages.js

# Avec Apache JMeter
jmeter -n -t tests/load/cassandra.jmx
```

## Troubleshooting

### Connection Refused

```bash
# Vérifier que Cassandra écoute
netstat -an | grep 9042

# Vérifier les logs
docker logs cassandra-dev
```

### Timeout on Connections

```bash
# Augmenter les timeouts dans .env
CASSANDRA_TIMEOUT=30
CASSANDRA_CONNECT_TIMEOUT=30

# Vérifier les ressources
docker stats cassandra-dev
```

### Schema Not Found

```bash
# Vérifier le keyspace
cqlsh -e "DESCRIBE KEYSPACES"

# Réinitialiser le schéma
cqlsh -f db/cassandra/schema.cql
```

## Performance Tuning

### JVM Settings (for Cassandra)

```bash
# Augmenter la heap size
export MAX_HEAP_SIZE="4G"
export HEAP_NEWSIZE="1G"
cassandra
```

### Cassandra Tuning

```yaml
# cassandra.yaml
num_tokens: 256
concurrent_reads: 32
concurrent_writes: 32
concurrent_counter_writes: 32
disk_optimization_strategy: ssd
```

### Go Client Tuning

```go
cluster.PoolConfig.MaxConnsPerHost = 10
cluster.PoolConfig.MinConnsPerHost = 2
cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
```

## Version Compatibility

- Cassandra 4.0+
- Go 1.23.0+
- gocql v1.6.0+

## Next Steps

1. ✅ Schéma Cassandra défini
2. ✅ Repositories implémentés
3. ⏳ Tester avec données réelles
4. ⏳ Implémenter le caching Redis
5. ⏳ Ajouter les événements NATS
6. ⏳ Setup de réplication multi-region

## Support & Links

- [Cassandra Documentation](https://cassandra.apache.org/doc/latest/)
- [gocql Documentation](https://github.com/gocql/gocql)
- [CQL Query Language](https://cassandra.apache.org/doc/latest/cassandra/cql/index.html)
