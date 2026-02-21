#!/bin/bash
# Run all tests step-by-step with logs. Use: ./scripts/run-all-tests-with-logs.sh 2>&1 | tee test-run.log
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${ROOT}/test-logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
MAIN_LOG="${LOG_DIR}/run-${TIMESTAMP}.log"

log() { echo "[$(date +%H:%M:%S)] $*" | tee -a "$MAIN_LOG"; }
log_section() { echo ""; log "========== $* =========="; echo ""; }

cd "$ROOT"
exec 1> >(tee -a "$MAIN_LOG") 2>&1

log "Test run started. Full log: $MAIN_LOG"
log "Docker ports: Postgres 15432/15433, MySQL 13306, MongoDB 27018"

# ---------------------------------------------------------------------------
# STEP 1: Start Docker databases
# ---------------------------------------------------------------------------
log_section "STEP 1: Start Docker test databases"
docker compose -f docker-compose.test.yml up -d 2>/dev/null || {
  log "WARN: Docker compose failed (e.g. credential helper). If DBs already running, continue."
}
sleep 8
docker compose -f docker-compose.test.yml ps 2>/dev/null || true

# ---------------------------------------------------------------------------
# STEP 2: Wait for DBs healthy
# ---------------------------------------------------------------------------
log_section "STEP 2: Wait for databases healthy"
for i in 1 2 3 4 5 6 7 8 9 10; do
  if docker compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U testuser -d testdb 2>/dev/null; then
    log "PostgreSQL (15432) ready"
    break
  fi
  log "Waiting for PostgreSQL... ($i/10)"
  sleep 2
done
for i in 1 2 3 4 5 6 7 8 9 10; do
  if docker compose -f docker-compose.test.yml exec -T postgres-api-test pg_isready -U api_test -d ai_bi_test 2>/dev/null; then
    log "PostgreSQL API (15433) ready"
    break
  fi
  log "Waiting for PostgreSQL API... ($i/10)"
  sleep 2
done
for i in 1 2 3 4 5; do
  if docker compose -f docker-compose.test.yml exec -T mysql-test mysqladmin ping -h localhost -u testuser -ptestpass 2>/dev/null; then
    log "MySQL (13306) ready"
    break
  fi
  log "Waiting for MySQL... ($i/5)"
  sleep 2
done
for i in 1 2 3 4 5; do
  if docker compose -f docker-compose.test.yml exec -T mongodb-test mongosh --eval "db.adminCommand('ping')" 2>/dev/null; then
    log "MongoDB (27018) ready"
    break
  fi
  log "Waiting for MongoDB... ($i/5)"
  sleep 2
done

# ---------------------------------------------------------------------------
# STEP 3: API migrations + E2E env (postgres-api-test = nestjs-supabase :15433)
# ---------------------------------------------------------------------------
log_section "STEP 3: Run API migrations (postgres-api-test :15433)"
cd apps/api
export DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test"
[ ! -f test/.env.e2e ] && cp test/.env.e2e.example test/.env.e2e && log "Created test/.env.e2e from example"
bun run db:migrate 2>/dev/null || log "WARN: Drizzle migrations failed (DB may not be up)"
# Apply extra migrations (0013-0021) not in drizzle journal - required for full schema
for f in 0013_refactor_to_dynamic_data_sources 0014_pipeline_lifecycle 0015_pipeline_scheduling 0016_pipeline_incremental_sync_fixes 0017_add_polling_trigger_type 0018_add_transform_script 0019_remove_column_mappings 0020_add_migration_state 0021_add_dbt_models_to_destination_schemas; do
  if [ -f "src/database/drizzle/migrations/${f}.sql" ]; then
    PGPASSWORD=api_test_pass psql -h localhost -p 15433 -U api_test -d ai_bi_test -f "src/database/drizzle/migrations/${f}.sql" 2>/dev/null || true
  fi
done
# ETL jobs + pgmq (requires pgmq extension - postgres-api-test uses pg16-pgmq image)
bun run db:migrate:etl 2>/dev/null || log "WARN: ETL migration failed (ensure postgres-api-test has pgmq)"
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 4: Seed ETL test data
# ---------------------------------------------------------------------------
log_section "STEP 4: Seed ETL test data"
cd apps/etl
[ -d .venv ] && source .venv/bin/activate
python3 scripts/seed_test_data.py 2>/dev/null || .venv/bin/python scripts/seed_test_data.py 2>/dev/null || {
  log "WARN: Seed failed. Ensure Docker DBs are running."
}
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 5: Start ETL service (background)
# ---------------------------------------------------------------------------
log_section "STEP 5: Start ETL service (port 8001)"
cd apps/etl
[ -d .venv ] && source .venv/bin/activate
if ! lsof -i :8001 2>/dev/null | grep -q LISTEN; then
  ( .venv/bin/python -m uvicorn main:app --host 0.0.0.0 --port 8001 2>&1 | tee "${LOG_DIR}/etl-${TIMESTAMP}.log" & )
  sleep 5
  log "ETL started in background"
else
  log "ETL already running on 8001"
fi
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 6: API unit tests
# ---------------------------------------------------------------------------
log_section "STEP 6: API unit tests"
cd apps/api
bun run test -- --passWithNoTests 2>&1 | tee "${LOG_DIR}/api-unit-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 7: All pipeline direction tests (7 combinations)
# ---------------------------------------------------------------------------
log_section "STEP 7: All pipeline direction tests"
log "Postgres→Postgres, MySQL→Postgres, Postgres→MySQL, MySQL→MongoDB,"
log "MongoDB→MySQL, Postgres→MongoDB, MongoDB→Postgres"
cd apps/etl
ETL_BASE_URL="${ETL_BASE_URL:-http://localhost:8001}" .venv/bin/python -m pytest tests/test_all_pipeline_directions.py -v -m integration --tb=short 2>&1 | tee "${LOG_DIR}/pipeline-directions-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 8: ETL unit tests (no Docker required)
# ---------------------------------------------------------------------------
log_section "STEP 8: ETL unit tests"
cd apps/etl
.venv/bin/python -m pytest tests/ -v -m "not integration" --tb=short 2>&1 | tee "${LOG_DIR}/etl-unit-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 9: ETL integration tests
# ---------------------------------------------------------------------------
log_section "STEP 9: ETL integration tests"
cd apps/etl
ETL_BASE_URL="${ETL_BASE_URL:-http://localhost:8001}" .venv/bin/python -m pytest tests/test_etl_api_integration.py -v -m integration --tb=short 2>&1 | tee "${LOG_DIR}/etl-integration-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 10: Parallel pipeline tests
# ---------------------------------------------------------------------------
log_section "STEP 10: Parallel pipeline tests (different databases)"
cd apps/etl
ETL_INTEGRATION_FULL=1 ETL_BASE_URL="${ETL_BASE_URL:-http://localhost:8001}" .venv/bin/python -m pytest tests/test_etl_parallel_pipelines.py -v -m integration --tb=short 2>&1 | tee "${LOG_DIR}/parallel-pipelines-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 11: API E2E tests
# ---------------------------------------------------------------------------
log_section "STEP 11: API E2E tests"
cd apps/api
DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test" bun run test:e2e 2>&1 | tee "${LOG_DIR}/api-e2e-${TIMESTAMP}.log" || true
cd "$ROOT"

# ---------------------------------------------------------------------------
# STEP 12: Load testing (k6)
# ---------------------------------------------------------------------------
log_section "STEP 12: Load testing (k6)"
if command -v k6 &>/dev/null; then
  log "k6 API load test..."
  if lsof -i :3000 2>/dev/null | grep -q LISTEN; then
    k6 run --vus 5 --duration 15s tests/load-tests/k6-api.js 2>&1 | tee "${LOG_DIR}/load-api-${TIMESTAMP}.log" || true
  else
    log "WARN: NestJS API not running on 3000. Start with: cd apps/api && bun run start"
    log "Skipping k6 API load test."
  fi
  log "k6 ETL load test..."
  ETL_BASE_URL="${ETL_BASE_URL:-http://localhost:8001}" k6 run --vus 10 --duration 15s tests/load-tests/k6-etl.js 2>&1 | tee "${LOG_DIR}/load-etl-${TIMESTAMP}.log" || true
  log "k6 ETL parallel pipeline load test..."
  ETL_BASE_URL="${ETL_BASE_URL:-http://localhost:8001}" k6 run --vus 3 --iterations 9 tests/load-tests/k6-etl-parallel-pipelines.js 2>&1 | tee "${LOG_DIR}/load-etl-parallel-${TIMESTAMP}.log" || true
else
  log "k6 not installed. Skip load tests. Install: https://k6.io/docs/getting-started/installation/"
fi

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------
log_section "DONE"
log "Logs saved in: $LOG_DIR"
log "Main log: $MAIN_LOG"
log "To stop: docker compose -f docker-compose.test.yml down"
