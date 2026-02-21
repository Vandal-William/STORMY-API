# Cassandra Architecture Documentation

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     Message Service                      │
│  ┌───────────────────────────────────────────────────┐   │
│  │                      HTTP API                      │   │
│  │  GET /messages, POST /messages, PUT /messages...   │   │
│  └─────────────────────────┬─────────────────────────┘   │
│                            │                              │
│  ┌─────────────────────────▼─────────────────────────┐   │
│  │                    Handlers                        │   │
│  │  MessageHandler   ConversationHandler             │   │
│  └─────────────────────────┬─────────────────────────┘   │
│                            │                              │
│  ┌─────────────────────────▼─────────────────────────┐   │
│  │                  Services Layer                    │   │
│  │  MessageService   ConversationService             │   │
│  │  (Business Logic, Validation, Transactions)       │   │
│  └─────────────────────────┬─────────────────────────┘   │
│                            │                              │
│  ┌─────────────────────────▼─────────────────────────┐   │
│  │               Repository Interface                 │   │
│  │  MessageRepository    ConversationRepository       │   │
│  └────────┬──────────────────────────────────┬────────┘   │
│           │                                  │             │
│  ┌────────▼────────┐              ┌────────▼────────┐   │
│  │  In-Memory Repo │              │  Cassandra Repo  │   │
│  │  (Development)  │              │  (Production)    │   │
│  └────────┬────────┘              └────────┬────────┘   │
│           │                                │              │
│           └────────────┬───────────────────┘              │
│                        │                                  │
└────────────────────────┼──────────────────────────────────┘
                         │
             ┌───────────▼───────────┐
             │   Apache Cassandra    │
             │   Distributed DB      │
             └───────────────────────┘
```

## Data Model - ER Diagram

```
┌─────────────────────┐      ┌──────────────────────────┐
│   conversations     │      │  users (external ref)    │
├─────────────────────┤      ├──────────────────────────┤
│ id (PK)             │      │ id (uuid)                │
│ type (enum)         │      │ name                     │
│ name                │      │ email                    │
│ description         │      └──────────────────────────┘
│ avatar_url          │             ▲
│ created_by (FK)────────────────────┘
│ created_at          │
│ updated_at          │
└─────────────────────┘
         │
         │ 1:N
         │
    ┌────┴──────────────────────────┐
    │                               │
    │                    ┌─────────────────────────────┐
    │                    │  conversation_members       │
    │                    ├─────────────────────────────┤
    │                    │ id (PK)                     │
    │                    │ conversation_id (FK)────┐   │
    │                    │ user_id (FK)            │   │
    │                    │ role (enum)             │   │
    │                    │ is_muted (boolean)      │   │
    │                    │ joined_at               │   │
    │                    │ left_at (soft delete)   │   │
    │                    └─────────────────────────┘   │
    │                                                   │
    │
    │ 1:N
    │
    ├─────────────────────────────┐
    │                             │
┌───┴──────────────┐    ┌────────▼────────────┐
│    messages      │    │  message_status     │
├──────────────────┤    ├─────────────────────┤
│ id (PK)          │    │ id (PK)             │
│ conversation_id  │    │ message_id (FK)─┐   │
│ sender_id (FK)   │    │ user_id (FK)    │   │
│ content          │    │ status (enum)   │   │
│ type (enum)      │    │ delivered_at    │   │
│ reply_to_id(FK)──┼──┐ │ read_at         │   │
│ is_forwarded     │  │ └─────────────────┘   │
│ is_edited        │  │                       │
│ is_deleted       │  │ ┌──────────────────┐  │
│ created_at       │  │ │ message_attach.  │  │
│ updated_at       │  │ ├──────────────────┤  │
└──────────────────┘  │ │ id (PK)          │  │
         │            │ │ message_id (FK)──┼──┘
         │            │ │ file_url         │
         │            │ │ file_name        │
         │            │ │ file_type        │
         │            │ │ file_size        │
         │            │ │ thumbnail_url    │
         │            │ └──────────────────┘
         │            │
         └────────────┘
```

## Denormalization Strategy

### Why Denormalize?

Cassandra is a **write-optimized** database best suited for:
- Time-series-like data (messages ordered by time)
- Fast pagination of large datasets
- Avoiding expensive JOIN operations

### Denormalized Tables

#### 1. `conversation_messages` (Clustering Index)
**Purpose**: Fast retrieval of messages in a conversation

```
Partition Key: conversation_id
Clustering Keys: created_at (DESC), id

SELECT * FROM conversation_messages 
WHERE conversation_id = ? 
ORDER BY created_at DESC 
LIMIT 50;
```

**Benefits**:
- Single query for all messages in conversation
- Built-in pagination via clustering
- Handles thousands of messages efficiently

#### 2. `user_conversations` (User Index)
**Purpose**: Quick discovery of user's conversations

```
Partition Key: user_id
Clustering Keys: last_activity (DESC), conversation_id

SELECT conversation_id, conversation_type 
FROM user_conversations 
WHERE user_id = ? 
ORDER BY last_activity DESC;
```

**Benefits**:
- User can quickly see most recent conversations
- Natural sorting by activity
- Avoids full table scan of conversations

## Query Patterns

### 1. Create Message
```go
// 1. Insert into primary table
INSERT INTO messages (...) VALUES (...)

// 2. Insert into denormalized table (same data)
INSERT INTO conversation_messages (...) VALUES (...)

// 3. Create delivery status records
INSERT INTO message_status (...) VALUES (...)  // Per recipient
```

**Time Complexity**: O(1)
**Consistency**: Strong (same partition)

### 2. Get Messages in Conversation
```go
SELECT * FROM conversation_messages
WHERE conversation_id = ?
ORDER BY created_at DESC
LIMIT ? ALLOW FILTERING;
```

**Time Complexity**: O(n) where n = limit
**Consistency**: Local reads (same node)

### 3. Get User's Conversations
```go
SELECT * FROM user_conversations
WHERE user_id = ?
ORDER BY last_activity DESC
LIMIT ?;
```

**Time Complexity**: O(m) where m = limit
**Consistency**: Local reads

### 4. Update Message
```go
UPDATE messages 
SET content = ?, is_edited = true, updated_at = now
WHERE id = ?;

// Note: Does NOT automatically update conversation_messages
// Would need separate update or treat as denormalized write
```

### 5. Delete Message (Soft Delete)
```go
UPDATE messages 
SET is_deleted = true, updated_at = now
WHERE id = ?;
```

**Why Soft Delete?**
- Maintains referential integrity (for replies)
- Enables audit trail
- Easier to recover from accidents

## Consistency Patterns

### Strong Consistency
```go
// Single partition update (strongly consistent)
UPDATE messages 
SET content = ? 
WHERE id = ? AND conversation_id = ?;
```

### Eventual Consistency (Denormalization)
```go
// Separate table update (eventually consistent)
UPDATE conversation_messages 
SET content = ? 
WHERE conversation_id = ? AND created_at = ? AND id = ?;
```

**Note**: With proper TTLs and eventual consistency, this is acceptable for messaging apps where old data being temporarily stale is OK.

## Scalability Patterns

### 1. Partition Key Design

✅ **Good**: `conversation_id` as partition
- Evenly distributed (varies by messaging pattern)
- Most queries access single partition
- Supports clustering by time

❌ **Bad**: `user_id` as partition for messages
- Hot partition (active users = many writes)
- Queries across users would need scatter-gather

### 2. Token Ring Distribution

With proper replication_factor:
```
replication_factor = 3  // Production

Tokens distributed: 2^63 = 9.2×10^18 token range
Each node owns: 2^63 / num_nodes
```

### 3. Write Amplification

```
Single message write:
  messages table → 1 disk write
  conversation_messages (denormalized) → 1 disk write
  message_status (per recipient) → N disk writes
  
Total: N+2 writes
```

**Optimizable with**:
- Batch writes (atomic at partition level)
- Async status updates via NATS

## Scalability Limits

| Metric | Limit | Mitigation |
|--------|-------|-----------|
| Rows per partition | ~20 million | Good - messages per conversation rarely exceed |
| Blob size | 1 MB | Use references for large files |
| Column count | 10,000+ | Good - schema has ~10 columns |
| Cluster size | 1000+ nodes | Designed for large clusters |
| Write throughput | 50k-100k ops/sec/node | Scale horizontally |
| Read throughput | 10k-30k ops/sec/node | Add read replicas |

## Performance Characteristics

### Write Performance

|Operation|Latency|Throughput|
|----------|-------|----------|
|Insert 1 message|5-10ms|10k-50k msgs/sec|
|Batch insert (10)|8-15ms|100k-500k msgs/sec|
|Update (soft delete)|3-5ms|20k-100k ops/sec|

### Read Performance

|Query Type|Latency|Notes|
|-----------|-------|-----|
|Get message by ID|2-5ms|Single row lookup|
|Messages in conversation|5-20ms|50 rows, sorted by time|
|User's conversations|3-8ms|Recent activity index|
|Search messages|50-500ms|Requires full scan (bad)|

## Backup & Recovery Strategy

### 1. Snapshot-Based Backups
```bash
# Create snapshot
nodetool snapshot message_service -t backup_$(date +%s)

# Located in: /var/lib/cassandra/snapshots/
```

### 2. Write-Ahead Logs
```
Cassandra automatically maintains WAL for crash recovery
Located in: /var/lib/cassandra/commitlog
```

### 3. Cross-Region Replication
```yaml
# In schema
CREATE KEYSPACE message_service WITH replication = {
  'class': 'NetworkTopologyStrategy',
  'dc1': 3,
  'dc2': 2
}
```

## Monitoring & Alerting

### Key Metrics to Monitor

```
1. Write Latency
   - p50 (median): <10ms
   - p99: <50ms

2. Read Latency
   - p50: <5ms
   - p99: <20ms

3. JVM GC Pause
   - G1GC: <100ms
   
4. Compaction State
   - Queue size: <5
   - Pending: <1000

5. Memtable Size
   - Heap usage: <30%
```

### Health Checks

```bash
# Cluster topology
nodetool status

# Node info
nodetool info

# Pending compactions
nodetool compactionstats

# Keyspace stats
nodetool tablestats message_service
```

## Cost Optimization

### For AWS Cloud

```
Option 1: Cassandra on EC2 (Self-managed)
- Cost: $0.05-0.10 per hour per node
- Flexibility: High
- Maintenance: Operator responsible

Option 2: Astra DB (Managed)
- Cost: $0.11 per million write units
- Flexibility: Medium
- Maintenance: DataStax managed

Option 3: ElastiCache (Cluster Mode)
- Cost: Better for caching layer
- Not suitable as primary DB
```

### Cost per Message

```
Assuming 1M messages/day:
- Storage: ~10GB/month = ~$1/month
- I/O: 3-5 writes per message = $15-30/month
- Total: ~$25-50/month for 1M msgs/day
```

## Future Enhancements

### 1. Materialized Views (Deprecated, but useful)
```cql
-- Not recommended in Cassandra 4.0+, use denormalization instead
```

### 2. Secondary Indexes (Use carefully)
```cql
CREATE INDEX ON messages(sender_id);  -- Can cause hot spots
```

### 3. Search Integration
```
Solr Index:
- Full-text search on message content
- Deployed as index within Cassandra
```

### 4. Time-Window Compaction
```cql
CREATE TABLE messages_by_day (
  date date,
  created_at timestamp,
  -- ...
) WITH CLUSTERING ORDER BY (created_at DESC)
  AND compaction = { 'class': 'TimeWindowCompactionStrategy' };
```

## References

- [Cassandra Data Modeling Best Practices](https://cassandra.apache.org/doc/latest/cassandra/data_modeling/intro.html)
- [gocql Driver](https://github.com/gocql/gocql)
- [Partition Key Design](https://cassandra.apache.org/doc/latest/cassandra/data_modeling/conceptual.html)
