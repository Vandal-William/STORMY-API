# Cassandra Migration Guide

## Migration Path from In-Memory to Cassandra

This document outlines the migration strategy from InMemory repositories to Cassandra.

## Phase 1: Setup (Pre-Migration)

### 1. Install Tools
```bash
# Install gocql driver
go get github.com/gocql/gocql

# Install Cassandra (local dev)
brew install cassandra              # macOS
sudo apt-get install cassandra      # Linux
docker pull cassandra:4.1           # Docker
```

### 2. Start Cassandra
```bash
# Option A: Local
cassandra

# Option B: Docker
docker run -d -p 9042:9042 cassandra:4.1

# Option C: Docker Compose (recommended)
make db-up      # Starts Cassandra, Redis, NATS
```

### 3. Initialize Schema
```bash
# Method 1: Python script (recommended)
python3 db/cassandra/setup.py localhost 9042

# Method 2: cqlsh
cqlsh -f db/cassandra/schema.cql

# Method 3: Bash script
bash db/cassandra/setup.sh localhost 9042

# Method 4: Make target
make db-init
```

### 4. Verify Schema
```bash
# Connect to Cassandra
cqlsh

# List keyspaces
DESCRIBE KEYSPACES;

# Verify tables
USE message_service;
DESCRIBE TABLES;

# Check table schema
DESCRIBE TABLE messages;
```

## Phase 2: Dual-Write Implementation (Safe Transition)

### Step 1: Create Cassandra Repositories

**Already done**: See `internal/repository/cassandra_message.go` and `cassandra_conversation.go`

### Step 2: Update DI Container

```go
// cmd/api/main.go

// Create both repositories
inMemRepo := repository.NewInMemoryMessageRepository()
cassandraRepo := repository.NewCassandraMessageRepository(cassandraClient)

// Start with in-memory in main.go
// Later, switch to Cassandra
messageService := service.NewMessageService(inMemRepo)
```

### Step 3: Implement Gradual Migration Strategy

Option A: Feature Flag Approach
```go
// internal/service/message.go
func (s *MessageService) Create(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
    if os.Getenv("USE_CASSANDRA") == "true" {
        return s.cassandraRepo.Create(ctx, msg)
    }
    return s.inMemoryRepo.Create(ctx, msg)
}
```

Option B: Dual-Write Approach
```go
func (s *MessageService) Create(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
    // Write to both
    inMemResult, inMemErr := s.inMemoryRepo.Create(ctx, msg)
    cassandraResult, cassandraErr := s.cassandraRepo.Create(ctx, msg)
    
    // Log any divergence
    if inMemErr != nil || cassandraErr != nil {
        logger.Warn("Dual-write divergence", "inMemErr", inMemErr, "cassandraErr", cassandraErr)
    }
    
    // Return primary
    if inMemErr == nil {
        return inMemResult, nil
    }
    return cassandraResult, cassandraErr
}
```

Option C: Read from Cassandra, Write to Both
```go
func (s *MessageService) GetByID(ctx context.Context, id gocql.UUID) (*domain.Message, error) {
    // Read from Cassandra (primary)
    return s.cassandraRepo.GetByID(ctx, id)
}

func (s *MessageService) Create(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
    // Write to Cassandra first
    if err := s.cassandraRepo.Create(ctx, msg); err != nil {
        return nil, err
    }
    
    // Write to in-memory for warmup (can fail)
    _ = s.inMemoryRepo.Create(ctx, msg)
    
    return msg, nil
}
```

## Phase 3: Data Migration (if starting with existing data)

### Scenario: Migrate from SQL Database

```bash
# 1. Export data from SQL
mysqldump -u user -p database messages > messages.sql

# 2. Transform to CSV
# Create script: db/cassandra/transform.py

# 3. Load into Cassandra
cqlsh -e "COPY message_service.messages FROM 'messages.csv' WITH HEADER=true;"
```

### Cassandra Bulk Loading Script

```python
# db/cassandra/migrate_data.py
#!/usr/bin/env python3

import csv
from cassandra.cluster import Cluster
from cassandra.util import uuid_from_time
import sys

def migrate_messages(csv_file, contact_points=['localhost']):
    cluster = Cluster(contact_points=contact_points)
    session = cluster.connect('message_service')
    
    prepared = session.prepare(
        """INSERT INTO messages 
        (id, conversation_id, sender_id, content, type, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)"""
    )
    
    with open(csv_file, 'r') as f:
        reader = csv.DictReader(f)
        for i, row in enumerate(reader):
            session.execute(prepared, [
                uuid.UUID(row['id']),
                uuid.UUID(row['conversation_id']),
                uuid.UUID(row['sender_id']),
                row['content'],
                row['type'],
                datetime.fromisoformat(row['created_at']),
                datetime.fromisoformat(row['updated_at']),
            ])
            
            if (i + 1) % 1000 == 0:
                print(f"Migrated {i + 1} messages")
    
    cluster.shutdown()
    print("Migration completed!")

if __name__ == "__main__":
    migrate_messages(sys.argv[1] if len(sys.argv) > 1 else "messages.csv")
```

## Phase 4: Cutover (Go Live)

### Prerequisites Checklist
- [ ] Cassandra cluster is stable and healthy
- [ ] Schema is initialized
- [ ] Data migration completed (if applicable)
- [ ] Dual-write tests passed
- [ ] Performance tests show acceptable latency
- [ ] Backup strategy tested
- [ ] Rollback plan documented

### Cutover Steps

1. **Enable Cassandra in Code**
   ```go
   // cmd/api/main.go
   cassandraClient, _ := cassandra.NewClient(cfg.Cassandra)
   messageRepo := repository.NewCassandraMessageRepository(cassandraClient)
   conversationRepo := repository.NewCassandraConversationRepository(cassandraClient)
   ```

2. **Deploy to Staging**
   ```bash
   git checkout cassandra-migration
   docker build -t message-service:cassandra .
   kubectl apply -f k8s/message-service.yaml
   ```

3. **Run Tests**
   ```bash
   make test-integration
   make test-load    # Optional load tests
   ```

4. **Monitor Metrics**
   - Query latency (target: <10ms p99)
   - Error rate (target: <0.1%)
   - Storage growth
   - CPU/Memory usage

5. **Canary Deployment** (if using K8s)
   ```bash
   # Deploy to 10% of pods
   kubectl set image deployment/message-service \
     message-service=message-service:cassandra \
     --record \
     --max-surge=1 \
     --max-unavailable=0
   ```

6. **Full Rollout**
   ```bash
   # After 1 hour monitoring, rollout to rest
   kubectl rollout status deployment/message-service
   ```

7. **Remove In-Memory Fallback**
   ```go
   // After 1-2 weeks of successful Cassandra operation
   // Remove inMemoryRepo completely
   // Only use CassandraRepository
   ```

## Phase 5: Optimization (Post-Migration)

### 1. Index Optimization
```bash
# Enable query logging to find missing indexes
nodetool statusbinary
```

### 2. Compaction Tuning
```yaml
# cassandra.yaml
compaction_throughput_mb_per_sec: 64  # Adjust based on hardware
```

### 3. Read Repair
```go
// In gocql connection
cluster.ReadRepair = gocql.HealDigests
```

### 4. Token Awareness
```go
cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(
    gocql.RoundRobinHostPolicy(),
)
```

## Rollback Plan

If issues occur during migration:

### Option 1: Failed Cassandra Connection
```go
// Keep fallback in code temporarily
func (s *MessageService) Create(ctx context.Context, msg *domain.Message) {
    err := s.cassandraRepo.Create(ctx, msg)
    if err != nil {
        logger.Error("Cassandra failed, using fallback", "error", err)
        return s.inMemoryRepo.Create(ctx, msg)
    }
    return err
}
```

### Option 2: Immediate Rollback
```bash
# Revert deployment
kubectl rollout undo deployment/message-service
git checkout main

# Redeploy with in-memory
docker build -t message-service:main .
docker push message-service:main
kubectl set image deployment/message-service message-service=message-service:main
```

### Option 3: Infrastructure Failure
```bash
# Restore Cassandra from snapshot
nodetool restore -from_timestamp <snapshot_id> <table_name>

# Or restore from backup
./scripts/restore_cassandra.sh backup_2024_02_20
```

## Monitoring & Validation

### Automated Tests
```bash
# Before each deployment
make test-integration

# Performance baseline
make test-load

# Schema validation
cqlsh -e "DESCRIBE KEYSPACE message_service;"
```

### Health Checks
```go
// Add health endpoint that checks Cassandra
GET /health/cassandra

# Returns:
{
  "status": "ok",
  "cassandra": {
    "connected": true,
    "latency_ms": 2.5,
    "keyspace": "message_service"
  }
}
```

### Dashboard Queries

Prometheus queries for monitoring:
```
# Message write latency (p99)
histogram_quantile(0.99, rate(message_write_latency_bucket[5m]))

# Error rate
rate(message_errors_total[5m])

# Cassandra connection pool size
cassandra_connection_pool_size

# Message count (cardinality)
count(messages)
```

## Troubleshooting Common Issues

### Issue 1: Connection Timeouts
```
Problem: "failed to connect to cassandra: timed out"

Solutions:
1. Check Cassandra is running: docker ps | grep cassandra
2. Verify host/port: CASSANDRA_HOSTS=localhost CASSANDRA_PORT=9042
3. Increase timeouts: CASSANDRA_CONNECT_TIMEOUT=30
4. Check firewall: telnet localhost 9042
```

### Issue 2: Schema Mismatch
```
Problem: "table does not exist" errors

Solutions:
1. Verify keyspace: cqlsh -e "DESCRIBE KEYSPACES"
2. Re-initialize: make db-init
3. Check env: CASSANDRA_KEYSPACE=message_service
```

### Issue 3: Data Not Persisting
```
Problem: Data disappears after restart

Solutions:
1. Check volumes: docker volume ls | grep cassandra
2. Verify permissions: chmod -R 655 /var/lib/cassandra
3. Check disk space: df -h /var/lib/cassandra
```

### Issue 4: High Latency
```
Problem: Queries taking >100ms

Solutions:
1. Check load: nodetool status
2. Monitor GC: watch -n 1 "jps -l | grep cassandra"
3. Adjust JVM heap: MAX_HEAP_SIZE=4G
```

## Success Criteria

✅ **Migration is successful when:**
- [ ] All tests pass
- [ ] Query latency: p50 <5ms, p99 <20ms
- [ ] Error rate: <0.01%
- [ ] Storage verified: ~1KB per message
- [ ] Backup/restore tested
- [ ] No data loss
- [ ] Monitoring alerts working
- [ ] Documentation updated

## Timeline Example

```
Week 1:
  - Day 1: Setup Cassandra infrastructure
  - Day 2: Initialize schema
  - Day 3: Implement dual-write
  - Day 4: Run integration tests

Week 2:
  - Day 1: Canary deployment (10%)
  - Day 2-3: Monitor metrics
  - Day 4: Full rollout (if stable)

Week 3:
  - Remove in-memory fallback
  - Archive old test data
  - Update documentation
```

## References

- [Cassandra Documentation](https://cassandra.apache.org/doc/latest/)
- [Data Migration Best Practices](https://cassandra.apache.org/doc/latest/cassandra/managing/tools/cqlsh.html)
- [Zero-Downtime Deployments](https://martinfowler.com/bliki/BlueGreenDeployment.html)
