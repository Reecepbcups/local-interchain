import json
import os
import time

import httpx

from util_base import contracts_json_path, default_contracts_json, parent_dir

# from util_contracts import get_file_hash
from util_req import RequestBase, send_request

BLOCK_TIME = 2


def get_tx_hash(res: str | dict) -> str:
    # String is a JSON blob, but was input as a string
    if isinstance(res, str):
        try:
            res = json.loads(res)
        except:
            pass

    tx_hash = ""
    if isinstance(res, dict):
        tx_hash = res["txhash"]
    else:
        tx_hash = res

    return tx_hash


def query_tx(query_base: RequestBase, res: str | dict) -> dict:
    tx_hash = get_tx_hash(res)
    if len(tx_hash) == 0:
        return dict(error="tx_hash is empty")

    time.sleep(BLOCK_TIME)
    res = send_request(
        query_base,
        f"tx {tx_hash} --output json",
    )
    return dict(tx=res)


def get_chain_start_time_from_logs() -> int:
    logs_path = os.path.join(parent_dir, "configs", "logs.json")

    with open(logs_path, "r") as f:
        logs = dict(json.load(f))

    return int(logs.get("start-time", -1))


def _upload_file(URL: str, chain_id: str, key_name: str, abs_path: str) -> dict:
    print(f"[upload_file] ({chain_id}) {abs_path}")

    data = {
        "chain_id": chain_id,
        "key-name": key_name,
        "file-name": abs_path,
    }

    url = URL
    if url.endswith("/"):
        url += "upload"
    else:
        url += "/upload"

    r = httpx.post(
        url,
        json=data,
        headers={"Content-Type": "application/json"},
        timeout=120,
    )

    if r.status_code != 200:
        return dict(error=r.text)

    return json.loads(r.text.replace("\n", ""))


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
