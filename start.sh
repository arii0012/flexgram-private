#!/bin/bash
set -e
if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env — please edit passwords and JWT_SECRET!"
fi
docker compose up -d --build
echo ""
echo "FlexGram is starting..."
echo "API: http://localhost:8080"
echo "Check: curl http://localhost:8080/"
