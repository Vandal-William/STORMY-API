#!/bin/bash
kubectl exec cassandra-d7d6cc445-xtg82 -- cqlsh -e "
CREATE KEYSPACE IF NOT EXISTS msg_service WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1} AND durable_writes = true;
USE msg_service;
CREATE TABLE IF NOT EXISTS conversations (id UUID PRIMARY KEY, ctype TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS conversation_members (id UUID PRIMARY KEY, cid UUID NOT NULL, uid UUID NOT NULL);
CREATE TABLE IF NOT EXISTS messages (id UUID PRIMARY KEY, cid UUID NOT NULL, sid UUID NOT NULL);
CREATE TABLE IF NOT EXISTS message_attachments (id UUID PRIMARY KEY, mid UUID NOT NULL);
CREATE TABLE IF NOT EXISTS message_status (id UUID PRIMARY KEY, mid UUID NOT NULL, uid UUID NOT NULL);
CREATE TABLE IF NOT EXISTS user_conversations (uid UUID, cid UUID, PRIMARY KEY ((uid), cid));
CREATE TABLE IF NOT EXISTS conversation_messages (cid UUID, created_at TIMESTAMP, mid UUID, PRIMARY KEY ((cid), created_at, mid));
"
