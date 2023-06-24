import json
from dataclasses import dataclass
from enum import Enum

import httpx

# Make an enum of type request_type


@dataclass(frozen=True)
class RelayBase:
    URL: str
    chain_id: str


def send_relay_request(
    base: RelayBase = RelayBase("", ""),
    action: str = "",
    cmd: str = "",
    returnText: bool = False,
) -> dict:
    if base.URL == "":
        raise Exception("send_request URL is empty")

    data = {
        "chain-id": base.chain_id,
        "action": action,
        "cmd": cmd,
    }
    # print("[relayer data]", data)

    print("[relayer]", data["cmd"])

    r = httpx.post(
        base.URL, json=data, headers={"Content-Type": "application/json"}, timeout=120
    )

    if returnText:
        return dict(text=r.text)

    try:
        # Is there ever a case this does not work?
        return json.loads(r.text)
    except:
        return {"parse_error": r.text}