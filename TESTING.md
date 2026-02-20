# Testing Guide: API, ETL, Load Tests

**Local only** – no Docker/CI in GitHub. Tests run on your machine with Docker databases.

See `tests/README.md` for consolidated layout: `apps/api/test/`, `apps/etl/tests/`, `tests/load-tests/`.

End-to-end testing for all APIs and ETL pipelines using Docker databases.

## Ports (Avoid Conflicts)

| Service       | Default Host Port | Used For              |
|---------------|-------------------|-----------------------|
| PostgreSQL    | 5432, 5433        | Often occupied        |
| MySQL         | 3306              | Often occupied        |
| MongoDB       | 27017             | Often occupied        |

**Test ports** (docker-compose.test.yml):

| Service       | Test Port | Container |
|---------------|-----------|-----------|
| PostgreSQL    | **15432** | postgres-test |
| PostgreSQL API| **15433** | postgres-api-test |
| MySQL         | **13306** | mysql-test |
| MongoDB       | **27018** | mongodb-test |

## Quick Start: Full Test Run with Logs

**Run everything step-by-step with logs saved:**

```bash
./scripts/run-all-tests-with-logs.sh 2>&1 | tee test-run.log
```

This script:
1. Starts Docker databases (Postgres 15432/15433, MySQL 13306, MongoDB 27018)
2. Waits for DBs healthy
3. Runs API migrations
4. Seeds ETL test data
5. Starts ETL service (background)
6. Runs ETL unit tests
7. Runs **all 7 pipeline direction tests** (Postgres↔Postgres, MySQL↔Postgres, etc.)
8. Runs ETL integration tests
9. Runs parallel pipeline tests
10. Runs API E2E tests
11. **Load testing (k6)** at the end

Logs are saved in `test-logs/`.

## Manual Quick Start

```bash
# 1. Start Docker databases
docker compose -f docker-compose.test.yml up -d

# 2. Run API migrations
cd apps/api
DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test" bun run db:migrate

# 3. Seed ETL test data (Postgres, MySQL, MongoDB)
cd apps/etl
python scripts/seed_test_data.py

# 4. Start ETL service (Terminal 1)
cd apps/etl && ./run.sh

# 5. Test all ETL endpoints (Terminal 2)
cd apps/etl && ./scripts/test_all_endpoints.sh

# 6. Run all tests
./scripts/run-all-tests.sh
```

## API (NestJS) Tests

### Unit Tests
```bash
cd apps/api
bun run test
```

### E2E Tests
Requires `postgres-api-test` running on port 15433.

```bash
cd apps/api
cp test/.env.e2e.example test/.env.e2e
# Edit test/.env.e2e if needed
DATABASE_URL="postgresql://api_test:api_test_pass@localhost:15433/ai_bi_test" bun run test:e2e
```

The E2E tests use `MockAuthGuard` so no real Supabase auth is required.

## ETL (FastAPI) Tests

### Quick Endpoint Test (curl)

Tests all ETL endpoints against local Docker DBs. No pytest/singer_sdk required.

```bash
# Requires: Docker DBs up, seed done, ETL running on :8001
cd apps/etl
./scripts/test_all_endpoints.sh
# Or: ./scripts/test_all_endpoints.sh http://localhost:8001
```

Covers: `/`, `/health`, `/test-connection`, `/discover-schema`, `/collect`, `/delta-check`, `/run-meltano-pipeline`, `/dbt-models`, auth rejection.

### Integration Tests
Requires:
1. Docker databases (postgres-test, mysql-test, mongodb-test)
2. ETL service running: `./run.sh` (port 8001)
3. Seed data: `python scripts/seed_test_data.py`

```bash
cd apps/etl
# Terminal 1: start ETL
./run.sh

# Terminal 2: run integration tests
ETL_BASE_URL=http://localhost:8001 pytest tests/test_etl_api_integration.py -v
```

Tests cover:
- `GET /`, `GET /health`
- `POST /test-connection` (postgresql, mysql, mongodb)
- `POST /discover-schema/{source_type}` (postgresql, mysql, mongodb)
- `POST /collect/{source_type}` (postgresql, mysql, mongodb)
- `POST /run-meltano-pipeline` (postgres-to-mongodb, mongodb-to-postgres)
- `POST /delta-check/{source_type}`

### All Pipeline Direction Tests (7 combinations)

```bash
cd apps/etl
pytest tests/test_all_pipeline_directions.py -v -m integration
```

| Source    | Destination | Test                           |
|-----------|-------------|--------------------------------|
| Postgres  | Postgres    | test_postgres_to_postgres      |
| MySQL     | Postgres    | test_mysql_to_postgres         |
| Postgres  | MySQL       | test_postgres_to_mysql         |
| MySQL     | MongoDB     | test_mysql_to_mongodb          |
| MongoDB   | MySQL       | test_mongodb_to_mysql          |
| Postgres  | MongoDB     | test_postgres_to_mongodb       |
| MongoDB   | Postgres    | test_mongodb_to_postgres       |

## Parallel Pipeline Execution

The system supports **parallel pipeline execution** for real-time usage:

- **ETL**: FastAPI handles concurrent requests; multiple `run-meltano-pipeline` calls run in parallel
- **API job processor**: Processes up to 5 pipeline jobs concurrently per queue (`PGMQ_PARALLEL_WORKERS`)
- **Manual runs**: `POST /pipelines/:id/run` fires async execution; multiple runs don't block each other

### Parallel ETL Tests

```bash
cd apps/etl
# Requires: Docker DBs + ETL service
pytest tests/test_etl_parallel_pipelines.py -v -m integration
```

## Load Tests (k6)

Install k6: https://k6.io/docs/getting-started/installation/

```bash
# API load test (default: 5 VUs, 30s)
k6 run tests/load-tests/k6-api.js

# ETL load test (10 VUs, parallel-style)
k6 run tests/load-tests/k6-etl.js

# ETL parallel pipeline load (5 VUs, 15 iterations of run-meltano-pipeline)
k6 run tests/load-tests/k6-etl-parallel-pipelines.js

# Custom: 10 VUs, 1 minute
k6 run --vus 10 --duration 1m tests/load-tests/k6-api.js
```

### Testing deployed endpoints (cloud.api.mantrixflow.com, cloud.api.etl.server.mantrixflow.com)

```bash
# Skip TLS verification if k6 fails with x509 certificate errors (macOS keychain)
K6_INSECURE_SKIP_TLS_VERIFY=true API_BASE_URL=https://cloud.api.mantrixflow.com k6 run --vus 5 --duration 30s tests/load-tests/k6-api.js
K6_INSECURE_SKIP_TLS_VERIFY=true ETL_BASE_URL=https://cloud.api.etl.server.mantrixflow.com k6 run --vus 10 --duration 30s tests/load-tests/k6-etl.js
```

**Parallel pipeline test** requires local ETL + local Docker DBs (Postgres 15432, MongoDB 27018); the deployed ETL cannot reach your localhost.

### Live ETL full flow (collect + emit + run-meltano-pipeline)

Tests the **live** ETL API (`cloud.api.etl.server.mantrixflow.com`). The ETL runs on Vercel, so it can only reach **cloud-accessible** databases (Supabase, MongoDB Atlas, etc.). Localhost DBs work only with local ETL (`ETL_BASE_URL=http://localhost:8001`).

**With cloud databases** (Supabase Postgres + MongoDB Atlas):

```bash
SOURCE_PG_HOST=db.xxx.supabase.co SOURCE_PG_PORT=5432 \
SOURCE_PG_DATABASE=postgres SOURCE_PG_USER=postgres SOURCE_PG_PASSWORD=xxx \
SOURCE_PG_SCHEMA=public SOURCE_PG_TABLE=your_table \
DEST_MONGO_URI="mongodb+srv://user:pass@cluster.mongodb.net/dbname" DEST_MONGO_DB=dbname \
ETL_AUTH_TOKEN=your-jwt ETL_BASE_URL=https://cloud.api.etl.server.mantrixflow.com \
K6_INSECURE_SKIP_TLS_VERIFY=true \
k6 run --vus 3 --duration 60s tests/load-tests/k6-etl-live-full.js
```

**With local ETL + local Docker DBs** (Docker + seed + ETL on :8001):

```bash
ETL_BASE_URL=http://localhost:8001 ETL_AUTH_TOKEN=test-token \
k6 run --vus 3 --duration 60s tests/load-tests/k6-etl-live-full.js
```

The script exercises: `/health`, `/`, `/test-connection`, `/discover-schema`, `/collect`, `/run-meltano-pipeline`.

## Dummy Data

### PostgreSQL (testdb, port 15432)
- `test_users`: 20 rows (id, name, email, age)
- `test_orders`: 10 rows (id, user_id, amount, status)

### MySQL (testdb, port 13306)
- `test_products`: 15 rows (id, name, sku, price, stock)

### MongoDB (testdb, port 27018)
- `test_customers`: 10 docs (_id, name, email, balance)

### API DB (ai_bi_test, port 15433)
Migrations create all schema. Seed organizations/users for full E2E if needed.

## Docker Commands

```bash
# Start
docker compose -f docker-compose.test.yml up -d

# Logs
docker compose -f docker-compose.test.yml logs -f

# Stop
docker compose -f docker-compose.test.yml down
```

## Environment Overrides

| Variable        | Default  | Description                    |
|-----------------|----------|--------------------------------|
| PG_TEST_PORT    | 15432    | PostgreSQL test port          |
| PG_TEST_HOST    | localhost| PostgreSQL host               |
| MYSQL_TEST_PORT | 13306    | MySQL test port               |
| MONGO_TEST_PORT | 27018    | MongoDB test port             |
| ETL_BASE_URL    | localhost:8001 | ETL service URL          |
| API_BASE_URL    | localhost:3000 | API URL for load tests   |
