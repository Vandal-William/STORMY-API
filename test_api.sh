#!/bin/bash

# Test health endpoint
echo "=== Testing /info endpoint ==="
curl -s http://message-service:3001/info
echo ""

# Test create conversation
echo "=== Testing POST /conversations ==="
curl -s -X POST http://message-service:3001/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Conversation",
    "type": "group",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "member_ids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"]
  }'
echo ""
echo ""

# Test list user conversations
echo "=== Testing GET /users/:user_id/conversations ==="
curl -s -X GET "http://message-service:3001/users/550e8400-e29b-41d4-a716-446655440000/conversations"
echo ""
echo ""

# Test create message
echo "=== Testing POST /messages ==="
curl -s -X POST http://message-service:3001/messages \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "550e8400-e29b-41d4-a716-446655440003",
    "sender_id": "550e8400-e29b-41d4-a716-446655440000",
    "content": "Hello, this is a test message",
    "type": "text"
  }'
echo ""
echo ""

echo "=== All tests completed ==="
