# Salesforce to PostgreSQL Setup Guide

This guide sets up Salesforce as a MantrixFlow source and PostgreSQL as the
destination.

Important invariant: Salesforce rows do not load directly into PostgreSQL.
MantrixFlow runs the normal path:

```text
Go API -> Python ELT -> DuckDB raw staging -> dbt SQL -> existing PostgreSQL table
```

The runner does not create destination tables. Create every PostgreSQL
destination table before running the pipeline.

## 1. Services

Run the active services:

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

## 2. Required Backend Environment

Set the existing app and ELT variables first:

```bash
DATABASE_URL=postgresql://...
ENCRYPTION_MASTER_KEY=at-least-32-characters
ELT_PYTHON_SERVICE_URL=http://127.0.0.1:8000
ELT_INTERNAL_TOKEN=...
CALLBACK_TOKEN=...
API_PUBLIC_URL=http://127.0.0.1:5000
APP_WEB_URL=http://localhost:3000
```

For Salesforce OAuth Connected App login, also set:

```bash
SALESFORCE_OAUTH_CLIENT_ID=...
SALESFORCE_OAUTH_CLIENT_SECRET=...
SALESFORCE_OAUTH_REDIRECT_BASE_URL=https://api.example.com
```

The callback URL configured in Salesforce must be:

```text
https://api.example.com/api/v1/salesforce/oauth/callback
```

For local OAuth testing, expose the Go API with a tunnel and use the tunnel URL
as `SALESFORCE_OAUTH_REDIRECT_BASE_URL`.

## 3. Salesforce Connected App

In Salesforce Setup -> App Manager -> New Connected App:

- Enable OAuth Settings.
- Callback URL:
  `https://api.example.com/api/v1/salesforce/oauth/callback`
- OAuth scopes:
  - `api`
  - `refresh_token`
  - `offline_access`
  - `openid`

Copy the Consumer Key and Consumer Secret into the Go API env vars above.

### Long-term OAuth policy

For the older Connected App policy screen, set:

```text
Refresh Token Policy: Refresh token is valid until revoked
```

For the newer External Client App screen, Salesforce may only show idle-expiry
options. Use this production-friendly setting:

```text
Refresh Token Policy: Expire refresh token if not used for specific time
Refresh Token Validity Period: use the largest period your org allows
Refresh Token Validity Unit: Day(s) or Month(s), depending on availability
IP Relaxation: Relax IP restrictions, unless your deployment uses fixed allowed IPs
```

With this policy, the token is kept alive whenever MantrixFlow uses it. Scheduled
syncs, discovery, preview, and connection tests refresh Salesforce access and
persist the current OAuth token state back to the encrypted saved connection.

Important limits:

- OAuth refresh tokens are not truly permanent. Salesforce can still revoke them
  if an admin revokes app access, the app policy changes, the connected user is
  disabled, the sandbox is refreshed, or the refresh token is idle longer than
  the configured limit.
- If your org only allows idle expiry, schedule the pipeline to run more often
  than the idle window. Example: with a 30-day idle window, run at least weekly.
- After changing the Salesforce token policy, reconnect MantrixFlow once. Old
  rejected refresh tokens do not become valid again.
- For admin-managed integrations that must avoid refresh-token idle expiry, use
  JWT bearer auth with a certificate/private key and a dedicated integration
  user.

## 4. Auth Modes

Salesforce supports these connection configs.

OAuth refresh token, recommended:

```json
{
  "type": "salesforce",
  "auth_mode": "oauth2_refresh",
  "client_id": "consumer-key",
  "client_secret": "consumer-secret",
  "refresh_token": "refresh-token",
  "is_sandbox": false,
  "api_version": "v60.0",
  "connector_phase": 6
}
```

JWT bearer:

```json
{
  "type": "salesforce",
  "auth_mode": "oauth2_jwt",
  "client_id": "consumer-key",
  "username": "integration-user@example.com",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "is_sandbox": false,
  "api_version": "v60.0",
  "connector_phase": 6
}
```

Username, password, security token:

```json
{
  "type": "salesforce",
  "auth_mode": "basic",
  "username": "integration-user@example.com",
  "password": "salesforce-password",
  "security_token": "salesforce-security-token",
  "is_sandbox": false,
  "api_version": "v60.0",
  "connector_phase": 6
}
```

Secrets are encrypted by the Go API before persistence. Sensitive fields include
`refresh_token`, `client_secret`, `private_key`, `password`, `security_token`,
and `access_token`.

## 5. OAuth Start Flow

For a saved Salesforce data source, start OAuth through the Go API:

```http
POST /api/v1/organizations/{organizationId}/salesforce/oauth/start
Authorization: Bearer <supabase-jwt>
Content-Type: application/json
```

```json
{
  "dataSourceId": "salesforce-data-source-id",
  "isSandbox": false,
  "returnPath": "/workspace/connections/salesforce-data-source-id"
}
```

The response contains `authorizeUrl`. Redirect the browser there. Salesforce
returns to the Go API callback, and the Go API stores an encrypted
`oauth2_refresh` connection config on the same data source.

## 6. Discovery and Stream Names

Salesforce discovery is dynamic by default. It calls Salesforce global/object
describe APIs and returns standard and custom objects.

Use canonical stream keys:

```text
salesforce.Account
salesforce.Contact
salesforce.Custom_Object__c
```

DuckDB raw tables use:

```text
salesforce__Account
salesforce__Contact
salesforce__Custom_Object__c
```

Discovery marks:

- Primary key: `Id`
- Incremental candidates: `SystemModstamp`, then `LastModifiedDate`, then other
  date/datetime/integer fields
- Formula fields: `calculated: true`
- Shield fields: `encrypted: true`
- Delete lineage columns: `_sf_is_deleted`, `_sf_deleted_at`
- Sync lineage columns: `_sf_synced_at`, `_sf_formula_refreshed_at`

## 7. Pipeline Setup

1. Create a Salesforce source connection.
2. Test the connection.
3. Discover source schema.
4. Select objects such as `salesforce.Account`.
5. Create a PostgreSQL destination connection.
6. Create the destination table manually in PostgreSQL.
7. Write dbt SQL that selects from DuckDB raw staging and projects into the
   destination shape.
8. Run the pipeline.

Example dbt model:

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

Example destination table:

```sql
create table if not exists public.salesforce_accounts (
  id text primary key,
  name text,
  system_modstamp timestamptz,
  _sf_is_deleted boolean,
  _sf_deleted_at timestamptz,
  _sf_synced_at timestamptz
);
```

## 8. Bulk API and Incremental Sync

Full sync can use Bulk API 2.0 automatically for large objects. The default
threshold is `100000` rows and can be overridden per connection/stream with:

```json
{
  "bulk_threshold_rows": 500000
}
```

Incremental sync uses SOQL REST pagination with the selected replication key.
Deleted records are staged as rows with `_sf_is_deleted = true` and
`_sf_deleted_at`.

Formula refresh can be enabled per stream:

```json
{
  "formula_refresh_interval_hours": 24
}
```

## 9. Real-Time CDC Streamer

The Phase 5 CDC service lives in:

```text
apps/salesforce-streamer
```

Run it separately from the ELT server:

```bash
cd apps/salesforce-streamer
python -m pip install -r requirements.txt
python main.py
```

Required env:

```bash
DATABASE_URL=postgresql://...
SALESFORCE_STREAM_SUBSCRIPTIONS_JSON='[...]'
SALESFORCE_STREAMER_IDLE_FALLBACK_SECONDS=3600
```

The streamer subscribes to Salesforce CDC events and enqueues existing
MantrixFlow `incremental-sync` jobs into pgmq. If CDC goes quiet or the
subscription fails, it enqueues a polling catch-up job so delivery still uses
the normal pipeline path.

Deployment note: the Pub/Sub gRPC adapter expects Salesforce Pub/Sub generated
protobuf bindings to be available in the runtime image.

## 10. Troubleshooting

- `invalid_grant` on sandbox: the sandbox may have been refreshed. Reconnect
  Salesforce.
- `Salesforce access was revoked`: reconnect OAuth.
- Destination table missing: create the PostgreSQL table manually.
- Column mismatch: align dbt SQL output columns with the destination table.
- No custom object visible: check Salesforce object permissions for the
  connected user/app.
- Shield field masked: grant `View Encrypted Data` or exclude the field.
