# Structured Pipeline Workspace — UI Walkthrough

These steps apply to every pipeline test. Source-specific stream details remain in the per-source guides.

## 1. Create a draft pipeline

1. Open **Workspace → Data Pipelines**.
2. Select **New pipeline**.
3. Enter a name and choose an existing source connection.
4. Select **Create and configure**. The Source tab opens.

The draft can be left and resumed at any time from the pipeline directory.

## 2. Configure the source

1. Test the shared source connection.
2. Discover or refresh the source catalog.
3. Select the streams required by this pipeline.
4. For each selected stream, configure full-table or incremental replication, cursor, primary keys, and connector options.
5. Save the stream configuration.
6. Select a stream to inspect its schema and sample preview.

Replacing the source clears pipeline-specific stream selections and requires an impact confirmation.

## 3. Create and publish transformations

1. Open **Transformations** and select **New transformation**.
2. Add a name, stable output model, one or more selected input streams, and SQL.
3. Save the draft revision.
4. Run validation and preview.
5. Review affected destinations and publish the revision.

Draft edits do not affect runs until publication.

## 4. Add destinations

1. Open **Destinations** and select **Add destination**.
2. Choose a tested destination connection.
3. Select one or more published transformations.
4. Configure `schema.table`, upsert keys, and enabled state for each assignment.
5. Test the connection, review the assignments, and save.

SQL remains owned by the reusable transformation and is never duplicated in a destination.

## 5. Configure scheduling and activation

1. Open **General**.
2. Configure an hourly interval from 1–24 hours or a standard five-field cron expression.
3. Save the UTC schedule.
4. Validate the pipeline.
5. Activate it only after a source, streams, published transformation, tested destination, and assignment are ready.

## 6. Run and inspect

1. Use **Run now** for all enabled destinations, or run one destination from its row.
2. Open **Runs** to search and paginate executions.
3. Select a run to inspect its destination, published revision IDs, rows, duration, and errors.
4. Verify destination checkpoints advance independently.

## 7. Verify GitHub synchronization

1. Open **GitHub** and connect a repository.
2. Generate the relationship-based YAML preview.
3. Confirm source streams, reusable transformations, destinations, and assignments are separate sections.
4. Push and pull the definition and inspect synchronization history.
5. Verify unknown transformation references, duplicate keys, and conflicting schedules are rejected.

## Responsive checks

- At 390px, root tabs scroll horizontally and configuration editors use full pages.
- Transformation, destination, and run tables render as compact card rows.
- Source catalogs and record previews remain contained and horizontally scrollable.
- All controls have visible keyboard focus and status is conveyed by text and icon as well as color.
