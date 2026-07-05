from __future__ import annotations

import logging

from salesforce_streamer.config import load_settings
from salesforce_streamer.service import SalesforceStreamer


logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")


def main() -> None:
    settings = load_settings()
    SalesforceStreamer(settings).run_forever()


if __name__ == "__main__":
    main()
