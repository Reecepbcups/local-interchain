import json

import httpx

# TODO: helpers/relayer.py


class Relayer:
    def __init__(self, API_URL: str, CHAIN_ID: str, log_output: bool = False):
        self.api = API_URL
        self.api_relayer = f"{API_URL}/relayer"
        self.chain_id = CHAIN_ID
        self.log_output = log_output

    def exec(self, cmd: str, return_text: bool = False) -> dict:
        if self.api == "":
            raise Exception("send_request URL is empty")

        data = {
            "chain_id": self.chain_id,
            "action": "execute",
            "cmd": cmd,
        }

        if self.log_output:
            print("[relayer]", data["cmd"])

        r = httpx.post(
            self.api_relayer,
            json=data,
            headers={"Content-Type": "application/json"},
            timeout=120,
        )

        if return_text:
            return dict(text=r.text)

        try:
            # Is there ever a case this does not work?
            return json.loads(r.text)
        except:
            return {"parse_error": r.text}

    def create_wasm_connection(
        self, path: str, src: str, dst: str, order: str, version: str
    ):
        if not src.startswith("wasm."):
            src = f"wasm.{src}"

        if not dst.startswith("wasm."):
            dst = f"wasm.{dst}"

        self.exec(
            f"rly transact channel {path} --src-port {src} --dst-port {dst} --order {order} --version {version}"
        )

        pass

    def flush(self, path: str, channel: str, log_output: bool = False) -> dict:
        res = self.exec(
            f"rly transact flush {path} {channel}",
        )
        if log_output:
            print(res)
        return res

    def get_channels(self) -> dict:
        r = httpx.post(
            self.api_relayer,
            json={
                "chain_id": self.chain_id,
                "action": "get_channels",
            },
            headers={"Content-Type": "application/json"},
        )
        return json.loads(r.text)
