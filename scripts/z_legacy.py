import os

import httpx
from util_base import contracts_path


def download_deps():
    files = [
        "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw20_base.wasm",
        "https://github.com/CosmWasm/cw-plus/releases/latest/download/cw4_group.wasm",
        "https://github.com/CosmWasm/cw-nfts/releases/latest/download/cw721_base.wasm",
    ]

    # download the file to the contracts/ folder
    for url in files:
        name = url.split("/")[-1]
        file_path = os.path.join(contracts_path, name)

        if not os.path.exists(file_path):
            print(f"Downloading {name} to {file_path}")
            r = httpx.get(url, allow_redirects=True)
            with open(file_path, "wb") as f:
                f.write(r.content)


def download_contracts():
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
        name, codeId = file.split(" ")
        os.system(f"junod q wasm code {codeId} {contracts_path}/{name}")
