# architecture
- MANTrixFlow uses DuckDB-staged ELT: source → dlt → ephemeral DuckDB → dbt-duckdb transforms → delivery to client destination. No raw _dlt_* tables land in client DB. Confidence: 0.90
- dbt models are authored exclusively via inline UI SQL editor (Monaco). No GitHub dbt projects, no Meltano. Confidence: 0.85
- Sync mode and replication key belong to the destination node, not the source node. Source panel owns only stream selection and connection. Confidence: 0.80
- Pipeline builder model: one source node feeds multiple plain destination nodes via direct edges. No branch/groups concept. Plus icon on source adds a destination. Confidence: 0.75
- All schema.table references use a single string input format (e.g., "public.users"), not separate schema and table fields. Parse by splitting on first dot. Confidence: 0.65
- Destination delivery writes into existing tables only. Never create new tables in client destinations. If destination table does not exist, fail with clear error. Confidence: 0.65
