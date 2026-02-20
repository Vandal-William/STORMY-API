-- Test data seeding for message_service keyspace
-- This file inserts sample conversations, members, and messages

-- Conversations
INSERT INTO message_service.conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440001, 1, 'direct', 'Alice & Bob Direct', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440002, 1, 'group', 'Project Discussion', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.conversations (id, created_by, conversation_type, name, created_at, updated_at)
VALUES (650e8400-e29b-41d4-a716-446655440003, 2, 'group', 'Random Chat', toTimestamp(now()), toTimestamp(now()));

-- Conversation Members
INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440100, 650e8400-e29b-41d4-a716-446655440001, 1, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440101, 650e8400-e29b-41d4-a716-446655440001, 2, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440102, 650e8400-e29b-41d4-a716-446655440002, 1, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440103, 650e8400-e29b-41d4-a716-446655440002, 2, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440104, 650e8400-e29b-41d4-a716-446655440002, 3, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440105, 650e8400-e29b-41d4-a716-446655440003, 2, 'member', toTimestamp(now()));

INSERT INTO message_service.conversation_members (id, conversation_id, user_id, role, joined_at)
VALUES (750e8400-e29b-41d4-a716-446655440106, 650e8400-e29b-41d4-a716-446655440003, 3, 'member', toTimestamp(now()));

-- Messages
INSERT INTO message_service.messages (id, conversation_id, sender_id, content, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440001, 650e8400-e29b-41d4-a716-446655440001, 1, 'Hey Bob, how are you?', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.messages (id, conversation_id, sender_id, content, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440002, 650e8400-e29b-41d4-a716-446655440001, 2, 'Hi Alice! I am doing great, thanks for asking!', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.messages (id, conversation_id, sender_id, content, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440003, 650e8400-e29b-41d4-a716-446655440002, 1, 'Team, please review the latest design specs', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.messages (id, conversation_id, sender_id, content, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440004, 650e8400-e29b-41d4-a716-446655440002, 2, 'Looks good! I will update the frontend accordingly', toTimestamp(now()), toTimestamp(now()));

INSERT INTO message_service.messages (id, conversation_id, sender_id, content, created_at, updated_at)
VALUES (750e8400-e29b-41d4-a716-446655440005, 650e8400-e29b-41d4-a716-446655440003, 3, 'Anyone up for lunch later?', toTimestamp(now()), toTimestamp(now()));

-- User Conversations Index
INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (1, 650e8400-e29b-41d4-a716-446655440001, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (1, 650e8400-e29b-41d4-a716-446655440002, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (2, 650e8400-e29b-41d4-a716-446655440001, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (2, 650e8400-e29b-41d4-a716-446655440002, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (2, 650e8400-e29b-41d4-a716-446655440003, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (3, 650e8400-e29b-41d4-a716-446655440002, toTimestamp(now()));

INSERT INTO message_service.user_conversations (user_id, conversation_id, last_activity)
VALUES (3, 650e8400-e29b-41d4-a716-446655440003, toTimestamp(now()));
