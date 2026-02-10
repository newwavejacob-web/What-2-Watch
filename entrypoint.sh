#!/bin/sh
set -e

echo "=== Seeding database ==="
./seed-db

echo ""
echo "=== Starting server ==="
exec ./vibe-server
