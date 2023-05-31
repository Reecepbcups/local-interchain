"""
This test the rest server to ensures it functions properly.

You could also do in curl
curl http://localhost:8080/ --include --header "Content-Type: application/json" -X POST --data '{"chain-id":"localjuno-1","action":"q","cmd":"bank total"}'
"""

import json
import os
import time
from dataclasses import dataclass

import httpx

current_dir = os.path.dirname(os.path.realpath(__file__))
with open(os.path.join(current_dir, "configs", "server.json")) as f:
    server_config = json.load(f)["server"]

PORT = server_config["port"]
HOST = server_config["host"]
URL = f"http://{HOST}:{PORT}/"


@dataclass
class Request:
    chain_id: str
    action: str
    cmd: str

    def to_dict(self) -> dict:
        return {
            "chain-id": self.chain_id,
            "action": self.action,
            "cmd": self.cmd,
        }


def bin_request(cmd: str, asJSON: bool = False) -> str | dict:
    data = Request("localjuno-1", "bin", cmd).to_dict()
    r = httpx.post(URL, json=data, headers={"Content-Type": "application/json"})
    if asJSON:
        return json.loads(r.text)
    return r.text


def query_request(cmd: str, asJSON: bool = False) -> str | dict:
    data = Request("localjuno-1", "query", cmd).to_dict()
    r = httpx.post(URL, json=data, headers={"Content-Type": "application/json"})

    if asJSON:
        return json.loads(r.text)

    return r.text


def docker_command(cmd: str, asJSON: bool = False) -> str | dict:
    data = Request("localjuno-1", "exec", cmd).to_dict()
    r = httpx.post(URL, json=data, headers={"Content-Type": "application/json"})

    if asJSON:
        return json.loads(r.text)

    return r.text


# Test

res = bin_request("keys list --keyring-backend=test")
print(res)

res = bin_request(
    "tx decode ClMKUQobL2Nvc21vcy5nb3YudjFiZXRhMS5Nc2dWb3RlEjIIpwISK2p1bm8xZGM3a2MyZzVrZ2wycmdmZHllZGZ6MDl1YTlwZWo1eDNsODc3ZzcYARJmClAKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECxjGMmYp4MlxxfFWi9x4u+jOleJVde3Cru+HnxAVUJmgSBAoCCH8YNBISCgwKBXVqdW5vEgMyMDQQofwEGkDPE4dCQ4zUh6LIB9wqNXDBx+nMKtg0tEGiIYEH8xlw4H8dDQQStgAe6xFO7I/oYVSWwa2d9qUjs9qyB8r+V0Gy",
    asJSON=True,
)
print(res)

res = bin_request("config keyring-backend test")
print(res)

res = bin_request("config node %RPC%")
print(res)

res = bin_request("config", asJSON=True)
print(res)

res = bin_request("keys list --output=json", asJSON=True)
print(res)

res = query_request(
    "bank balances juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 --output=json",
    asJSON=True,
)
print(res)

# Transaction test
res = bin_request(
    "tx bank send acc0 juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 500ujuno --fees 5000ujuno --node %RPC% --chain-id %CHAIN_ID% --yes --output json",
    asJSON=True,
)
print(res)

txHash = ""
if isinstance(res, dict):
    txHash = res["txhash"]

if len(txHash) > 0:
    time.sleep(7)
    res = query_request(
        f"tx {txHash} --output json",
        asJSON=True,
    )
    print(res)

res = query_request(
    "bank balances juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 --output=json",
    asJSON=True,
)
print(res)
