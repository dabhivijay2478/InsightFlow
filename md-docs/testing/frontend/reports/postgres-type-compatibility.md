# PostgreSQL Type and Scale Compatibility

Two retained Chrome UI pipelines validated broad type fidelity and incremental scale.

## Full-type pipeline

Pipeline: `Postgres Full Sync` (`01132030-99a5-4885-b777-f1e4231daf73`)

Six selected tables were extracted. The broad type table exercised 30 explicitly declared PostgreSQL type columns plus its serial primary key:

| Type family | Validated types |
| :--- | :--- |
| Integers | serial primary key, `smallint`, `integer`, `bigint` |
| Exact and floating point | `numeric`, `real`, `double precision` |
| Boolean and bit | `boolean`, `bit`, `bit varying` |
| Text | `character`, `character varying`, `text` |
| Binary | `bytea` |
| Date/time | `date`, `time without time zone`, `time with time zone`, `timestamp without time zone`, `timestamp with time zone`, `interval` |
| Network | `inet`, `cidr`, `macaddr` |
| Structured | `json`, `jsonb` |
| Identifier | `uuid` |
| Arrays | `integer[]`, `bigint[]`, `text[]`, `uuid[]`, `jsonb[]` |

The latest run succeeded, extracted all six tables, and retained the three representative rows produced by the broad type model.

## Incremental scale pipeline

Pipeline: `Postgres Incremental Sync` (`67c49a7b-f564-4fa4-a352-55d1bc799519`)

- Stream: `large_dataset`
- Cursor: `updated_at`
- Upsert key: `id`
- Initial full load: 8,000 rows
- Delta: 2,000 inserts plus 25 updates
- Incremental run: 2,025 rows
- Final destination count: 10,000 rows
- Verification: 2,000 IDs above 8,000 and 25 expected updated IDs

The source fixtures, destination rows, pipelines, connections, and run history were not deleted.
