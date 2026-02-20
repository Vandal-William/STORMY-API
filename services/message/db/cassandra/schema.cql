-- Keyspace pour le service de messages
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
  created_by INT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Table conversation_members
CREATE TABLE IF NOT EXISTS conversation_members (
  id UUID PRIMARY KEY,
  conversation_id UUID,
  user_id INT,
  role TEXT,
  is_muted BOOLEAN,
  joined_at TIMESTAMP,
  left_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_members_user_id 
  ON conversation_members(user_id);

CREATE INDEX IF NOT EXISTS idx_conversation_members_conversation_id 
  ON conversation_members(conversation_id);

-- Table messages
CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY,
  conversation_id UUID,
  sender_id INT,
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
  user_id INT,
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
  user_id INT,
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
