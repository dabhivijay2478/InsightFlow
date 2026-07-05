# Salesforce Streamer

Persistent Salesforce Pub/Sub CDC bridge for MantrixFlow.

This service is intentionally separate from the polling ELT server. It subscribes
to Salesforce Change Data Capture channels, converts CDC events into existing
MantrixFlow pipeline jobs, and writes those jobs to pgmq so the normal Go API ->
Python ELT -> DuckDB -> dbt -> destination flow remains the only delivery path.

## Environment

- `DATABASE_URL`: Postgres URL with access to the `pgmq` extension.
- `SALESFORCE_STREAM_SUBSCRIPTIONS_JSON`: JSON array of subscription configs.
- `SALESFORCE_STREAMER_IDLE_FALLBACK_SECONDS`: seconds before a polling catch-up
  job is enqueued when no CDC event is seen. Defaults to `3600`.

Subscription shape:

```json
[
  {
    "pipelineId": "pipeline-uuid",
    "runId": "",
    "organizationId": "org-uuid",
    "userId": "user-uuid",
    "accessToken": "00D...",
    "instanceUrl": "https://mydomain.my.salesforce.com",
    "tenantId": "00D...",
    "objects": ["Account", "Contact"],
    "replayPreset": "LATEST"
  }
]
```

The actual Pub/Sub API gRPC stub is isolated behind `SalesforcePubSubClient`.
Deployments should provide generated Salesforce Pub/Sub protobuf bindings in the
runtime image. Without those bindings, the service fails loudly instead of
pretending CDC is active.
