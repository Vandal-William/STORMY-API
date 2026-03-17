-- Cassandra Message Service - Database Initialization
-- Utilisation: POD=$(kubectl get pods -l app=cassandra -o jsonpath='{.items[0].metadata.name}') && kubectl exec -it $POD -- cqlsh localhost < cassandra-init.cql
-- Supprimer le keyspace s'il existe
DROP KEYSPACE IF EXISTS message_service;

-- Créer le keyspace
CREATE KEYSPACE IF NOT EXISTS message_service
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
}
AND durable_writes = true;

USE message_service;

-- Table conversations
CREATE TABLE IF NOT EXISTS conversations (
  id UUID PRIMARY KEY,
  conversation_type TEXT,
  name TEXT,
  description TEXT,
  avatar_url TEXT,
  created_by UUID,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Table conversation_members
CREATE TABLE IF NOT EXISTS conversation_members (
  id UUID PRIMARY KEY,
  conversation_id UUID,
  user_id UUID,
  role TEXT,
  is_muted BOOLEAN,
  joined_at TIMESTAMP,
  left_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_members_conversation_id 
  ON conversation_members(conversation_id);

-- Table messages
CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY,
  conversation_id UUID,
  sender_id UUID,
  content TEXT,
  message_type TEXT,
  reply_to_id UUID,
  is_forwarded BOOLEAN,
  is_edited BOOLEAN,
  is_deleted BOOLEAN,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id 
  ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_messages_sender_id 
  ON messages(sender_id);

CREATE INDEX IF NOT EXISTS idx_messages_created_at 
  ON messages(created_at);

-- Table message_attachments
CREATE TABLE IF NOT EXISTS message_attachments (
  id UUID PRIMARY KEY,
  message_id UUID,
  file_url TEXT,
  file_name TEXT,
  file_type TEXT,
  file_size INT,
  thumbnail_url TEXT
);

CREATE INDEX IF NOT EXISTS idx_message_attachments_message_id 
  ON message_attachments(message_id);

-- Table message_status
CREATE TABLE IF NOT EXISTS message_status (
  id UUID PRIMARY KEY,
  message_id UUID,
  user_id UUID,
  status TEXT,
  delivered_at TIMESTAMP,
  read_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_message_status_message_id 
  ON message_status(message_id);

CREATE INDEX IF NOT EXISTS idx_message_status_user_id 
  ON message_status(user_id);

-- Table user_conversations
CREATE TABLE IF NOT EXISTS user_conversations (
  user_id UUID,
  conversation_id UUID,
  conversation_type TEXT,
  last_activity TIMESTAMP,
  PRIMARY KEY ((user_id), last_activity, conversation_id)
) WITH CLUSTERING ORDER BY (last_activity DESC);

-- Table conversation_messages
CREATE TABLE IF NOT EXISTS conversation_messages (
  conversation_id UUID,
  created_at TIMESTAMP,
  id UUID,
  sender_id UUID,
  content TEXT,
  message_type TEXT,
  PRIMARY KEY ((conversation_id), created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC);

-- ============================================
-- INSERTION DES DONNÉES DE TEST
-- ============================================

-- Conversations
INSERT INTO conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440001, 'direct', 'Alice & Bob Direct', toTimestamp(now()), toTimestamp(now()));

INSERT INTO conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440001, 'group', 'Project Discussion', toTimestamp(now()), toTimestamp(now()));

INSERT INTO conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440003, 550e8400-e29b-41d4-a716-446655440002, 'group', 'Random Chat', toTimestamp(now()), toTimestamp(now()));

-- Conversation Members
INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440100, 650e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440001, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440101, 650e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440002, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440102, 650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440001, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440103, 650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440002, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440104, 650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440003, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440105, 650e8400-e29b-41d4-a716-446655440003, 550e8400-e29b-41d4-a716-446655440002, 'member', toTimestamp(now()));

INSERT INTO conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440106, 650e8400-e29b-41d4-a716-446655440003, 550e8400-e29b-41d4-a716-446655440003, 'member', toTimestamp(now()));

-- Messages
INSERT INTO messages (id, conversation_id, sender_id, content, message_type, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440001, 650e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440001, 'Hey Bob, how are you?', 'text', toTimestamp(now()), toTimestamp(now()));

INSERT INTO messages (id, conversation_id, sender_id, content, message_type, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440002, 650e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440002, 'Hi Alice! I am doing great, thanks for asking!', 'text', toTimestamp(now()), toTimestamp(now()));

INSERT INTO messages (id, conversation_id, sender_id, content, message_type, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440003, 650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440001, 'Team, please review the latest design specs', 'text', toTimestamp(now()), toTimestamp(now()));

INSERT INTO messages (id, conversation_id, sender_id, content, message_type, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440004, 650e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440002, 'Looks good! I will update the frontend accordingly', 'text', toTimestamp(now()), toTimestamp(now()));

INSERT INTO messages (id, conversation_id, sender_id, content, message_type, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440005, 650e8400-e29b-41d4-a716-446655440003, 550e8400-e29b-41d4-a716-446655440003, 'Anyone up for lunch later?', 'text', toTimestamp(now()), toTimestamp(now()));

-- Conversation Messages (dénormalisé pour pagination rapide)
INSERT INTO conversation_messages (conversation_id, created_at, id, sender_id, content, message_type)
VALUES (650e8400-e29b-41d4-a716-446655440001, toTimestamp(now()), 750e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440001, 'Hey Bob, how are you?', 'text');

INSERT INTO conversation_messages (conversation_id, created_at, id, sender_id, content, message_type)
VALUES (650e8400-e29b-41d4-a716-446655440001, toTimestamp(now()), 750e8400-e29b-41d4-a716-446655440002, 550e8400-e29b-41d4-a716-446655440002, 'Hi Alice! I am doing great, thanks for asking!', 'text');

INSERT INTO conversation_messages (conversation_id, created_at, id, sender_id, content, message_type)
VALUES (650e8400-e29b-41d4-a716-446655440002, toTimestamp(now()), 750e8400-e29b-41d4-a716-446655440003, 550e8400-e29b-41d4-a716-446655440001, 'Team, please review the latest design specs', 'text');

INSERT INTO conversation_messages (conversation_id, created_at, id, sender_id, content, message_type)
VALUES (650e8400-e29b-41d4-a716-446655440002, toTimestamp(now()), 750e8400-e29b-41d4-a716-446655440004, 550e8400-e29b-41d4-a716-446655440002, 'Looks good! I will update the frontend accordingly', 'text');

INSERT INTO conversation_messages (conversation_id, created_at, id, sender_id, content, message_type)
VALUES (650e8400-e29b-41d4-a716-446655440003, toTimestamp(now()), 750e8400-e29b-41d4-a716-446655440005, 550e8400-e29b-41d4-a716-446655440003, 'Anyone up for lunch later?', 'text');

-- User Conversations
INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440001, 650e8400-e29b-41d4-a716-446655440001, 'direct', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440001, 650e8400-e29b-41d4-a716-446655440002, 'group', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440002, 650e8400-e29b-41d4-a716-446655440001, 'direct', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440002, 650e8400-e29b-41d4-a716-446655440002, 'group', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440002, 650e8400-e29b-41d4-a716-446655440003, 'group', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440003, 650e8400-e29b-41d4-a716-446655440002, 'group', toTimestamp(now()));

INSERT INTO user_conversations (user_id, conversation_id, conversation_type, last_activity)
VALUES (550e8400-e29b-41d4-a716-446655440003, 650e8400-e29b-41d4-a716-446655440003, 'group', toTimestamp(now()));
