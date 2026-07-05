from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable

from .config import StreamSubscription


@dataclass(frozen=True)
class SalesforceCDCEvent:
    object_name: str
    replay_id: bytes
    record_ids: list[str]
    change_type: str
    commit_timestamp: int | None = None


class SalesforcePubSubClient:
    """Thin adapter for Salesforce Pub/Sub API generated gRPC bindings."""

    def subscribe(self, subscription: StreamSubscription) -> Iterable[SalesforceCDCEvent]:
        try:
            import grpc  # noqa: F401
            from salesforce_pubsub import pubsub_api_pb2, pubsub_api_pb2_grpc  # type: ignore
        except Exception as exc:
            raise RuntimeError(
                "Salesforce Pub/Sub protobuf bindings are not installed. "
                "Generate them from Salesforce's Pub/Sub API proto and make "
                "`salesforce_pubsub.pubsub_api_pb2(_grpc)` importable."
            ) from exc

        raise NotImplementedError(
            "Wire SalesforcePubSubClient.subscribe to pubsub_api_pb2_grpc.PubSubStub "
            "for this deployment's generated protobuf package."
        )
