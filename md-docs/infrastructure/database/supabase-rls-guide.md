# Supabase RLS bootstrap

`sql/supabase_rls.sql` is the canonical browser-access contract for the
`public` schema. The Go server embeds that file and checks it on every startup
after `AutoMigrate`.

## Startup behavior

1. Plain PostgreSQL installations without `auth.uid()` are detected and
   skipped.
2. Supabase installations take a transaction-scoped advisory lock.
3. The server compares the embedded SQL checksum with the marker comment on
   `public.app_user_org_id()`.
4. The SQL is reapplied when the checksum changed, any public table has RLS
   disabled, or `anon` has any public table grant.
5. The transaction fails and the server does not start unless every public
   table has RLS enabled and anonymous table grants are zero.

`APPLY_SUPABASE_RLS_ONCE` is enabled by default in every environment. Set it to
`false` only as an emergency opt-out when the direct database role cannot
perform DDL. The next deployment should restore the required privileges and
re-enable the bootstrap.

Do not enable `FORCE ROW LEVEL SECURITY`. The main server connects as the table
owner and must bypass browser policies for organization-scoped API operations,
workers, billing reconciliation, callbacks, and migrations.

## Exposure contract

- `anon` has no table privileges in `public`.
- `authenticated` has no privileges on public business tables; browser clients
  use Supabase Auth and call the organization-scoped Go API for data.
- `data_source_connections`, normalized child configuration records, billing
  internals, notification read receipts, Slack/Zendesk OAuth state,
  Slack/GitHub secrets, Oria history, run, tool, transfer, evidence, and audit
  tables, Simulation Platform scenarios/runs/evidence/assertions/twins, and
  legacy agent tables remain API-only.
- Pipeline status uses authenticated Go API polling rather than direct Realtime
  table subscriptions.

When adding a public table, add RLS defense in depth and keep it API-only unless
the architecture and this guide are explicitly revised together.

## Permission rollback

The API-only grant migration is non-destructive and has an explicit emergency
rollback at `sql/supabase_browser_grants_rollback.sql`. Revert or replace the Go
release containing the final revoke sweep before applying that SQL; otherwise
the next bootstrap will revoke the restored grants again. The rollback restores
only the former authenticated table privileges, retains RLS, and never changes
customer rows. Re-run the two-organization JWT checks after either direction.

## Verification

Run the non-destructive audit:

```bash
psql "$DATABASE_URL" -X -f sql/schema_cleanup_audit.sql
```

Apply the same bootstrap without starting HTTP workers:

```bash
go run ./cmd/rls-bootstrap
```

The first result set must have no row where `rls_enabled` is false. The second
must show no `anon` grants. Test with two organization JWTs before changing an
organization policy, and verify direct reads of secret tables are denied.

## Schema cleanup policy

Do not put table or column drops in the startup bootstrap. An object can be
removed only after all of the following are true:

1. no production code, SQL, background worker, notification, GitHub YAML, or
   analytics query references it;
2. foreign-key, view, trigger, index, and publication dependencies are known;
3. existing rows are zero or exported and reconciled;
4. a reversible migration has completed one rollback window.

The current legacy pipeline graph/schema columns are not cleanup candidates:
the scheduler, compatibility compiler, analytics service, and migration code
still read them. The legacy agent tables are fail-closed, but are also retained
because `agent_conversations` contains data.
