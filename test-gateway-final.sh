#!/bin/sh

# 1. Login pour obtenir cookies
echo "=== 1. Login sur user-service ==="
curl -s -c /tmp/cookies.txt -X POST http://user-service:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@admin.com","password":"admin123"}' | head -c 150
echo ""

# 2. Afficher les cookies
echo ""
echo "=== 2. Cookies reçus ==="
cat /tmp/cookies.txt | grep -v "^#" | grep -v "^$"

# 3. Test DIRECT sur user-service
echo ""
echo "=== 3. Test DIRECT /auth/me (user-service:3000) ==="
curl -s -b /tmp/cookies.txt http://user-service:3000/auth/me -w ' [HTTP %{http_code}]\n'

# 4. Test VIA GATEWAY 
echo ""
echo "=== 4. Test VIA GATEWAY /auth/me (gateway:8080) ==="
curl -s -b /tmp/cookies.txt http://gateway:8080/auth/me -w ' [HTTP %{http_code}]\n'

# 5. Test /messages via gateway
echo ""
echo "=== 5. Test VIA GATEWAY /messages/conversations ==="
curl -s -b /tmp/cookies.txt http://gateway:8080/messages/conversations -w ' [HTTP %{http_code}]\n' | head -c 200
echo ""

# 6. Check gateway logs pour voir si cookies sont transmis
echo ""
echo "=== 6. Gateway logs (derniers tokens) ==="
