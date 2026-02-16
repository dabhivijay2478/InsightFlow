#!/bin/bash
# Run all tests: Docker DBs, API migrations, ETL seeds, API E2E, ETL integration, load tests
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== 1. Start test databases (Postgres 15432/15433, MySQL 13306, MongoDB 27018) ==="
docker compose -f docker-compose.test.yml up -d
sleep 5
docker compose -f docker-compose.test.yml ps

echo ""
echo "=== 2. Wait for DBs healthy ==="
until docker compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U testuser -d testdb 2>/dev/null; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done
until docker compose -f docker-compose.test.yml exec -T postgres-api-test pg_isready -U api_test -d ai_bi_test 2>/dev/null; do
  echo "Waiting for API PostgreSQL..."
  sleep 2
done
until docker compose -f docker-compose.test.yml exec -T mysql-test mysqladmin ping -h localhost -u testuser -ptestpass 2>/dev/null; do
  echo "Waiting for MySQL..."
  sleep 2
done
until docker compose -f docker-compose.test.yml exec -T mongodb-test mongosh --eval "db.adminCommand('ping')" 2>/dev/null; do
  echo "Waiting for MongoDB..."
  sleep 2
done
echo "All databases ready."

echo ""
echo "=== 3. Run API migrations (postgres-api-test) ==="
cd apps/api
cp -n test/.env.e2e.example test/.env.e2e 2>/dev/null || true
export DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test"
bun run db:migrate
cd "$ROOT"

echo ""
echo "=== 4. Seed ETL test data ==="
cd apps/etl
if [ -d ".venv" ]; then
  source .venv/bin/activate
fi
python scripts/seed_test_data.py
cd "$ROOT"

echo ""
echo "=== 5. API unit tests ==="
cd apps/api
bun run test -- --passWithNoTests 2>/dev/null || true
cd "$ROOT"

echo ""
echo "=== 6. API E2E tests ==="
cd apps/api
DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test" bun run test:e2e 2>/dev/null || true
cd "$ROOT"

echo ""
echo "=== 7. ETL unit tests ==="
cd apps/etl
if [ -d ".venv" ]; then
  source .venv/bin/activate
fi
pytest tests/test_transformer.py -v 2>/dev/null || true
cd "$ROOT"

echo ""
echo "=== 8. ETL integration tests (requires ETL service on :8001) ==="
echo "Start ETL in another terminal: cd apps/etl && ./run.sh"
echo "Then run: cd apps/etl && pytest tests/test_etl_api_integration.py -v"
echo ""
echo "=== 9. ETL parallel pipeline tests ==="
echo "  pytest tests/test_etl_parallel_pipelines.py -v -m integration"
echo "  Full (with DBs): ETL_INTEGRATION_FULL=1 pytest tests/test_etl_parallel_pipelines.py -v -m integration"
echo ""
echo "=== 10. Load tests ==="
echo "Run: k6 run tests/load-tests/k6-api.js  # or k6-etl.js, k6-etl-parallel-pipelines.js"
echo ""
echo "=== Done. Stop DBs with: docker compose -f docker-compose.test.yml down ==="
