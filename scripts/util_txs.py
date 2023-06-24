import json
import os
import time

import httpx

from helpers.transactions import RequestBuilder, send_request
from util_base import contracts_json_path, default_contracts_json, parent_dir

# from util_contracts import get_file_hash

BLOCK_TIME = 2


def get_chain_start_time_from_logs() -> int:
    logs_path = os.path.join(parent_dir, "configs", "logs.json")

    with open(logs_path, "r") as f:
        logs = dict(json.load(f))

    return int(logs.get("start-time", -1))





def get_cache_or_default(contracts: dict, ictest_chain_start: int) -> dict:
    with open(contracts_json_path, "r") as f:
        cache_time = dict(json.load(f)).get("start_time", 0)

    if cache_time == 0 or cache_time != ictest_chain_start:
        # reset cache, and set cache time to current ictest time
        contracts["start_time"] = ictest_chain_start
        contracts["file_cache"] = {}

        # write to file
        with open(contracts_json_path, "w") as f:
            json.dump(contracts, f, indent=4)

    return contracts
