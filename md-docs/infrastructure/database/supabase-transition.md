# Supabase transition plan

## Current phase

Supabase PostgreSQL, PGMQ, Row Level Security, Auth/JWKS, admin user operations,
and Storage integrations remain unchanged. The application connects through a
normal PostgreSQL `DATABASE_URL`; new control-plane storage contracts use only
portable PostgreSQL/GORM operations. Supabase-specific Auth and bootstrap code
stays in explicit adapters.

The future database Compose files are cold preparation. They must not be
started or pointed at production in this phase.

## Future phase A: database

1. Validate extension/version compatibility, especially PGMQ.
2. Provision PostgreSQL, PgBouncer, TLS, private networking, WAL-G, and restore
   drills.
3. Use `pg_dump`/`pg_restore` for a controlled window or logical replication
   for lower downtime.
4. Apply expand-only migrations and reconcile sequences, owners, privileges,
   RLS policies, queue schemas, and checksums.
5. Run shadow reads and production smoke checks.
6. Switch only `DATABASE_URL`/`DIRECT_URL`, preserving a tested rollback window.

## Future phase B: verification

Verify organizations, memberships, pipeline versions, schedules, runs,
checkpoints, simulation metadata, evidence metadata, and PGMQ delivery/recovery.
Do not contract old schema until both application versions and rollback windows
have expired.

## Future phase C: Auth

Migrate Supabase Auth separately. The database move must not alter JWT behavior,
issuer/audience validation, user IDs, invitations, or service-role operations.
Replace those adapters only after the database cutover is stable.
