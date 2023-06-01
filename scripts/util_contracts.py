import base64
import hashlib
import json
import os
import time

from util_base import contracts_path, current_dir, parent_dir
from util_req import RequestBase, send_request


def get_file_hash(rel_file_path: str) -> str:
    BUF_SIZE = 65536  # 64k chunks
    md5 = hashlib.md5()
    sha1 = hashlib.sha1()

    file_path = os.path.join(contracts_path, rel_file_path)

    # if file_path does not exist, throw error
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"File not found: {file_path}")

    with open(file_path, "rb") as f:
        while True:
            data = f.read(BUF_SIZE)
            if not data:
                break
            md5.update(data)
            sha1.update(data)

    return sha1.hexdigest()


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
    value: str
    # res = send_request(bin_base, msg)

    if "--output=json" not in msg:
        msg += " --output=json"

    res = send_request(
        bin_base, f"tx wasm instantiate {codeId} {msg} --label={label} {flags}"
    )
    time.sleep(2.5)

    if isinstance(res, dict):
        value = res["txhash"]
    else:
        value = json.loads(res)["txhash"]

    print(f"{value=}")

    addr = get_contract_address(query_base, value)
    return addr


def b64encode(MSG: dict):
    if isinstance(MSG, str):
        MSG = json.loads(MSG)

    return base64.b64encode(remove_spaces(MSG)).decode("utf-8")


def remove_spaces(MSG: dict):
    return json.dumps(MSG, separators=(",", ":")).encode("utf-8")
