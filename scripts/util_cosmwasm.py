import base64
import hashlib
import json
import os
import time
from dataclasses import dataclass

from httpx import get, post

from util_base import contracts_path, current_dir, parent_dir
from util_contracts import get_contract_address

# from util_req import RequestBase, send_request
from util_req import RequestBase, RequestType

# make a request builder


class RequestBuilder:
    def __init__(self, apiEndpoint: str, chainID: str):
        self.apiEndpoint = apiEndpoint
        self.chainID = chainID

        if self.apiEndpoint == "":
            raise Exception("RequestBuilder apiEndpoint is empty")

        if self.chainID == "":
            raise Exception("RequestBuilder chainID is empty")

    # TODO: Add specific for each?
    def bin(self) -> RequestBase:
        return RequestBase(self.apiEndpoint, self.chainID, RequestType.BIN)

    def query(self) -> RequestBase:
        """
            # on actual query, do this check (currently in send_request)
            if base.request_type == RequestType.QUERY:
        if cmd.lower().startswith("query "):
            cmd = cmd[6:]
        elif cmd.lower().startswith("q "):
            cmd = cmd[2:]
        """
        return RequestBase(self.apiEndpoint, self.chainID, RequestType.QUERY)

    # TODO: i do not use anywhere atm, remove?
    def exec(self) -> RequestBase:
        return RequestBase(self.apiEndpoint, self.chainID, RequestType.EXEC)


@dataclass
class ActionHandler:
    ChainId: str = ""
    Action: str = ""
    Cmd: str = ""

    def __init__(self, chainId: str, action: str, cmd: str):
        self.ChainId = chainId
        self.Action = action
        self.Cmd = cmd

    def to_json(self) -> str:
        return json.dumps(self.__dict__)


# util_req.py
def send_request(
    base: RequestBase = RequestBase("", "", RequestType.BIN),
    cmd: str = "",
    returnText: bool = False,
) -> dict:
    if base.request_type == RequestType.QUERY:
        if cmd.lower().startswith("query "):
            cmd = cmd[6:]
        elif cmd.lower().startswith("q "):
            cmd = cmd[2:]

    data = ActionHandler(base.chain_id, base.request_type.value, cmd).to_json()
    print("[send_request]", data)
    r = post(base.URL, json=data, headers={"Content-Type": "application/json"})

    # This is messy, clean up
    if returnText:
        return dict(text=r.text)

    try:
        # Is there ever a case this does not work?
        return json.loads(r.text)
    except:
        return {"parse_error": r.text}


# TODO: type handler this better with a dataclass
@dataclass
class TransactionResponse:
    TxHash: str = ""
    RawLog: str | None = ""


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


class CosmWasm:
    def __init__(self, api: str, chainId: str):
        self.api = api  # http://localhost:8080
        self.chainId = chainId
        self.contractAddr = ""
        # action handlers here

        rb = RequestBuilder(self.api, self.chainId)
        self.query_base = rb.query()
        self.bin_base = rb.bin()

        # the last obtained Tx hash
        self.tx_hash = ""

    def get_latest_tx_hash(self) -> str:
        return self.tx_hash

    def instantiate_contract(
        self, codeId: int | str, msg: str | dict, label: str, flags: str = ""
    ) -> "CosmWasm":
        # not sure if we want this logic or not...
        # if len(self.contractAddr) > 0:
        #     raise Exception("Contract address already set")

        # if "--output=json" not in flags:
        #     msg += " --output=json"

        # TODO: properly handle the quotes
        if isinstance(msg, dict):
            msg = json.dumps(msg)

        res = send_request(
            self.bin_base, f"tx wasm instantiate {codeId} {msg} --label={label} {flags}"
        )
        # Get chain block time here instead
        time.sleep(2.5)

        # just have send_request return this?
        tx_res = get_transaction_response(res)

        contractAddr = get_contract_address(self.query_base, tx_res.TxHash)
        print(f"[instantiate_contract] {contractAddr=}\n")

        self.tx_hash = tx_res.TxHash
        self.contractAddr = contractAddr
        return self

    def execute_contract(
        self,
        msg: str | dict,
        flags: str = "",
    ) -> "CosmWasm":
        txHash: str | None

        if "--output=json" not in flags:
            flags += " --output=json"

        if isinstance(msg, dict):
            msg = json.dumps(msg)

        cmd = f"tx wasm execute {self.contractAddr} {msg} {flags}"
        print("[execute_contract]", cmd)

        res = send_request(self.bin_base, cmd)

        # just have send_request return this?
        tx_res = get_transaction_response(res)

        self.tx_hash = tx_res.TxHash

        return self

    def query_contract(self, msg: str | dict) -> dict:
        # TODO: properly handle the quotes
        if isinstance(msg, dict):
            msg = json.dumps(msg)

        cmd = f"query wasm contract-state smart {self.contractAddr} {msg}"
        res = send_request(self.query_base, cmd)
        return res

    @staticmethod
    def download_base_contracts():
        files = [
            "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw20_base.wasm",
            "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw4_group.wasm",
            "https://github.com/CosmWasm/cw-nfts/releases/latest/download/cw721_base.wasm",
        ]

        # download the file to the contracts/ folder
        for url in files:
            name = url.split("/")[-1]
            file_path = os.path.join(contracts_path, name)

            if os.path.exists(file_path):
                continue

            print(f"Downloading {name} to {file_path}")
            r = get(url, allow_redirects=True)
            with open(file_path, "wb") as f:
                f.write(r.content)

    @staticmethod
    def get_file_hash(rel_file_path: str, chainId: str) -> str:
        BUF_SIZE = 65536  # 64k chunks
        sha1 = hashlib.sha1()

        file_path = os.path.join(contracts_path, rel_file_path)

        if not os.path.exists(file_path):
            raise FileNotFoundError(f"File not found: {file_path}")

        sha1.update(chainId.replace("-", "").encode("utf-8"))
        with open(file_path, "rb") as f:
            while True:
                data = f.read(BUF_SIZE)
                if not data:
                    break
                sha1.update(data)

        return sha1.hexdigest()


if __name__ == "__main__":
    CosmWasm.download_base_contracts()

    # cw = CosmWasm(api="http://localhost:8080", chainId="localjuno-1")

    # cw.instantiate_contract(1, {}, label="contract", flags="--output=json")
    # cw.execute_contract({"increment": {}}, flags="--output=json")
    # cw.query_contract({"get_count": {}})

    pass
