#
# SHows how to relay between contracts

"""

pip install httpx

Steps:
- Compile https://github.com/0xekez/cw-ibc-example
- Copy to ./contracts/cw_ibc_example.wasm

- local-ic start two

- Init both contracts
- Create a channel/connection between both (counter-1 version)
- Execute on a contract, which then sends a packet to the other contract andc increments the counter
- Query the counter to ensure it increased.
"""

import os
import time
from base64 import b64decode

import httpx
from api_test import send_request
from util_base import URL, contracts_path
from util_contracts import (
    b64encode,
    execute_contract,
    instantiate_contract,
    query_contract,
    remove_spaces,
)
from util_relayer import RelayBase, send_relay_request
from util_req import RequestBase, RequestType
from util_txs import store_contract

WASM_FILE_NAME = "cw_ibc_example.wasm"

KEY_NAME = "acc0"
KEY_NAME2 = "second0"

CHAIN_ID = "localjuno-1"
CHAIN_ID2 = "localjuno-2"

# simplify? This is pretty ugly and overly complex.
relayer_base = RequestBase(f"{URL}relayer", CHAIN_ID, RequestType.BIN)

bin_base = RequestBase(URL, CHAIN_ID, RequestType.BIN)
query_base = RequestBase(URL, CHAIN_ID, RequestType.QUERY)

bin_base2 = RequestBase(URL, CHAIN_ID2, RequestType.BIN)
query_base2 = RequestBase(URL, CHAIN_ID2, RequestType.QUERY)

# ew
r = RelayBase(f"{URL}relayer", CHAIN_ID)


def setup_env(bases: list[RequestBase] = []):
    for b in bases:
        send_request(b, f"config keyring-backend test")
        send_request(b, f"config output json")


def main():
    FLAGS = "--home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json --gas=auto --gas-adjustment=2.0"

    absolute_path = os.path.abspath(__file__)
    parent_dir = os.path.dirname(os.path.dirname(absolute_path))
    contracts_dir = os.path.join(parent_dir, "contracts")

    # Upload to chain A and B
    if True:
        print("⚙️ Setting env configuration")
        setup_env([bin_base, bin_base2])

        print("\n📝 Uploading Contracts")
        codeIdA = store_contract(
            bin_base,
            KEY_NAME,
            os.path.join(contracts_dir, WASM_FILE_NAME),
        )
        codeIdB = store_contract(
            bin_base2,
            KEY_NAME2,
            os.path.join(contracts_dir, WASM_FILE_NAME),
        )

        # instantiate both on each chain
        print("\n🪞 Instantiate Contracts")
        addrA = instantiate_contract(
            query_base=query_base,
            bin_base=bin_base,
            codeId=codeIdA,
            msg="{}",
            label="contractA",
            flags=f"--from {KEY_NAME} --no-admin {FLAGS}",
        )
        addrB = instantiate_contract(
            bin_base=bin_base2,
            query_base=query_base2,
            codeId=codeIdB,
            msg="{}",
            label="contractB",
            flags=f"--from {KEY_NAME2} --no-admin {FLAGS}",
        )

        print("📤 Create Contract Connection")
        send_relay_request(
            r,
            "execute",
            f"rly transact channel juno-ibc-1 --src-port wasm.{addrA} --dst-port wasm.{addrB} --order unordered --version counter-1",
        )

        # print(send_relay_request(r, action="get_channels"))
    else:
        # If we already uploaded the contracts and instantiated them, we can just skip the above steps.
        addrA = "juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"
        addrB = "juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"

    # 'Increment {}' execute
    if True:
        print("\n⚔️ Execute increment")
        txHash = execute_contract(
            bin_base=bin_base2,
            contract_addr=addrA,
            msg='{"increment":{"channel":"channel-1"}}',
            flags=f"--from {KEY_NAME2} {FLAGS}",
        )

    # Flush Packets
    print("\n📨 Flush Packets...")
    send_relay_request(r, "execute", "rly transact flush juno-ibc-1 channel-1")

    # Ensure the Tx count increased += 1
    if True:
        print("\n❓ Query Count")
        res = query_contract(
            query_base=query_base,
            contract_addr=addrA,
            msg='{"get_count":{"channel":"channel-1"}}',
        )
        print(f"{res=}")
        print()


if __name__ == "__main__":
    main()
