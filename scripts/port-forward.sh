#!/bin/bash

# Script pour relancer les port-forwards Kubernetes
# Usage: bash scripts/port-forward.sh

echo "=== Killing existing port-forwards ==="
pkill -f "kubectl port-forward"
sleep 1
echo "✓ Port-forwards killed"

echo ""
echo "=== Starting new port-forwards ==="
echo "• PostgreSQL on 5432"
kubectl port-forward svc/postgres 5432:5432 &
sleep 1

echo "• Cassandra on 9042"
kubectl port-forward svc/cassandra 9042:9042 &
sleep 1

echo "• Gateway on 8080"
kubectl port-forward svc/gateway 8080:8080 &
sleep 1

echo ""
echo "✓ All port-forwards started!"
echo ""
echo "Services available at:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Cassandra: localhost:9042"
echo "  - Gateway: localhost:8080"
echo ""
echo "Press Ctrl+C to stop all port-forwards"
