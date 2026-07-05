# Salesforce to PostgreSQL Testing Guide

Use this guide to verify the Salesforce connector from unit tests through a
manual Salesforce-to-PostgreSQL pipeline run.

## 1. Automated Regression Tests

Python ELT:

```bash
cd apps/server/elt-server
./.venv/bin/python -m pytest tests/test_salesforce_source.py tests/test_sync_route.py tests/test_state_sync.py
```

Go API:

```bash
cd apps/server/main-server
GOCACHE=/tmp/mantrixflow-go-build \
GOMODCACHE=/tmp/mantrixflow-go-mod \
GOPROXY=file:///Users/vijay.d/go/pkg/mod/cache/download \
go test ./...
```

Frontend:

```bash
cd apps/app
bun test app/workspace/connections/__tests__/credentialForm.test.ts lib/pipelines/__tests__/schema-table.test.ts
npm run build
```

Salesforce streamer syntax:

```bash
cd /Users/vijay.d/vijay.d/Vapps/incomplete/ai-bi
apps/server/elt-server/.venv/bin/python -m compileall apps/salesforce-streamer
```

## 2. Local Service Smoke Test

Start services:

```bash
cd apps/server/main-server
go run ./cmd/server
```

```bash
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio
```

```bash
cd apps/app
bun run dev
```

Open the app and create:

- One Salesforce source
- One PostgreSQL destination
- One pipeline using `salesforce.Account`

## 3. Test Connection

Expected result:

- Go API accepts Salesforce as a valid source connector.
- Go API encrypts saved secrets.
- ELT `/test-connection` authenticates with Salesforce.
- UI receives no raw secret values back.

Manual checks:

1. Create a Salesforce source connection.
2. Click test connection.
3. Confirm success message includes visible object count.
4. Reopen the connection and verify secret fields are masked.

Failure checks:

- Wrong refresh token should produce a reconnect-required message.
- Wrong sandbox OAuth after a sandbox refresh should produce the sandbox refresh
  message.
- Wrong username/password/security token should not leak any credential value in
  the error.

## 4. Discovery Test

Discover schema for the Salesforce connection.

Expected:

- `salesforce` schema is present.
- Standard objects like `Account` and `Contact` are present when visible to the
  Salesforce user.
- Custom objects ending in `__c` are present when visible.
- Stream keys are canonical, for example `salesforce.Account`.
- Raw DuckDB table names are `salesforce__Account`.
- `Id` is the primary key.
- Incremental candidates include `SystemModstamp` or another suitable key.
- Formula fields have `calculated: true`.
- Shield fields have `encrypted: true`.

Focused object discovery:

```json
{
  "source_type": "salesforce",
  "connection_config": {
    "type": "salesforce",
    "auth_mode": "oauth2_refresh",
    "client_id": "...",
    "client_secret": "...",
    "refresh_token": "...",
    "is_sandbox": false
  },
  "table_name": "Account"
}
```

## 5. Preview Test

Preview `salesforce.Account`.

Expected:

- Preview returns up to the requested limit.
- Records include Salesforce fields and lineage columns.
- Response does not include raw credentials.

Check with a small limit first:

```json
{
  "source_type": "salesforce",
  "resource": "salesforce.Account",
  "limit": 5
}
```

## 6. PostgreSQL Destination Table Test

Create the destination table before running the pipeline:

```sql
drop table if exists public.salesforce_accounts;

create table public.salesforce_accounts (
  id text primary key,
  name text,
  system_modstamp timestamptz,
  _sf_is_deleted boolean,
  _sf_deleted_at timestamptz,
  _sf_synced_at timestamptz
);
```

Negative test:

1. Drop the destination table.
2. Run the pipeline.
3. Confirm the run fails at preflight with a missing destination table error.
4. Confirm the runner does not create the table.

## 7. Full Sync Test

Use a dbt model like:

```sql
select
  Id as id,
  Name as name,
  SystemModstamp as system_modstamp,
  _sf_is_deleted,
  _sf_deleted_at,
  _sf_synced_at
from {{ source('raw', 'salesforce__Account') }}
```

Run a full sync.

Expected:

- Pipeline enters the queue and runs through the Go dispatcher.
- ELT stages Salesforce raw rows in DuckDB.
- dbt model runs against DuckDB raw staging.
- PostgreSQL delivery upserts into the existing destination table.
- Callback metadata includes `delivery_outputs`, `staging_size_bytes`,
  `dbt_models_run`, and `no_pk_warnings`.
- DuckDB staging file is deleted after state extraction.

Validate PostgreSQL:

```sql
select count(*) from public.salesforce_accounts;

select id, name, system_modstamp, _sf_is_deleted
from public.salesforce_accounts
order by system_modstamp desc nulls last
limit 10;
```

## 8. Incremental Sync Test

1. Run an initial full sync.
2. Update one Account in Salesforce.
3. Run incremental sync.
4. Confirm only changed rows are staged and delivered.
5. Confirm checkpoint state advances.

Expected:

- REST SOQL query uses the replication key watermark.
- Destination row is upserted.
- Existing unchanged rows are not duplicated.

## 9. Delete Detection Test

1. Create a test Account in Salesforce.
2. Run sync and confirm it appears in PostgreSQL.
3. Delete the Account in Salesforce.
4. Run incremental sync.
5. Confirm a staged/delete lineage row arrives.

Expected row shape in downstream SQL:

```text
_sf_is_deleted = true
_sf_deleted_at is not null
```

Whether the destination keeps deleted records, filters them, or marks them is a
dbt SQL decision.

## 10. Formula Refresh Test

Configure stream settings:

```json
{
  "salesforce.Account": {
    "formula_refresh_interval_hours": 1
  }
}
```

Expected:

- Scheduled incremental runs refresh formula fields when due.
- Rows include `_sf_formula_refreshed_at`.
- Refresh rows merge into the same raw table.

## 11. Bulk API Test

For a large object, set a low threshold in test:

```json
{
  "salesforce.Account": {
    "bulk_threshold_rows": 1
  }
}
```

Run a full sync.

Expected:

- ELT creates a Salesforce Bulk API 2.0 query job.
- ELT polls until `JobComplete`.
- ELT downloads result CSV batches.
- Rows stage into DuckDB and flow through the same dbt/delivery path.

## 12. OAuth Callback Test

With backend env configured:

```bash
SALESFORCE_OAUTH_CLIENT_ID=...
SALESFORCE_OAUTH_CLIENT_SECRET=...
SALESFORCE_OAUTH_REDIRECT_BASE_URL=https://api.example.com
APP_WEB_URL=https://app.example.com
```

Start OAuth:

```http
POST /api/v1/organizations/{organizationId}/salesforce/oauth/start
```

Expected:

- Response returns `authorizeUrl`.
- Salesforce redirects back to `/api/v1/salesforce/oauth/callback`.
- Existing data source connection is saved with `auth_mode = oauth2_refresh`.
- Stored refresh token and client secret are encrypted.
- UI redirects back with `salesforce=connected`.

## 13. CDC Streamer Test

Syntax test:

```bash
apps/server/elt-server/.venv/bin/python -m compileall apps/salesforce-streamer
```

Queue bridge test in an environment with pgmq:

1. Set `DATABASE_URL`.
2. Set `SALESFORCE_STREAM_SUBSCRIPTIONS_JSON`.
3. Run `python main.py`.
4. Simulate a subscription failure or idle timeout.
5. Confirm `incremental_sync` receives a polling fallback job.

Expected pgmq payload:

```json
{
  "name": "incremental-sync",
  "data": {
    "pipelineId": "...",
    "organizationId": "...",
    "userId": "...",
    "triggerType": "salesforce_cdc"
  }
}
```

Live CDC requires Salesforce Pub/Sub generated protobuf bindings in the runtime
image and Salesforce Change Data Capture enabled for the selected objects.

## 14. Security Checklist

- No credential value appears in UI responses.
- No credential value appears in ELT or Go error messages.
- `refresh_token`, `client_secret`, `private_key`, `password`, and
  `security_token` are encrypted at rest.
- Frontend never calls ELT directly.
- ELT callbacks use internal callback auth.
- Destination tables are never created or altered by the runner.

## 15. Troubleshooting Matrix

| Symptom | Check |
| --- | --- |
| OAuth callback fails | Callback URL exactly matches Salesforce Connected App |
| `invalid_grant` | OAuth access revoked, or sandbox refreshed |
| Custom object missing | Salesforce user/app lacks object permission |
| Shield value masked | Grant `View Encrypted Data` or exclude field |
| Destination table error | Create PostgreSQL table manually |
| Column mismatch | Match dbt output names to destination columns |
| Incremental misses row | Check replication key and Salesforce timestamp |
| Bulk job timeout | Increase Bulk timeout or use REST for that stream |
