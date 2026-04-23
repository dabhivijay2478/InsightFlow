# Pipeline 44 — SQLite Source → SQLite Destination

**Source streams:** 3 | **Destination:** SQLite (different file)

> dbt SQL identical to `41-sqlite-to-postgres.md`.

---

## Connections
```json
{ "database": "/absolute/path/to/source.db" }
{ "database": "/absolute/path/to/analytics.db" }
```

> ⚠️ Source and destination **must be different files**. Same file will cause lock contention.

---

## Destination DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS task_board (
    task_id    INTEGER PRIMARY KEY,
    task_title TEXT, task_status TEXT, urgency TEXT,
    due_on     TEXT, created_on TEXT, updated_on TEXT
);
CREATE TABLE IF NOT EXISTS task_notes (
    note_id  INTEGER PRIMARY KEY,
    task_ref INTEGER, note_body TEXT, added_on TEXT
);
CREATE TABLE IF NOT EXISTS tag_master (
    tag_id INTEGER PRIMARY KEY, tag_name TEXT, hex_color TEXT
);
```

---

## Verify
```bash
SRC=/absolute/path/to/source.db
DEST=/absolute/path/to/analytics.db

sqlite3 $DEST "SELECT urgency FROM task_board GROUP BY urgency;"
sqlite3 $DEST "SELECT created_on FROM task_board LIMIT 3;"   # ISO string
sqlite3 $DEST "SELECT hex_color FROM tag_master LIMIT 5;"   # starts with #

# Confirm no row duplication
sqlite3 $DEST "SELECT task_id, COUNT(*) cnt FROM task_board GROUP BY task_id HAVING cnt>1;"
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| Source and dest same path | Phase 3 fails: `database is locked` |
| Dest file does not exist | SQLite auto-creates it — OK |
| Dest file not writable (permissions) | Phase 3 fails: `unable to open database file` |
