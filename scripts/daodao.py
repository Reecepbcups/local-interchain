"""
Steps:
- Download contracts from the repo based off release version
- Ensure ictest is started for a chain
- Upload to JUNO (store) and init
- Connect together or what ever'
- Profit
"""

from api_test import send_request
from util_base import URL
from util_contracts import b64encode, instantiate_contract, remove_spaces
from util_req import RequestBase, RequestType
from util_txs import store_contract

KEY_NAME = "acc0"
CHAIN_ID = "localjuno-1"

bin_base = RequestBase(URL, CHAIN_ID, RequestType.BIN)
query_base = RequestBase(URL, CHAIN_ID, RequestType.QUERY)


def main():
    send_request(bin_base, f"config keyring-backend test")
    FLAGS = "--home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json --gas=auto --gas-adjustment=2.0"

    dao_proposal_single_code_id = store_contract(
        bin_base, KEY_NAME, f"../contracts/dao_proposal_single.wasm"
    )

    dao_voting_native_staked_code_id = store_contract(
        bin_base, KEY_NAME, f"../contracts/dao_voting_native_staked.wasm"
    )

    dao_core_code_id = store_contract(bin_base, KEY_NAME, f"../contracts/dao_core.wasm")

    # https://github.com/DA0-DA0/dao-contracts/blob/main/scripts/create-v2-dao-native-voting.sh
    MODULE_MSG = {
        "allow_revoting": False,
        "max_voting_period": {"time": 604800},
        "close_proposal_on_execution_failure": True,
        "pre_propose_info": {"anyone_may_propose": {}},
        "only_members_execute": True,
        "threshold": {
            "threshold_quorum": {
                "quorum": {"percent": "0.20"},
                "threshold": {"majority": {}},
            }
        },
    }
    ENCODED_PROP_MESSAGE = b64encode(MODULE_MSG)
    # print(ENCODED_PROP_MESSAGE)

    VOTING_MSG = '{"owner":{"core_module":{}},"denom":"ujuno"}'
    ENCODED_VOTING_MESSAGE = b64encode(VOTING_MSG)
    # print(ENCODED_VOTING_MESSAGE)

    CW_CORE_INIT = remove_spaces(
        {
            "admin": "juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk",
            "automatically_add_cw20s": True,
            "automatically_add_cw721s": True,
            "description": "V2_DAO",
            "name": "V2_DAO",
            "proposal_modules_instantiate_info": [
                {
                    "admin": {"core_module": {}},
                    "code_id": dao_proposal_single_code_id,
                    "label": "v2_dao",
                    "msg": f"{ENCODED_PROP_MESSAGE}",
                }
            ],
            "voting_module_instantiate_info": {
                "admin": {"core_module": {}},
                "code_id": dao_voting_native_staked_code_id,
                "label": "test_v2_dao-cw-native-voting",
                "msg": f"{ENCODED_VOTING_MESSAGE}",
            },
        }
    ).decode("utf-8")

    # print(CW_CORE_INIT)

    addr = instantiate_contract(
        query_base=query_base,
        bin_base=bin_base,
        codeId=dao_core_code_id,
        msg=CW_CORE_INIT,
        label="dao_core",
        flags=f"--from {KEY_NAME} --no-admin {FLAGS}",
    )
    print(addr)


if __name__ == "__main__":
    main()
