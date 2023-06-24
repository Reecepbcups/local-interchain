import base64
import hashlib
import json
import os
import time

from httpx import get
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


def instantiate_contract(
    # bin_base: RequestBase, query_base: RequestBase, msg: str
    bin_base: RequestBase,
    query_base: RequestBase,
    codeId: int | str,
    msg: str,
    label: str,
    flags: str = "",
) -> str:
    # tx = f" {cw4CodeId} {init} --label=cw4 --from {KEY_NAME} --no-admin --home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json"
    txHash: str | None

    if "--output=json" not in flags:
        msg += " --output=json"

    res = send_request(
        bin_base, f"tx wasm instantiate {codeId} {msg} --label={label} {flags}"
    )
    time.sleep(2.5)

    if isinstance(res, dict):
        resVal = res.get("txhash")
        txHash = resVal if resVal is not None else res.get("raw_log")
    else:
        txHash = json.loads(res)["txhash"]

    if txHash is None:
        raise Exception("No txHash found", res)

    address = get_contract_address(query_base, txHash)
    print(f"[instantiate_contract] {address=}\n")
    return address


def execute_contract(
    bin_base: RequestBase,
    contract_addr: str,
    msg: str,
    flags: str = "",
) -> str:
    txHash: str | None

    if "--output=json" not in flags:
        msg += " --output=json"

    cmd = f"tx wasm execute {contract_addr} {msg} {flags}"
    print("[execute_contract]", cmd)

    res = send_request(bin_base, cmd)

    if isinstance(res, dict):
        resVal = res.get("txhash")
        txHash = resVal if resVal is not None else res.get("raw_log")
    else:
        txHash = json.loads(res)["txhash"]

    if txHash is None:
        raise Exception("No txHash found", res)

    return txHash


def query_contract(query_base: RequestBase, contract_addr: str, msg: str) -> dict:
    cmd = f"query wasm contract-state smart {contract_addr} {msg}"
    res = send_request(query_base, cmd)
    return res


def b64encode(MSG: dict):
    if isinstance(MSG, str):
        MSG = json.loads(MSG)

    return base64.b64encode(remove_spaces(MSG)).decode("utf-8")


def remove_spaces(MSG: dict):
    return json.dumps(MSG, separators=(",", ":")).encode("utf-8")
