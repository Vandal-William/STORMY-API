#!/bin/bash

# Script to initialize Cassandra schema for message service

kubectl exec cassandra-d7d6cc445-xtg82 -- cqlsh << 'CQL'
CREATE KEYSPACE IF NOT EXISTS message_service
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
AND durable_writes = true;

USE message_service;

CREATE TABLE IF NOT EXISTS conversations (
  id UUID PRIMARY KEY,
  conversation_type TEXT NOT NULL,
  name TEXT,
  description TEXT,
  avatar_url TEXT,
  created_by UUID NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS conversation_members (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL,
  user_id UUID NOT NULL,
  role TEXT NOT NULL DEFAULT 'member',
  is_muted BOOLEAN NOT NULL DEFAULT false,
  joined_at TIMESTAMP NOT NULL,
  left_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_members_user_id ON conversation_members(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_members_conversation_id ON conversation_members(conversation_id);

CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL,
  sender_id UUID NOT NULL,
  content TEXT,
  message_type TEXT NOT NULL DEFAULT 'text',
  reply_to_id UUID,
  is_forwarded BOOLEAN NOT NULL DEFAULT false,
  is_edited BOOLEAN NOT NULL DEFAULT false,
  is_deleted BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

CREATE TABLE IF NOT EXISTS message_attachments (
  id UUID PRIMARY KEY,
  message_id UUID NOT NULL,
  file_url TEXT NOT NULL,
  file_name TEXT,
  file_type TEXT,
  file_size INT,
  thumbnail_url TEXT
);

CREATE INDEX IF NOT EXISTS idx_message_attachments_message_id ON message_attachments(message_id);

CREATE TABLE IF NOT EXISTS message_status (
  id UUID PRIMARY KEY,
  message_id UUID NOT NULL,
  user_id UUID NOT NULL,
  status TEXT NOT NULL DEFAULT 'sent',
  delivered_at TIMESTAMP,
  read_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_message_status_message_id ON message_status(message_id);
CREATE INDEX IF NOT EXISTS idx_message_status_user_id ON message_status(user_id);

CREATE TABLE IF NOT EXISTS user_conversations (
  user_id UUID,
  conversation_id UUID,
  conversation_type TEXT,
  last_activity TIMESTAMP,
  PRIMARY KEY ((user_id), last_activity, conversation_id)
) WITH CLUSTERING ORDER BY (last_activity DESC);

CREATE TABLE IF NOT EXISTS conversation_messages (
  conversation_id UUID,
  created_at TIMESTAMP,
  id UUID,
  sender_id UUID,
  content TEXT,
  message_type TEXT,
  PRIMARY KEY ((conversation_id), created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC);
CQL

echo "Schema initialization complete!"
