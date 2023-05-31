"""
Steps:
- Download contracts from the repo based off release version
- Ensure ictest is started for a chain
- Upload to JUNO (store) and init
- Connect together or what ever'
- Profit
"""

import json
import os
import time
from dataclasses import dataclass

import httpx

from test_rest import Request, bin_request, docker_command, query_request

current_dir = os.path.dirname(os.path.realpath(__file__))
with open(os.path.join(current_dir, "configs", "server.json")) as f:
    server_config = json.load(f)["server"]

PORT = server_config["port"]
HOST = server_config["host"]
URL = f"http://{HOST}:{PORT}/"

res = bin_request("keys list --keyring-backend=test")
print(res)

# use https://github.com/DA0-DA0/dao-contracts/tree/main/scripts as a base
