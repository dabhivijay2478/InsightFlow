from __future__ import annotations

import json
import os
from dataclasses import dataclass
from typing import Any

from dotenv import load_dotenv


@dataclass(frozen=True)
class StreamSubscription:
    pipeline_id: str
    organization_id: str
    user_id: str
    access_token: str
    instance_url: str
    tenant_id: str
    objects: list[str]
    run_id: str = ""
    replay_preset: str = "LATEST"


@dataclass(frozen=True)
class Settings:
    database_url: str
    subscriptions: list[StreamSubscription]
    idle_fallback_seconds: int = 3600


def load_settings() -> Settings:
    load_dotenv()
    database_url = os.getenv("DATABASE_URL", "").strip()
    if not database_url:
        raise RuntimeError("DATABASE_URL is required")
    raw_subscriptions = os.getenv("SALESFORCE_STREAM_SUBSCRIPTIONS_JSON", "[]")
    try:
        parsed = json.loads(raw_subscriptions)
    except json.JSONDecodeError as exc:
        raise RuntimeError("SALESFORCE_STREAM_SUBSCRIPTIONS_JSON must be valid JSON") from exc
    if not isinstance(parsed, list):
        raise RuntimeError("SALESFORCE_STREAM_SUBSCRIPTIONS_JSON must be a JSON array")
    subscriptions = [_subscription_from_dict(item) for item in parsed]
    idle = int(os.getenv("SALESFORCE_STREAMER_IDLE_FALLBACK_SECONDS", "3600"))
    return Settings(database_url=database_url, subscriptions=subscriptions, idle_fallback_seconds=max(idle, 60))


def _subscription_from_dict(raw: Any) -> StreamSubscription:
    if not isinstance(raw, dict):
        raise RuntimeError("Each Salesforce stream subscription must be an object")
    objects = raw.get("objects") or []
    if not isinstance(objects, list) or not all(isinstance(item, str) and item.strip() for item in objects):
        raise RuntimeError("Salesforce stream subscription objects must be a non-empty string array")
    return StreamSubscription(
        pipeline_id=str(raw.get("pipelineId") or "").strip(),
        run_id=str(raw.get("runId") or "").strip(),
        organization_id=str(raw.get("organizationId") or "").strip(),
        user_id=str(raw.get("userId") or "").strip(),
        access_token=str(raw.get("accessToken") or "").strip(),
        instance_url=str(raw.get("instanceUrl") or "").strip().rstrip("/"),
        tenant_id=str(raw.get("tenantId") or "").strip(),
        objects=[item.strip() for item in objects],
        replay_preset=str(raw.get("replayPreset") or "LATEST").strip(),
    )
