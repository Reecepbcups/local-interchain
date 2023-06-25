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
- Execute on a contract
- -> sends a packet to the other contract and inc the counter
- Query the counter to ensure it increased.
"""

import os

from helpers.cosmwasm import CosmWasm
from helpers.relayer import Relayer
from helpers.transactions import RequestBuilder
from util_base import API_URL

WASM_FILE_NAME = "cw_ibc_example.wasm"

KEY_NAME = "acc0"
KEY_NAME2 = "second0"

CHAIN_ID = "localjuno-1"
CHAIN_ID2 = "localjuno-2"


def setup_env(rbs: list[RequestBuilder] = []):
    for rb in rbs:
        rb.bin("config keyring-backend test", log_output=True)
        rb.bin("config output json", log_output=True)


def main():
    absolute_path = os.path.abspath(__file__)
    parent_dir = os.path.dirname(os.path.dirname(absolute_path))
    contracts_dir = os.path.join(parent_dir, "contracts")

    relayer = Relayer(API_URL, CHAIN_ID)

    if True:
        print("⚙️ Setting env configuration")
        setup_env(
            [
                RequestBuilder(apiEndpoint=API_URL, chainID=CHAIN_ID),
                RequestBuilder(apiEndpoint=API_URL, chainID=CHAIN_ID2),
            ]
        )

        print("\n📝 Uploading Contracts")

        contract_1 = CosmWasm(API_URL, CHAIN_ID)
        contract_2 = CosmWasm(API_URL, CHAIN_ID2)

        codeIdA = contract_1.store_contract(
            KEY_NAME, os.path.join(contracts_dir, WASM_FILE_NAME)
        )
        codeIdB = contract_2.store_contract(
            KEY_NAME2, os.path.join(contracts_dir, WASM_FILE_NAME)
        )

        print("\n🪞 Instantiate Contracts on both chains")
        contract_1.instantiate_contract(
            account_key=KEY_NAME,
            codeId=codeIdA,
            msg="{}",
            label="contractA",
            flags="",
        )
        contract_2.instantiate_contract(
            account_key=KEY_NAME2,
            codeId=codeIdB,
            msg="{}",
            label="contractB",
            flags="",
        )

        print("\n📤 Create Contract Connection")
        relayer.create_wasm_connection(
            path="juno-ibc-1",
            src=contract_1.contractAddr,
            dst=contract_2.contractAddr,
            order="unordered",
            version="counter-1",
        )
    else:
        # If we already uploaded the contracts and instantiated them
        # # we can just skip the above steps.
        contract_1 = CosmWasm(
            API_URL,
            CHAIN_ID,
            contractAddrOverride="juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8",  # noqa: E501
        )
        contract_2 = CosmWasm(
            API_URL,
            CHAIN_ID2,
            contractAddrOverride="juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8",  # noqa: E501
        )

    print(relayer.get_channels())

    # 'Increment {}' execute on chain b, check chainA after.
    print("\n⚔️ Execute increment")
    contract_2.execute_contract(
        accountKey=KEY_NAME2, msg={"increment": {"channel": "channel-1"}}
    )

    print("\n📨 Flush Packets...")
    relayer.flush("juno-ibc-1", "channel-1", log_output=True)

    # Ensure the Tx count increased += 1
    if True:
        print("\n❓ Query Count")
        contract1Res = contract_1.query_contract(
            {"get_count": {"channel": "channel-1"}}
        )
        print(f"{contract1Res=}")


if __name__ == "__main__":
    main()
