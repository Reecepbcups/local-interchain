import json
from dataclasses import dataclass

# from util_req import RequestBase, send_request
# from util_req import RequestBase, RequestType
from enum import Enum

from httpx import get, post

from util_base import contracts_path, current_dir, parent_dir

# from util_contracts import get_contract_address


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


class RequestType(Enum):
    BIN = "bin"
    QUERY = "query"
    EXEC = "exec"


@dataclass(frozen=True)
class RequestBase:
    URL: str
    chain_id: str
    request_type: RequestType


# TODO: type handler this better with a dataclass
@dataclass
class TransactionResponse:
    TxHash: str = ""
    RawLog: str | None = ""


@dataclass
class ActionHandler:
    chain_id: str = ""
    action: str = ""
    cmd: str = ""

    def __init__(self, chain_id: str, action: str, cmd: str):
        self.chain_id = chain_id
        self.action = action
        self.cmd = cmd

    def to_json(self) -> dict:
        return {
            "chain_id": self.chain_id,
            "action": self.action,
            "cmd": self.cmd,
        }


class RequestBuilder:
    def __init__(self, apiEndpoint: str, chainID: str):
        self.apiEndpoint = apiEndpoint
        self.chainID = chainID

        if self.apiEndpoint == "":
            raise Exception("RequestBuilder apiEndpoint is empty")

        if self.chainID == "":
            raise Exception("RequestBuilder chainID is empty")

    # TODO: Add specific for each?
    def bin(self, cmd: str, log_output: bool = False) -> dict:
        rb = RequestBase(self.apiEndpoint, self.chainID, RequestType.BIN)
        return send_request(rb, cmd, log_output=log_output)

    def query(self, cmd: str, log_output: bool = False) -> dict:
        """
            # on actual query, do this check (currently in send_request)
            if base.request_type == RequestType.QUERY:
        if cmd.lower().startswith("query "):
            cmd = cmd[6:]
        elif cmd.lower().startswith("q "):
            cmd = cmd[2:]
        """
        rb = RequestBase(self.apiEndpoint, self.chainID, RequestType.QUERY)
        return send_request(rb, cmd, log_output=log_output)

    # What / when is response?
    def query_tx(self, response: str | dict) -> dict:
        tx_hash = get_tx_hash(response)
        if len(tx_hash) == 0:
            return dict(error="tx_hash is empty")

        res = self.query(f"tx {tx_hash} --output json")
        return dict(tx=res)

    # TODO: i do not use anywhere atm, remove?
    def exec(self, cmd: str, log_output: bool = False) -> dict:
        rb = RequestBase(self.apiEndpoint, self.chainID, RequestType.EXEC)
        return send_request(rb, cmd, log_output=log_output)


# util_req.py
def send_request(
    base: RequestBase = RequestBase("", "", RequestType.BIN),
    cmd: str = "",
    returnText: bool = False,
    log_output: bool = False,
) -> dict:
    if base.request_type == RequestType.QUERY:
        if cmd.lower().startswith("query "):
            cmd = cmd[6:]
        elif cmd.lower().startswith("q "):
            cmd = cmd[2:]

    data = ActionHandler(base.chain_id, base.request_type.value, cmd).to_json()
    if log_output:
        print("[send_request data]", data)

    r = post(base.URL, json=data, headers={"Content-Type": "application/json"})

    if log_output:
        print("[send_request resp]", r.text)

    # This is messy, clean up
    if returnText:
        return dict(text=r.text)

    try:
        # Is there ever a case this does not work?
        return json.loads(r.text)
    except:
        return {"parse_error": r.text}


def get_transaction_response(send_req_res: str | dict) -> TransactionResponse:
    txr = TransactionResponse()

    if isinstance(send_req_res, str):
        try:
            json.loads(send_req_res)
        except:
            txr.RawLog = send_req_res
            return txr

        txHash = json.loads(send_req_res)["txhash"]
        txr.TxHash = txHash

    if isinstance(send_req_res, dict):
        thash = send_req_res.get("txhash")
        txr.TxHash = thash if thash is not None else ""
        txr.RawLog = send_req_res.get("raw_log")

    if txr.TxHash is None:
        raise Exception("No txHash found", send_req_res)

    return txr
