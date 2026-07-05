from __future__ import annotations

import logging
import threading
import time
from dataclasses import dataclass

from .config import Settings, StreamSubscription
from .pgmq import IncrementalSyncJob, PGMQBridge
from .pubsub import SalesforceCDCEvent, SalesforcePubSubClient

logger = logging.getLogger("salesforce-streamer")


@dataclass
class SubscriptionHealth:
    last_event_at: float
    last_fallback_at: float


class SalesforceStreamer:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self.bridge = PGMQBridge(settings.database_url)
        self.pubsub = SalesforcePubSubClient()
        self.health: dict[str, SubscriptionHealth] = {}

    def run_forever(self) -> None:
        threads = []
        for subscription in self.settings.subscriptions:
            key = self._key(subscription)
            self.health[key] = SubscriptionHealth(last_event_at=time.time(), last_fallback_at=0)
            thread = threading.Thread(target=self._run_subscription, args=(subscription,), daemon=True)
            thread.start()
            threads.append(thread)
        while True:
            self._enqueue_idle_fallbacks()
            time.sleep(30)

    def _run_subscription(self, subscription: StreamSubscription) -> None:
        key = self._key(subscription)
        while True:
            try:
                for event in self.pubsub.subscribe(subscription):
                    self.health[key] = SubscriptionHealth(
                        last_event_at=time.time(),
                        last_fallback_at=self.health[key].last_fallback_at,
                    )
                    self._enqueue_event(subscription, event)
            except Exception as exc:
                logger.error("Salesforce Pub/Sub subscription failed for %s: %s", key, exc)
                self._enqueue_fallback(subscription, reason="cdc_subscription_error")
                time.sleep(30)

    def _enqueue_event(self, subscription: StreamSubscription, event: SalesforceCDCEvent) -> None:
        checkpoint = {
            "salesforceCdc": {
                "object": event.object_name,
                "replayId": event.replay_id.hex(),
                "recordIds": event.record_ids,
                "changeType": event.change_type,
                "commitTimestamp": event.commit_timestamp,
            }
        }
        self.bridge.enqueue_incremental_sync(
            IncrementalSyncJob(
                pipeline_id=subscription.pipeline_id,
                run_id=subscription.run_id,
                organization_id=subscription.organization_id,
                user_id=subscription.user_id,
                trigger_type="salesforce_cdc",
                checkpoint=checkpoint,
            )
        )

    def _enqueue_idle_fallbacks(self) -> None:
        now = time.time()
        for subscription in self.settings.subscriptions:
            key = self._key(subscription)
            health = self.health.get(key)
            if health is None:
                continue
            if now - health.last_event_at < self.settings.idle_fallback_seconds:
                continue
            if now - health.last_fallback_at < self.settings.idle_fallback_seconds:
                continue
            self._enqueue_fallback(subscription, reason="cdc_idle_polling_catchup")
            self.health[key] = SubscriptionHealth(last_event_at=health.last_event_at, last_fallback_at=now)

    def _enqueue_fallback(self, subscription: StreamSubscription, *, reason: str) -> None:
        logger.info("Enqueueing Salesforce polling fallback for %s (%s)", self._key(subscription), reason)
        self.bridge.enqueue_incremental_sync(
            IncrementalSyncJob(
                pipeline_id=subscription.pipeline_id,
                run_id=subscription.run_id,
                organization_id=subscription.organization_id,
                user_id=subscription.user_id,
                trigger_type=reason,
                checkpoint={"salesforceCdcFallback": {"objects": subscription.objects, "reason": reason}},
            )
        )

    @staticmethod
    def _key(subscription: StreamSubscription) -> str:
        return f"{subscription.organization_id}:{subscription.pipeline_id}:{','.join(subscription.objects)}"
