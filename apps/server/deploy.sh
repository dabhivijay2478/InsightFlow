#!/bin/bash
# ============================================================================
# MantrixFlow — Deploy Script
# Pulls latest code from both repos and rebuilds containers
# Usage: ./deploy.sh [api|etl]   (no arg = deploy both)
# ============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

COMPOSE_FILE="docker-compose.prod.yml"

deploy_api() {
    echo "→ Pulling arcyria-server..."
    cd arcyria-server && git pull origin main && cd ..
}

deploy_etl() {
    echo "→ Pulling arcyria-elt..."
    cd arcyria-elt && git pull origin main && cd ..
}

case "${1:-all}" in
    api)
        deploy_api
        echo "→ Rebuilding API..."
        docker compose -f "$COMPOSE_FILE" up -d --build api
        ;;
    etl)
        deploy_etl
        echo "→ Rebuilding ETL..."
        docker compose -f "$COMPOSE_FILE" up -d --build etl
        ;;
    all)
        deploy_api
        deploy_etl
        echo "→ Rebuilding all services..."
        docker compose -f "$COMPOSE_FILE" up -d --build
        ;;
    *)
        echo "Usage: $0 [api|etl|all]"
        exit 1
        ;;
esac

echo "→ Cleaning unused images..."
docker image prune -f

echo ""
echo "✓ Deploy complete"
echo ""
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
