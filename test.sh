#!/bin/sh
curl -s -c /tmp/c.txt -X POST http://user-service:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@admin.com","password":"admin123"}' >/dev/null 2>&1

echo "Direct (user-service:3000):"
curl -s -b /tmp/c.txt http://user-service:3000/auth/me -w ' [%{http_code}]' | head -c 200
echo ""

echo ""
echo "Gateway (gateway:8080):"
curl -s -b /tmp/c.txt http://gateway:8080/auth/me -w ' [%{http_code}]' | head -c 200
echo ""

echo ""
echo "Messages via Gateway:"
curl -s -b /tmp/c.txt http://gateway:8080/messages/conversations -w ' [%{http_code}]' | head -c 200
echo ""
