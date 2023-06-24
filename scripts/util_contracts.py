import base64
import json

from util_base import contracts_path, current_dir, parent_dir
from util_req import RequestBase, send_request


def get_contract_address(query_base: RequestBase, tx_hash: str) -> str:
    res_json = send_request(query_base, f"tx {tx_hash} --output=json")

    code = int(res_json["code"])
    if code != 0:
        raw = res_json["raw_log"]
        return raw

    contract_addr = ""
    for event in res_json["logs"][0]["events"]:
        for attr in event["attributes"]:
            if attr["key"] == "_contract_address":
                contract_addr = attr["value"]
                break

    return contract_addr


def b64encode(MSG: dict):
    if isinstance(MSG, str):
        MSG = json.loads(MSG)

    return base64.b64encode(remove_spaces(MSG)).decode("utf-8")


def remove_spaces(MSG: dict):
    return json.dumps(MSG, separators=(",", ":")).encode("utf-8")
