# Local Manual Testing Guide

This is the primary manual UI testing guide for the current MANTrixFlow ELT
stack:

- App: `apps/app`
- Go API: `apps/server/main-server`
- ELT server: `apps/server/elt-server`

Use this guide when you want to verify the source-to-destination pipeline flow
through the product UI.

For the full architecture guide, see
[source-to-destination-elt-flow.md](./source-to-destination-elt-flow.md).

## Prerequisites

Before testing, make sure you have:

- a working source connection you can use in the builder
- at least one destination connection
- local env files configured for the app, Go API, and ELT server
- a running database and Supabase setup if your local stack depends on them

Recommended local ports:

- App: `http://localhost:3000`
- Go API: `http://localhost:5000`
- ELT server: `http://localhost:8000`

## Start the local stack

### 1. Start the app

```bash
cd apps/app
bun install
bun run dev
```

Expected result:

- the app starts on `http://localhost:3000`
- the workspace loads without build errors

### 2. Start the Go API

```bash
cd apps/server/main-server
go run ./cmd/server
```

Expected result:

- the API boots without migration or config errors
- `GET http://localhost:5000/api/v1/health` succeeds

### 3. Start the ELT server

```bash
cd apps/server/elt-server
source .venv/bin/activate
python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --workers 1 --loop asyncio
```

Expected result:

- the ELT server boots without import or dependency errors
- `GET http://localhost:8000/health` succeeds

## Step-by-step manual UI test flow

### 1. Open the app and create or open a pipeline

Open `http://localhost:3000` and navigate to the pipelines builder.

Create a new pipeline or open an existing one you can safely edit.

Verify:

- the builder loads
- the source node is visible
- at least one destination can be added

### 2. Connect or confirm the source

Open the source drawer from the source node.

Confirm:

- the expected source connection is selected
- the source connection name and type are correct

Optional check:

- click `Test connection`

Verify:

- loading appears inside the button
- success is shown by toast and button state

### 3. Refresh discovery metadata

In the source drawer, click `Refresh tables`.

Verify:

- the button shows a loading state
- the source inventory refreshes
- the source node summary updates with available schema/table counts

### 4. Select multiple source tables

In the source drawer:

- include more than one source table
- choose one table as the active preview table

Verify:

- the included table list stays selected
- the active preview table updates the raw preview
- save/reload should keep the selected tables and preview table coherent

### 5. Confirm raw source preview

Use the source preview area to inspect sample rows for the active preview table.

Verify:

- preview rows load successfully
- columns match the selected preview table
- switching the preview table updates the preview instead of breaking it

### 6. Add the first destination

Use the source node `+` action to add a destination node.

Open the destination drawer.

Configure:

- destination connection
- destination schema
- write mode

Verify:

- the destination node appears directly connected to the source
- the connection is a plain source-to-destination edge

### 7. Add a second destination

Use the source node `+` action again and create a second destination.

Verify:

- the second destination connects directly to the same source
- both destinations remain independently editable
- deleting one destination does not remove the other

### 8. Configure sync mode and replication key

For destination A:

- choose `FULL_TABLE`

For destination B:

- choose `INCREMENTAL`
- enter a manual replication key

Verify:

- sync settings are destination-owned, not source-owned
- each destination keeps its own saved values
- reloading the builder preserves the values on the correct destination

### 9. Add destination normalisation rules

Open the `Normalisation` tab for a destination and add structural rules such as:

- a `cast`
- a `rename`

Verify:

- rules save correctly
- rules are attached only to that destination
- another destination can have different rules for the same source table

### 10. Write UI SQL dbt models

In the destination dbt editor:

- create at least one SQL model per required source table
- set output table names
- use `{{ source('raw', 'table_name') }}` references

Verify:

- the SQL editors save successfully
- output table names persist after reload
- the destination summary reflects model outputs

### 11. Validate SQL

Use `Validate SQL` for at least one model.

Verify:

- valid SQL returns output schema
- invalid SQL returns the exact validation error
- validation does not require a live source query

### 12. Preview destination model output

Use the destination preview tab and preview a selected model.

Verify:

- preview loads rows and columns
- destination normalisation is reflected in the output
- empty preview results do not crash the UI

### 13. Save the pipeline

Save the pipeline and refresh the page.

Verify after reload:

- selected source tables persist
- the source preview table remains coherent
- both destinations still exist
- destination sync mode and replication key persist
- normalisation rules persist
- dbt SQL models persist

### 14. Run a destination

Trigger a run from a destination node.

Verify:

- the run starts from the destination, not the source
- source node does not expose a run control
- the run drawer opens and shows progress

### 15. Review run details

Wait for the run to complete or reach a stable intermediate state.

Verify the run details show:

- `Stage`
- `dbt Transform`
- `Deliver`

Also verify:

- destination outputs are listed
- the run is associated with the destination you triggered

## Negative and pathology checks

Run these checks before closing the test cycle.

### Incremental destination without replication key

Try to save or run an `INCREMENTAL` destination without a replication key.

Expected result:

- validation blocks the action
- the UI points back to the destination configuration

### Invalid SQL

Enter broken SQL in a model and run `Validate SQL`.

Expected result:

- validation fails
- the exact DuckDB error is surfaced

### Empty preview

Preview a model that returns no rows.

Expected result:

- preview succeeds with empty rows
- the UI remains stable
- no hard failure screen appears

### Multiple destinations

Create at least two destinations from one source and run them separately.

Expected result:

- each destination remains independently runnable
- destination-specific sync settings remain isolated

## Optional backend spot checks

Use these checks when you want to confirm the UI is backed by healthy services.

### Health endpoints

- `GET http://localhost:5000/api/v1/health`
- `GET http://localhost:8000/health`

Expected result:

- both succeed while the stack is running

### Preview and validate-SQL proxy sanity

When UI preview or validate-SQL works, it confirms:

- the app can call the Go API
- the Go API can proxy to the ELT server
- destination-node context is being resolved correctly

### Callback and run metadata sanity

After a run finishes, verify the run details reflect:

- stage status
- dbt status
- delivery status

If you inspect backend data directly, the run should map back to the triggered
destination node and store ELT metadata in `pipeline_runs.run_metadata`.
