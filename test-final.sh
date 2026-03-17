#!/bin/sh
echo "=== Login alice_wonder ==="
curl -s -c /tmp/c.txt -X POST http://user-service:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice_wonder","password":"password123"}' 

echo ""
echo ""
echo "=== Cookies ==="
cat /tmp/c.txt | grep -v "^#"

echo ""
echo "=== Direct (user-service:3000) ==="
curl -s -b /tmp/c.txt http://user-service:3000/auth/me -w ' [HTTP %{http_code}]' | head -c 200
echo ""

echo ""
echo "=== Via Gateway (gateway:8080/auth/me) ==="
curl -s -b /tmp/c.txt http://gateway:8080/auth/me -w ' [HTTP %{http_code}]' | head -c 200
echo ""

echo ""
echo "=== Messages via Gateway ==="
curl -s -b /tmp/c.txt http://gateway:8080/messages/conversations -w ' [HTTP %{http_code}]' | head -c 200
echo ""
