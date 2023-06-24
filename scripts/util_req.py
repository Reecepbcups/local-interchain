import json
from dataclasses import dataclass
from enum import Enum

import httpx


# Make an enum of type request_type
class RequestType(Enum):
    BIN = "bin"
    QUERY = "query"
    EXEC = "exec"


@dataclass(frozen=True)
class RequestBase:
    URL: str
    chain_id: str
    request_type: RequestType


def send_request(
    base: RequestBase = RequestBase("", "", RequestType.BIN),
    cmd: str = "",
    returnText: bool = False,
) -> dict:
    if base.chain_id == "":
        raise Exception("send_request Chain ID is empty")

    if base.URL == "":
        raise Exception("send_request URL is empty")

    if base.request_type == RequestType.QUERY:
        if cmd.lower().startswith("query "):
            cmd = cmd[6:]
        elif cmd.lower().startswith("q "):
            cmd = cmd[2:]

    data = {
        "chain-id": base.chain_id,
        "action": base.request_type.value,
        "cmd": cmd,
    }
    print("[send_request]", data)
    r = httpx.post(base.URL, json=data, headers={"Content-Type": "application/json"})

    if returnText:
        return dict(text=r.text)

    try:
        # Is there ever a case this does not work?
        return json.loads(r.text)
    except:
        return {"parse_error": r.text}
