# SQLite Destination — Manual Testing Guide

SQLite is a file-based destination. The ELT server writes directly to a local `.db` file. No server process required.

---

## Credential Setup

```json
{
  "database": "/absolute/path/to/analytics.db"
}
```

> ⚠️ The ELT server process must have **read + write** permissions on the `.db` file and its parent directory.  
> The file is created automatically if it does not exist (unlike other destinations where the table must pre-exist).

---

## Destination Table Pre-creation

Even though SQLite creates the `.db` file automatically, the **destination table must still be pre-created**. Use the `sqlite3` CLI:

```bash
sqlite3 /absolute/path/to/analytics.db
```

```sql
-- Create analytics schema equivalent (SQLite uses one namespace)
CREATE TABLE IF NOT EXISTS stripe_customers_hd (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    name     TEXT,
    currency TEXT,
    created  INTEGER
);
```

> ℹ️ SQLite has no schemas. Destination table names are `tablename` without `schema.`. Set the destination as `main.tablename` in the MantrixFlow UI (SQLite always uses `main` as the default schema).

---

## Scenario D-SQ-1 — Basic Delivery: merge mode (with PK)

**Pre-create table:**
```sql
CREATE TABLE IF NOT EXISTS stripe_customers_hd (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    name     TEXT,
    currency TEXT,
    created  INTEGER
);
```

**Pipeline destination:** `main.stripe_customers_hd`

**Expected:** Upsert on `id`; no duplicates on second run.

**Verify:**
```bash
sqlite3 /path/to/analytics.db "SELECT COUNT(*) FROM stripe_customers_hd;"
```

---

## Scenario D-SQ-2 — Append Mode (no PK)

```sql
CREATE TABLE IF NOT EXISTS daily_revenue (
    currency      TEXT,
    charge_count  INTEGER,
    total_cents   INTEGER
);
```

**Expected:** Rows appended each run. `no_pk_warnings` in callback.

---

## Scenario D-SQ-3 — TEXT Affinity (SQLite type system)

SQLite uses dynamic type affinity. All numeric values stored in a `TEXT` column remain as text. Test type handling:

**Pre-create table with REAL column:**
```sql
CREATE TABLE IF NOT EXISTS orders_typed (
    id      TEXT PRIMARY KEY,
    amount  REAL,
    status  TEXT,
    placed_at TEXT
);
```

**Verify:**
```bash
sqlite3 /path/to/analytics.db "SELECT typeof(amount), amount FROM orders_typed LIMIT 3;"
# Expected: real|29.99
```

---

## Scenario D-SQ-4 — Large TEXT (SQLite has no size limit on TEXT)

SQLite `TEXT` columns have no practical size limit. Long text from GitHub commit messages or HubSpot notes will be stored fully.

**Pre-create table:**
```sql
CREATE TABLE IF NOT EXISTS github_commits_hd (
    sha       TEXT PRIMARY KEY,
    message   TEXT,
    author    TEXT,
    committed_at TEXT
);
```

---

## Scenario D-SQ-5 — Column Mismatch Error

**Pre-create table with a required column absent from dbt output:**
```sql
CREATE TABLE IF NOT EXISTS orders_strict (
    id      TEXT PRIMARY KEY,
    amount  REAL,
    missing_col TEXT NOT NULL
);
```

**dbt SQL** (no `missing_col`):
```sql
SELECT id, amount FROM {{ source('raw', 'public__orders') }}
```

**Expected:** Phase 0 fails with column mismatch error.

---

## Scenario D-SQ-6 — Missing Table Error

Do NOT create the table. Set destination to `main.nonexistent_table`.  
**Expected:** Phase 0 fails: table does not exist.

---

## Scenario D-SQ-7 — File Lock Contention

1. Open the SQLite file in another process (`sqlite3 /path/to/analytics.db`).
2. Run a pipeline that delivers to the same file.

**Expected:** Phase 3 delivery waits for lock or fails with `database is locked` error. Run status: `failed`.

---

## SQLite-Specific Notes

| Consideration | Detail |
|--------------|--------|
| Schema support | Only `main` schema; use `main.<table>` in destination panel |
| Type affinity | INTEGER, REAL, TEXT, BLOB, NUMERIC — SQLite ignores declared type at write time |
| Concurrency | Single writer at a time (WAL mode allows one writer + multiple readers) |
| Max file size | 281 TB (practical limit is disk space) |
| No `ALTER TABLE ADD COLUMN NOT NULL` | Adding new NOT NULL columns to existing tables requires recreation |

---

## Destination Verification Commands

```bash
# Row count
sqlite3 /path/to/analytics.db "SELECT COUNT(*) FROM <dest_table>;"

# Schema
sqlite3 /path/to/analytics.db ".schema <dest_table>"

# Duplicate check (merge mode)
sqlite3 /path/to/analytics.db \
  "SELECT id, COUNT(*) cnt FROM <dest_table> GROUP BY id HAVING cnt > 1;"

# All tables (ensure no _dlt_ tables)
sqlite3 /path/to/analytics.db ".tables"
```
