# Slack Pipeline E2E Manual Test

Last verified: 2026-05-19

This guide tests the Slack-native pipeline creation flow from Slack to a real
pipeline run. It uses the saved Neon source and RDS Postgres destination in the
local development org.

## Test Target

- Organization: `1a582ff0-030c-4fdc-b7a1-31cfd432829d`
- Slack workspace: `T04ACTA20LC`
- Slack command channel: `#random` / `C04AP371741`
- Source connection: `Neon Source` / `0a387af5-89e8-457a-a8a3-9a8dfb8e1b4b`
- Destination connection: `RDS PostgresSQL` / `6226f5a7-8037-401f-8620-26f32f5364de`
- Source stream: `public.orders`
- Destination table: `analytics.orders`

## 1. Start Local Services

Start Go:

```bash
cd apps/server/main-server
LOG_LEVEL=debug go run ./cmd/server
```

Start Next:

```bash
cd apps/app
bun dev
```

Expose the frontend port:

```bash
ngrok http 3000
```

Slack must use the public ngrok frontend host, not the Go API host.

## 2. Slack Dashboard URLs

Use the current ngrok host everywhere below. For the current local setup:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
```

Slash commands:

```text
/pipeline    -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
/connection  -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
/mantrixflow -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
```

Interactivity:

```text
Request URL:
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/actions

Options Load URL:
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/options
```

The Options Load URL is required for the source and destination dropdowns in
the Slack pipeline modal. If the modal shows "No result", re-save this URL.

Event subscriptions:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/events
```

OAuth redirect:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/oauth/callback
```

## 3. Prepare Source Table

The Neon source table should exist as `public.orders`.

Expected source columns:

```text
id uuid
user_id uuid
amount numeric(10,2)
status text
metadata jsonb
placed_at timestamptz
```

Optional seed SQL for Neon:

```sql
CREATE TABLE IF NOT EXISTS public.orders (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  amount NUMERIC(10,2) NOT NULL,
  status TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  placed_at TIMESTAMPTZ NOT NULL
);

INSERT INTO public.orders (id, user_id, amount, status, metadata, placed_at)
VALUES
  (
    '11111111-1111-4111-8111-111111111111',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
    125.50,
    'paid',
    '{"notes":"seed","channel":"web","region":"apac"}'::jsonb,
    '2026-05-01T10:00:00Z'
  ),
  (
    '22222222-2222-4222-8222-222222222222',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
    89.25,
    'shipped',
    '{"notes":"seed","channel":"mobile","region":"india"}'::jsonb,
    '2026-05-02T11:00:00Z'
  ),
  (
    '33333333-3333-4333-8333-333333333333',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
    240.00,
    'delivered',
    '{"notes":"seed","channel":"partner","region":"us"}'::jsonb,
    '2026-05-03T12:00:00Z'
  )
ON CONFLICT (id) DO UPDATE
SET
  user_id = EXCLUDED.user_id,
  amount = EXCLUDED.amount,
  status = EXCLUDED.status,
  metadata = EXCLUDED.metadata,
  placed_at = EXCLUDED.placed_at;
```

## 4. Prepare Destination Table

The runner intentionally does not create destination tables. Create the table
before running the pipeline.

Run this in the RDS Postgres destination:

```sql
CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.orders (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  amount NUMERIC(10,2) NOT NULL,
  status TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  placed_at TIMESTAMPTZ NOT NULL
);
```

This fixes the expected strict ELT error:

```text
Destination table analytics.orders does not exist.
Create the table in your destination database before running this pipeline.
```

If the Slack-created pipeline already exists and failed with that message, do
not recreate it. Create the destination table above, then rerun the existing
pipeline from Slack:

```text
/pipeline run slack-neon-orders-to-rds
```

Current local recovery note: `analytics.orders` has been created in the saved
`RDS PostgresSQL` destination for this test target.

## 5. Confirm Slack Connection Discovery

In Slack `#random`, run:

```text
/connection list
```

Expected result includes:

```text
Neon Source ... active ... postgres
RDS PostgresSQL ... active ... destination ... postgres
```

If source or destination dropdowns show no results in the modal:

1. Confirm the command is in `#random`.
2. Re-save the Slack Interactivity Options Load URL.
3. Confirm Go logs show `handleSlackBlockSuggestion: received request`.

## 6. Create Pipeline From Slack

In Slack `#random`, run:

```text
/pipeline create
```

Fill the modal:

```text
Name:
slack-neon-orders-to-rds

Source connection:
Neon Source

Source streams:
[{"stream_key":"public.orders","replication_method":"FULL_TABLE"}]

Destination connection:
RDS PostgresSQL

Destination schema:
analytics

Destination table map JSON:
{"public.orders":"analytics.orders"}

Write mode:
Replace

Default replication:
Full table

Default replication key:
leave blank

Schedule:
None

Run now:
unchecked for first create, optional after manual validation
```

Submit the modal.

Expected Slack response:

```text
Pipeline slack-neon-orders-to-rds created.
```

## 7. Validate And Run

List pipelines:

```text
/pipeline list
```

Check status:

```text
/pipeline status slack-neon-orders-to-rds
```

Run:

```text
/pipeline run slack-neon-orders-to-rds
```

Click the confirmation button.

Expected lifecycle messages:

```text
Pipeline slack-neon-orders-to-rds started
Pipeline slack-neon-orders-to-rds completed
```

If this is a rerun after the missing-table failure, the same pipeline ID can be
used. The important precondition is that `analytics.orders` exists before the
run starts.

## 8. Verify Destination Rows

Run this in RDS Postgres:

```sql
SELECT COUNT(*) FROM analytics.orders;

SELECT id, user_id, amount, status, metadata, placed_at
FROM analytics.orders
ORDER BY placed_at;
```

Expected count: `3`.

## Troubleshooting

### Modal Shows No Source Or Destination Results

Cause: Slack is not calling the Options Load URL.

Fix:

```text
Interactivity & Shortcuts -> Options Load URL
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/options
```

Save the page, then reopen the modal.

### Command Says To Use Another Channel

The current install is bound to `#random` / `C04AP371741`. Run the command from
that channel, or reconnect Slack and choose a different alert channel.

### Pipeline Fails Because Destination Table Does Not Exist

Create `analytics.orders` with the DDL in step 4, then rerun:

```text
/pipeline run slack-neon-orders-to-rds
```

### Pipeline Fails On `_dlt_load_id`

If a previous delivery partially inserted rows, dlt can fail while adding its
temporary row metadata:

```text
column "_dlt_load_id" of relation "orders" contains null values
```

For this E2E test, reset the destination table and dlt staging artifacts, then
rerun the Slack pipeline:

```sql
TRUNCATE TABLE analytics.orders;

ALTER TABLE analytics.orders DROP COLUMN IF EXISTS _dlt_id;
ALTER TABLE analytics.orders DROP COLUMN IF EXISTS _dlt_load_id;
ALTER TABLE analytics.orders DROP COLUMN IF EXISTS _dlt_root_id;
ALTER TABLE analytics.orders DROP COLUMN IF EXISTS _dlt_parent_id;
ALTER TABLE analytics.orders DROP COLUMN IF EXISTS _dlt_list_idx;

DROP TABLE IF EXISTS analytics._dlt_loads CASCADE;
DROP TABLE IF EXISTS analytics._dlt_version CASCADE;
DROP TABLE IF EXISTS analytics._dlt_pipeline_state CASCADE;
DROP SCHEMA IF EXISTS analytics_staging CASCADE;
```

Then rerun:

```text
/pipeline run slack-neon-orders-to-rds
```

Current local recovery note: `analytics.orders` has been reset to `0` rows in
the saved `RDS PostgresSQL` destination.

### Run Waits With `ELT staging disk busy`

The Go dispatcher checks free disk before it sends a run to the Python ELT
worker. By default it requires `2x` the organization's staging limit. A `pro`
org has a 20GB staging limit, so local dev needs about 40GB free:

```text
elt staging disk busy: available_gb=39.90 required_gb=20 minimum_available_gb=40.00 headroom_multiplier=2.00
```

For local Slack E2E testing, either free enough disk space or temporarily lower
the local dispatch headroom:

```bash
cd apps/server/main-server
ELT_STAGING_DISPATCH_HEADROOM_MULTIPLIER=1 LOG_LEVEL=debug go run ./cmd/server
```

If you clicked Run multiple times while the pipeline was waiting, cancel the
extra waiting attempt from Slack before rerunning:

```text
/pipeline cancel slack-neon-orders-to-rds
```

Do not set the multiplier below `2` in production unless the staging capacity
policy has been reviewed.

### OAuth Redirect Goes To Login

Use the localhost settings page to start OAuth, but keep the Slack callback on
ngrok. The Go env should have:

```text
APP_WEB_URL=http://localhost:3000
SLACK_OAUTH_REDIRECT_BASE_URL=https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
```

Restart Go and Next after env changes.
