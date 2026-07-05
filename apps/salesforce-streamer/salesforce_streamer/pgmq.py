from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any

import psycopg2


@dataclass(frozen=True)
class IncrementalSyncJob:
    pipeline_id: str
    organization_id: str
    user_id: str
    trigger_type: str
    run_id: str = ""
    checkpoint: dict[str, Any] | None = None


class PGMQBridge:
    def __init__(self, database_url: str) -> None:
        self.database_url = database_url

    def enqueue_incremental_sync(self, job: IncrementalSyncJob) -> None:
        data = {
            "pipelineId": job.pipeline_id,
            "runId": job.run_id,
            "organizationId": job.organization_id,
            "userId": job.user_id,
            "triggerType": job.trigger_type,
            "checkpoint": job.checkpoint or {},
        }
        payload = {
            "name": "incremental-sync",
            "data": data,
            "retryCount": 0,
            "maxRetries": 5,
        }
        with psycopg2.connect(self.database_url) as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT * FROM pgmq.send(%s, %s::jsonb)", ("incremental_sync", json.dumps(payload)))
