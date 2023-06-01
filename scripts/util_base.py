import json
import os

current_dir = os.path.dirname(os.path.realpath(__file__))
parent_dir = os.path.dirname(current_dir)

contracts_path = os.path.join(parent_dir, "contracts")

contracts_json_path = os.path.join(parent_dir, "configs", "contracts.json")


def default_contracts_json():
    if not os.path.exists(contracts_json_path):
        with open(contracts_json_path, "w") as f:
            f.write(json.dumps({"start_time": 0, "file_cache": {}}))


default_contracts_json()


# create contracts folder if not already
if not os.path.exists(contracts_path):
    os.mkdir(contracts_path)

server_config = {}
with open(os.path.join(parent_dir, "configs", "server.json")) as f:
    server_config = json.load(f)["server"]

PORT = server_config["port"]
HOST = server_config["host"]
URL = f"http://{HOST}:{PORT}/"
URL_UPLOAD = f"http://{HOST}:{PORT}/upload"
