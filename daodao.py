"""
Steps:
- Download contracts from the repo based off release version
- Ensure ictest is started for a chain
- Upload to JUNO (store) and init
- Connect together or what ever'
- Profit
"""

import base64
import json
import os
import time

import httpx

from test_rest import bin_request, query_request

KEY_NAME = "acc0"

current_dir = os.path.dirname(os.path.realpath(__file__))
parent_dir = os.path.dirname(current_dir)
with open(os.path.join(current_dir, "configs", "server.json")) as f:
    server_config = json.load(f)["server"]

PORT = server_config["port"]
HOST = server_config["host"]
URL = f"http://{HOST}:{PORT}/"
UPLOAD = f"http://{HOST}:{PORT}/upload"

contracts_path = os.path.join(current_dir, "contracts")
if not os.path.exists(contracts_path):
    os.mkdir(contracts_path)


import hashlib


def get_file_hash(rel_file_path: str) -> str:
    BUF_SIZE = 65536  # 64k chunks
    md5 = hashlib.md5()
    sha1 = hashlib.sha1()

    file_path = os.path.join(contracts_path, rel_file_path)

    with open(file_path, "rb") as f:
        while True:
            data = f.read(BUF_SIZE)
            if not data:
                break
            md5.update(data)
            sha1.update(data)

    return sha1.hexdigest()


def store_contract(rel_file_path: str) -> str | dict:
    print(f"Uploading {rel_file_path}")

    # create configs/.contracts.json if not already
    contracts_json_path = os.path.join(current_dir, "configs", ".contracts.json")
    if not os.path.exists(contracts_json_path):
        with open(contracts_json_path, "w") as f:
            f.write("{}")

    sha1 = get_file_hash(rel_file_path)
    with open(contracts_json_path, "r") as f:
        contracts_json = json.load(f)

    if sha1 in contracts_json:
        return contracts_json[sha1]

    data = {
        "chain-id": "localjuno-1",
        "key-name": KEY_NAME,
        "file-name": rel_file_path,
    }
    r = httpx.post(
        UPLOAD,
        json=data,
        headers={"Content-Type": "application/json"},
        timeout=120,
    )

    # TODO: Get code id and save to the contracts_json_path
    if r.status_code != 200:
        print(r.text)
        return r.text

    res = json.loads(r.text)

    contracts_json[sha1] = res["code_id"]
    with open(contracts_json_path, "w") as f:
        json.dump(contracts_json, f)

    return res["code_id"]


def getContractAddr(tx_hash: str) -> str:
    # junod q tx 6F6586DD1C2F8D1690EF47FE775A845ADF9D2301233FF9D36BF8570C55FD7297 --node http://localhost:36727
    res = query_request(f"tx {tx_hash} --output=json")
    print(res)

    try:
        res_json = json.loads(str(res))
    except:
        print(f"Error parsing JSON. {res}")
        return str(res)

    code = int(res_json["code"])
    if code != 0:
        raw = res_json["raw_log"]
        print(res_json)
        return raw

    contract_addr = ""

    print(res_json["logs"])

    for event in res_json["logs"][0]["events"]:
        for attr in event["attributes"]:
            if attr["key"] == "_contract_address":
                contract_addr = attr["value"]

    return contract_addr


def instantiate_contract(msg: str) -> str:
    # tx = f"tx wasm instantiate {cw4CodeId} {init} --label=cw4 --from {KEY_NAME} --no-admin --home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json"
    res = bin_request(msg, asJSON=False)
    print(res)
    time.sleep(2.5)
    addr = getContractAddr(json.loads(str(res))["txhash"])
    return addr


def new():
    bin_request(f"config keyring-backend test")
    FLAGS = "--home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json --gas=auto --gas-adjustment=2.0"

    # store cw4_group.wasm
    # dao_voting_cw4.wasm
    # dao_core.wasm
    # dao_proposal_single.wasm

    # bin_request("keys list --keyring-backend=test")

    daoProposalSingleCode = 16
    # daoProposalSingleCode = store_contract(
    #     f"../contracts/dao_proposal_single.wasm", asJSON=True
    # )
    print(daoProposalSingleCode)

    CORE_MOD_MSG = {
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
    ENCODED_CORE_MOD_MSG = base64.b64encode(
        json.dumps(CORE_MOD_MSG).encode("utf-8")
    ).decode("utf-8")

    # == CW4 ==
    cw4CodeId = 6
    # cw4CodeId = store_contract(f"../contracts/cw4_group.wasm", asJSON=True)["code_id"]
    # init = '{"members":[{"addr":"juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl","weight":100}]}'
    # tx = f"tx wasm instantiate {cw4CodeId} {init} --label=cw4 --from {KEY_NAME} --no-admin --home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json"
    # res = bin_request(tx)
    # time.sleep(2.5)
    # addr = getContractAddr(json.loads(res)["txhash"])
    # print(addr)

    # == DAO Voting CW4 ==
    daoVotingCW4Code = 14
    # # daoVotingCW4Code = store_contract(f"../contracts/dao_voting_cw4.wasm", asJSON=True)[
    # #     "code_id"
    # # ]
    # print(daoVotingCW4Code)
    VOTING_MSG = '{"cw4_group_code_id":6,"initial_members":[{"addr":"juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk","weight":30},{"addr":"juno16mrjtqffn3awme2eczhlpwzj7mnatkeluvhj6c","weight":1}]}'
    ENCODED_VOTING_MESSAGE = base64.b64encode(
        json.dumps(VOTING_MSG).encode("utf-8")
    ).decode("utf-8")
    # tx = f"tx wasm instantiate {daoVotingCW4Code} {VOTING_INIT} --label=dao_cw4 --from {KEY_NAME} --no-admin {FLAGS}"
    # res = bin_request(tx)
    # print(res)
    # time.sleep(2.5)
    # addr = getContractAddr(json.loads(res)["txhash"])
    # print(addr)

    ## == Admin Factory ==
    cwAdminFactoryCode = 15
    # cwAdminFactoryCode = store_contract(
    #     f"../contracts/cw_admin_factory.wasm", asJSON=True
    # )["code_id"]
    print(cwAdminFactoryCode)
    # tx = f"tx wasm instantiate {cwAdminFactoryCode} {{}} --label=cw_admin_factory --from {KEY_NAME} --no-admin {FLAGS}"
    # res = bin_request(tx)
    # print(res)
    # time.sleep(2.5)
    # adminFactory = getContractAddr(json.loads(res)["txhash"])
    # print(adminFactory)

    adminFactory = "juno1gpvd7r5u6v2et6fva445k80gukfc7kqsql3luvh4zqxp6lvwepeqhanamp"

    # == DAO Core ==
    daoCoreCode = 17
    # daoCoreCode = store_contract(f"../contracts/dao_core.wasm", asJSON=True)["code_id"]
    # print(daoCoreCode)

    CW_CORE_INIT = json.dumps(
        {
            "admin": "juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk",
            "automatically_add_cw20s": True,
            "automatically_add_cw721s": True,
            "description": "V2_DAO",
            "name": "V2_DAO",
            "image_url": "https://nftstorage.link/ipfs/bafkreidawbt34hsqio4lfrivviccm4ahyrmltkpgancjmlh7mzubriyate/",
            "proposal_modules_instantiate_info": [
                {
                    "admin": {"core_module": {}},
                    "code_id": daoProposalSingleCode,
                    "label": "v2_dao",
                    "msg": ENCODED_CORE_MOD_MSG,
                }
            ],
            "voting_module_instantiate_info": {
                "admin": {"core_module": {}},
                "code_id": daoVotingCW4Code,
                "label": "test_v2_dao-cw4-voting",
                "msg": ENCODED_VOTING_MESSAGE,
            },
        },
        separators=(",", ":"),
    )

    # print(CW_CORE_INIT)
    # ENCODED_CORE_MSG = base64.b64encode(CW_CORE_INIT.encode("utf-8")).decode("utf-8")

    # '{CW_CORE_INIT}'

    # --label=dao_core
    # tx = "tx wasm execute " + adminFactory + " '{}' --from acc0 " + FLAGS

    # INIT_MSG='{"instantiate_contract_with_self_admin":{"code_id":695, "label": "v2 subDAO subDAO", "instantiate_msg":"'$CW_CORE_ENCODED'"}}'

    # tx = f"tx wasm execute {adminFactory} " + CW_CORE_INIT + f" --from acc0 {FLAGS}"
    # print(tx)
    # # exit(1)
    # res = bin_request(tx)
    # print(res)
    # time.sleep(2.5)
    # # addr = getContractAddr(json.loads(res)["txhash"])
    # # print(addr)

    # Just launch the core dao
    tx = (
        f"tx wasm instantiate {daoCoreCode} "
        + CW_CORE_INIT
        + f" --label=dao_core --from acc0 --no-admin {FLAGS}"
    )
    print(tx)
    # exit(1)
    res = bin_request(tx)
    print(res)
    time.sleep(2.5)
    # addr = getContractAddr(json.loads(res)["txhash"])
    # print(addr)

    pass


def encode(MSG: dict):
    if isinstance(MSG, str):
        MSG = json.loads(MSG)

    return base64.b64encode(remove_spaces(MSG)).decode("utf-8")


def remove_spaces(MSG: dict):
    return json.dumps(MSG, separators=(",", ":")).encode("utf-8")


def attempt_3():
    bin_request(f"config keyring-backend test")
    FLAGS = "--home %HOME% --node %RPC% --chain-id %CHAIN_ID% --yes --output=json --gas=auto --gas-adjustment=2.0"

    dao_proposal_single_code_id = store_contract(
        f"../contracts/dao_proposal_single.wasm"
    )
    print(dao_proposal_single_code_id)

    dao_voting_native_staked_code_id = store_contract(
        f"../contracts/dao_voting_native_staked.wasm"
    )
    print(dao_voting_native_staked_code_id)

    dao_core_code_id = store_contract(f"../contracts/dao_core.wasm")
    print(dao_core_code_id)

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
    ENCODED_PROP_MESSAGE = encode(MODULE_MSG)
    print(ENCODED_PROP_MESSAGE)

    VOTING_MSG = '{"owner":{"core_module":{}},"denom":"ujuno"}'
    ENCODED_VOTING_MESSAGE = encode(VOTING_MSG)
    print(ENCODED_VOTING_MESSAGE)

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

    print(CW_CORE_INIT)

    tx = f"tx wasm instantiate {dao_core_code_id} {CW_CORE_INIT} --label=dao_core --from {KEY_NAME} --no-admin {FLAGS}"
    addr = instantiate_contract(tx)
    print(addr)


if __name__ == "__main__":
    attempt_3()
