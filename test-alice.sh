#!/bin/sh
echo "=== Login ==="
curl -s -c /tmp/c.txt -X POST http://user-service:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}' 

echo ""
echo ""
echo "=== Cookies saved ==="
cat /tmp/c.txt | grep -v "^#"

echo ""
echo "=== Direct test (user-service:3000) ==="
curl -s -b /tmp/c.txt http://user-service:3000/auth/me -w ' [%{http_code}]' | head -c 200
echo ""

echo ""
echo "=== Gateway test (gateway:8080/auth/me) ==="
curl -s -b /tmp/c.txt http://gateway:8080/auth/me -w ' [%{http_code}]' | head -c 200
echo ""

echo ""
echo "=== Messages via Gateway (gateway:8080/messages/conversations) ==="
curl -s -b /tmp/c.txt http://gateway:8080/messages/conversations -w ' [%{http_code}]' | head -c 200
echo ""
