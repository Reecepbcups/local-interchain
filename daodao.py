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
from dataclasses import dataclass

import httpx

from test_rest import Request, bin_request, docker_command, query_request

KEY_NAME = "acc0"

current_dir = os.path.dirname(os.path.realpath(__file__))
parent_dir = os.path.dirname(current_dir)
with open(os.path.join(current_dir, "configs", "server.json")) as f:
    server_config = json.load(f)["server"]

PORT = server_config["port"]
HOST = server_config["host"]
URL = f"http://{HOST}:{PORT}/"

contracts_path = os.path.join(current_dir, "contracts")
if not os.path.exists(contracts_path):
    os.mkdir(contracts_path)


def store_contract(rel_file_path: str, asJSON: bool = False) -> str | dict:
    print(f"Uploading {rel_file_path}")

    data = {
        "chain-id": "localjuno-1",
        "key-name": KEY_NAME,
        "file-name": rel_file_path,
    }
    r = httpx.post(
        f"{URL}upload",
        json=data,
        headers={"Content-Type": "application/json"},
        timeout=120,
    )
    if asJSON:
        return json.loads(r.text)

    return r.text


# From https://github.com/DA0-DA0/dao-contracts/releases


def download_contracts():
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
        name, codeId = file.split(" ")
        os.system(f"junod q wasm code {codeId} {current_dir}/contracts/{name}")


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
        "pre_propose_info": {"AnyoneMayPropose": {}},
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
    daoCoreCode = 0
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

    tx = f"tx wasm execute {adminFactory} '" + CW_CORE_INIT + f"' --from acc0 {FLAGS}"
    print(tx)
    # exit(1)
    res = bin_request(tx)
    print(res)
    time.sleep(2.5)
    # addr = getContractAddr(json.loads(res)["txhash"])
    # print(addr)

    pass


if __name__ == "__main__":
    new()
