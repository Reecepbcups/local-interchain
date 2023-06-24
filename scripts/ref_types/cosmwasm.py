import json
import os
import time
from base64 import b64decode

from httpx import get, post

from ref_types.transactions import (
    ActionHandler,
    RequestBuilder,
    TransactionResponse,
    get_transaction_response,
    send_request,
)
from util_base import (
    contracts_json_path,
    contracts_path,
    current_dir,
    default_contracts_json,
    get_cache_or_default,
    get_chain_start_time_from_logs,
    get_file_hash,
    parent_dir,
    update_cache,
)
from util_contracts import get_contract_address


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

    r = post(
        url,
        json=data,
        headers={"Content-Type": "application/json"},
        timeout=120,
    )

    if r.status_code != 200:
        return dict(error=r.text)

    return json.loads(r.text.replace("\n", ""))


class CosmWasm:
    def __init__(self, api: str, chainId: str):
        self.api = api  # http://localhost:8080
        self.chainId = chainId
        self.contractAddr = ""
        self.codeId: int = -1
        # action handlers here

        self.rb = RequestBuilder(self.api, self.chainId)

        self.default_flag_set = "--home=%HOME% --node=%RPC% --chain-id=%CHAIN_ID% --yes --output=json --keyring-backend=test --gas=auto --gas-adjustment=2.0"

        # the last obtained Tx hash
        self.tx_hash = ""

    def get_latest_tx_hash(self) -> str:
        return self.tx_hash

    def store_contract(self, key_name: str, abs_path: str) -> int:
        ictest_chain_start = get_chain_start_time_from_logs()
        if ictest_chain_start == -1:
            return -1

        default_contracts_json()

        with open(contracts_json_path, "r") as f:
            contracts = json.load(f)

        contracts = get_cache_or_default(contracts, ictest_chain_start)

        sha1 = get_file_hash(abs_path, self.chainId)
        if sha1 in contracts["file_cache"]:
            self.codeId = contracts["file_cache"][sha1]
            print(f"[Cache] CodeID={self.codeId} for {abs_path.split('/')[-1]}")
            return self.codeId

        # TODO: make upload_file public & require bin_base input.
        res = _upload_file(self.api, self.chainId, key_name, abs_path)
        if "error" in res:
            raise Exception(res["error"])

        self.codeId = update_cache(contracts, res["code_id"], sha1)
        return self.codeId

    def instantiate_contract(
        self,
        account_key: str,
        codeId: int | str,
        msg: str | dict,
        label: str,
        admin: str | None = None,
        flags: str = "",
    ) -> "CosmWasm":
        # not sure if we want this logic or not...
        # if len(self.contractAddr) > 0:
        #     raise Exception("Contract address already set")

        # if "--output=json" not in flags:
        #     msg += " --output=json"

        if admin is None and "--no-admin" not in flags:
            flags += "--no-admin"

        # TODO: properly handle the quotes
        if isinstance(msg, dict):
            msg = json.dumps(msg, separators=(",", ":"))

        cmd = f"""tx wasm instantiate {codeId} {msg} --label={label} --from={account_key} {self.default_flag_set} {flags}"""
        res = self.rb.bin(cmd)
        # Get chain block time here instead
        time.sleep(1)

        # just have send_request return this?
        tx_res = get_transaction_response(res)

        contractAddr = get_contract_address(self.rb, tx_res.TxHash)
        print(f"[instantiate_contract] {contractAddr=}\n")

        self.tx_hash = tx_res.TxHash
        self.contractAddr = contractAddr
        return self

    def execute_contract(
        self,
        accountKey: str,
        msg: str | dict,
        flags: str = "",
    ) -> "CosmWasm":
        # if "--output=json" not in flags:
        #     flags += " --output=json"

        if isinstance(msg, dict):
            msg = json.dumps(msg, separators=(",", ":"))

        cmd = f"tx wasm execute {self.contractAddr} {msg} --from={accountKey} {self.default_flag_set} {flags}"
        print("[execute_contract]", cmd)
        res = self.rb.query(cmd)

        tx_res = get_transaction_response(res)

        self.tx_hash = tx_res.TxHash

        return self

    def query_contract(self, msg: str | dict) -> dict:
        # TODO: properly handle the quotes
        if isinstance(msg, dict):
            msg = json.dumps(msg, separators=(",", ":"))

        cmd = f"query wasm contract-state smart {self.contractAddr} {msg}"
        res = self.rb.query(cmd)
        return res

    @staticmethod
    def download_base_contracts():
        files = [
            "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw20_base.wasm",
            "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw4_group.wasm",
            "https://github.com/CosmWasm/cw-nfts/releases/latest/download/cw721_base.wasm",
        ]

        for url in files:
            name = url.split("/")[-1]
            file_path = os.path.join(contracts_path, name)

            if os.path.exists(file_path):
                continue

            print(f"Downloading {name} to {file_path}")
            r = get(url, follow_redirects=True)
            with open(file_path, "wb") as f:
                f.write(r.content)

    @staticmethod
    def download_mainnet_daodao_contracts():
        # From https://github.com/DA0-DA0/dao-contracts/releases
        # v2.1.0
        files = """cw20_base.wasm 2443
        cw20_stake.wasm 2444
        cw20_stake_external_rewards.wasm 2445
        cw20_stake_reward_distributor.wasm 2446
        cw4_group.wasm 2447
        cw721_base.wasm 2448
        cw_admin_factory.wasm 2449
        cw_fund_distributor.wasm 2450
        cw_payroll_factory.wasm 2451
        cw_token_swap.wasm 2452
        cw_vesting.wasm 2453
        dao_core.wasm 2454
        dao_migrator.wasm 2455
        dao_pre_propose_approval_single.wasm 2456
        dao_pre_propose_approver.wasm 2457
        dao_pre_propose_multiple.wasm 2458
        dao_pre_propose_single.wasm 2459
        dao_proposal_condorcet.wasm 2460
        dao_proposal_multiple.wasm 2461
        dao_proposal_single.wasm 2462
        dao_voting_cw20_staked.wasm 2463
        dao_voting_cw4.wasm 2464
        dao_voting_cw721_staked.wasm 2465
        dao_voting_native_staked.wasm 2466"""

        for file in files.split("\n"):
            file = file.strip()
            name, codeId = file.split(" ")

            file_path = os.path.join(contracts_path, name)
            if os.path.exists(file_path):
                continue

            print(f"Downloading {name}")
            response = get(
                f"https://api.juno.strange.love/cosmwasm/wasm/v1/code/{codeId}",
                headers={
                    "accept": "application/json",
                },
                timeout=60,
            )
            data = response.json()

            binary = b64decode(data["data"])
            with open(file_path, "wb") as f:
                f.write(binary)


if __name__ == "__main__":
    CosmWasm.download_base_contracts()

    cw = CosmWasm(api="http://localhost:8080", chainId="localjuno-1")

    codeId = cw.store_contract("acc0", os.path.join(contracts_path, "cw721_base.wasm"))

    cw.instantiate_contract(
        "acc0",
        codeId,
        {
            "name": "name",
            "symbol": "NFT",
            # account in base.json genesis (acc0)
            "minter": "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl",
        },
        label="contract",
        flags="",
    )
    print(cw.tx_hash)
    print(cw.contractAddr)

    cw.execute_contract(
        "acc0",
        {
            "mint": {
                "token_id": "1",
                "owner": "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl",
                "token_uri": "https://reece.sh",
            }
        },
        flags="--output=json",
    )

    print(cw.query_contract({"contract_info": {}}))
    print(cw.query_contract({"all_nft_info": {"token_id": "1"}}))
