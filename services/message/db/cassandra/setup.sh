#!/bin/bash

# Script d'initialisation de Cassandra pour le service de messages
# Usage: ./setup.sh [CASSANDRA_HOST] [CASSANDRA_PORT]

CASSANDRA_HOST="${1:-localhost}"
CASSANDRA_PORT="${2:-9042}"

echo "Initializing Cassandra database at $CASSANDRA_HOST:$CASSANDRA_PORT..."

# Attendre que Cassandra soit prêt
echo "Waiting for Cassandra to be ready..."
for i in {1..30}; do
  if cqlsh $CASSANDRA_HOST $CASSANDRA_PORT --version > /dev/null 2>&1; then
    echo "Cassandra is ready!"
    break
  fi
  echo "Attempt $i/30: Waiting for Cassandra..."
  sleep 2
done

# Exécuter le schéma CQL
echo "Creating keyspace and tables..."
cqlsh $CASSANDRA_HOST $CASSANDRA_PORT -f db/cassandra/schema.cql

if [ $? -eq 0 ]; then
  echo "✓ Database initialization completed successfully!"
  echo ""
  echo "Created keyspace: message_service"
  echo "Created tables:"
  echo "  - conversations"
  echo "  - conversation_members"
  echo "  - messages"
  echo "  - conversation_messages (denormalized)"
  echo "  - user_conversations (denormalized)"
  echo "  - message_attachments"
  echo "  - message_status"
  exit 0
else
  echo "✗ Database initialization failed!"
  exit 1
fi
