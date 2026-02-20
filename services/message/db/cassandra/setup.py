#!/usr/bin/env python3
"""
Cassandra database initialization script for message service.
Supports robust initialization with retry logic and verification.
"""

import sys
import time
import logging
from pathlib import Path
from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider
from cassandra.pool import HostConnectionPool

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class CassandraInitializer:
    def __init__(self, contact_points, port=9042, username=None, password=None, retries=30, retry_delay=2):
        self.contact_points = contact_points if isinstance(contact_points, list) else [contact_points]
        self.port = port
        self.username = username
        self.password = password
        self.retries = retries
        self.retry_delay = retry_delay
        self.cluster = None
        self.session = None

    def connect(self):
        """Connect to Cassandra with retry logic."""
        auth_provider = None
        if self.username and self.password:
            auth_provider = PlainTextAuthProvider(username=self.username, password=self.password)

        for attempt in range(self.retries):
            try:
                logger.info(f"Connection attempt {attempt + 1}/{self.retries}...")
                self.cluster = Cluster(
                    contact_points=self.contact_points,
                    port=self.port,
                    auth_provider=auth_provider,
                    connect_timeout=10,
                    control_connection_timeout=10,
                )
                self.session = self.cluster.connect()
                logger.info("✓ Successfully connected to Cassandra!")
                return True
            except Exception as e:
                logger.warning(f"Connection failed: {e}")
                if attempt < self.retries - 1:
                    logger.info(f"Retrying in {self.retry_delay} seconds...")
                    time.sleep(self.retry_delay)
                else:
                    logger.error("✗ Failed to connect to Cassandra after all retries")
                    return False

    def initialize_schema(self, schema_file):
        """Load and execute the schema file."""
        if not self.session:
            logger.error("Not connected to Cassandra")
            return False

        try:
            schema_path = Path(schema_file)
            if not schema_path.exists():
                logger.error(f"Schema file not found: {schema_file}")
                return False

            logger.info(f"Reading schema from {schema_file}...")
            with open(schema_path, 'r') as f:
                schema_content = f.read()

            # Split by semicolon and execute each statement
            statements = [stmt.strip() for stmt in schema_content.split(';') if stmt.strip()]

            logger.info(f"Executing {len(statements)} statements...")
            for i, statement in enumerate(statements, 1):
                try:
                    logger.debug(f"Executing statement {i}/{len(statements)}")
                    self.session.execute(statement)
                except Exception as e:
                    logger.warning(f"Statement {i} error: {e}")
                    # Continue with next statement - some may fail if they already exist

            logger.info("✓ Schema initialization completed!")
            return True

        except Exception as e:
            logger.error(f"Schema initialization failed: {e}")
            return False

    def verify_schema(self):
        """Verify that the schema was created correctly."""
        if not self.session:
            logger.error("Not connected to Cassandra")
            return False

        try:
            logger.info("Verifying schema...")

            # Check keyspace
            keyspace_query = "SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = 'message_service'"
            result = self.session.execute(keyspace_query).one()
            if result:
                logger.info("✓ Keyspace 'message_service' exists")
            else:
                logger.error("✗ Keyspace 'message_service' not found")
                return False

            # Check tables
            table_query = "SELECT table_name FROM system_schema.tables WHERE keyspace_name = 'message_service'"
            tables = self.session.execute(table_query)
            table_names = [row.table_name for row in tables]

            expected_tables = [
                'conversations',
                'conversation_members',
                'messages',
                'conversation_messages',
                'user_conversations',
                'message_attachments',
                'message_status',
            ]

            logger.info(f"Found {len(table_names)} tables: {', '.join(sorted(table_names))}")

            missing_tables = set(expected_tables) - set(table_names)
            if missing_tables:
                logger.warning(f"Missing tables: {', '.join(missing_tables)}")
                return False

            logger.info("✓ All required tables found")
            return True

        except Exception as e:
            logger.error(f"Schema verification failed: {e}")
            return False

    def close(self):
        """Close the connection."""
        if self.cluster:
            self.cluster.shutdown()
            logger.info("Connection closed")

    def run(self, schema_file):
        """Run the full initialization process."""
        try:
            if not self.connect():
                return False

            if not self.initialize_schema(schema_file):
                return False

            if not self.verify_schema():
                return False

            logger.info("\n✓ Database initialization completed successfully!")
            logger.info("\nDATABASE SUMMARY:")
            logger.info("  Keyspace: message_service")
            logger.info("  Tables:")
            logger.info("    - conversations (stores conversation metadata)")
            logger.info("    - conversation_members (stores members and roles)")
            logger.info("    - messages (stores message content)")
            logger.info("    - conversation_messages (denormalized for pagination)")
            logger.info("    - user_conversations (index for user's conversations)")
            logger.info("    - message_attachments (stores file metadata)")
            logger.info("    - message_status (tracks delivery/read status)")

            return True

        finally:
            self.close()


def main():
    # Parse arguments
    contact_points = sys.argv[1] if len(sys.argv) > 1 else 'localhost'
    port = int(sys.argv[2]) if len(sys.argv) > 2 else 9042
    schema_file = sys.argv[3] if len(sys.argv) > 3 else 'db/cassandra/schema.cql'
    username = sys.argv[4] if len(sys.argv) > 4 else None
    password = sys.argv[5] if len(sys.argv) > 5 else None

    logger.info(f"Cassandra Database Initializer")
    logger.info(f"Host: {contact_points}")
    logger.info(f"Port: {port}")
    logger.info(f"Schema file: {schema_file}")

    initializer = CassandraInitializer(
        contact_points=contact_points,
        port=port,
        username=username,
        password=password,
    )

    success = initializer.run(schema_file)
    sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()
