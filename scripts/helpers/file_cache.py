import hashlib
import json
import os

fp = os.path.realpath(__file__)
root_dir = os.path.dirname(os.path.dirname(os.path.dirname(fp)))
contracts_storage_dir = os.path.join(root_dir, "contracts")

contracts_json_path = os.path.join(root_dir, "configs", "contracts.json")
logs_path = os.path.join(root_dir, "configs", "logs.json")


class Cache:
    @staticmethod
    def default_contracts_json():
        if not os.path.exists(contracts_json_path):
            with open(contracts_json_path, "w") as f:
                f.write(json.dumps({"start_time": 0, "file_cache": {}}))

    @staticmethod
    def get_chain_start_time_from_logs() -> int:
        with open(logs_path, "r") as f:
            logs = dict(json.load(f))

        return int(logs.get("start-time", -1))

    @staticmethod
    def get_cache_or_default(contracts: dict, ictest_chain_start: int) -> dict:
        with open(contracts_json_path, "r") as f:
            cache_time = dict(json.load(f)).get("start_time", 0)

        if cache_time == 0 or cache_time != ictest_chain_start:
            # reset cache, and set cache time to current ictest time
            contracts["start_time"] = ictest_chain_start
            contracts["file_cache"] = {}

            # write to file
            with open(contracts_json_path, "w") as f:
                json.dump(contracts, f, indent=4)

        with open(contracts_json_path, "r") as f:
            contracts = json.load(f)

        return contracts

    @staticmethod
    def update_cache(contracts: dict, code_id: str | int, shaHash: str) -> int:
        contracts["file_cache"][shaHash] = int(code_id)
        with open(contracts_json_path, "w") as f:
            json.dump(contracts, f, indent=4)
        return int(code_id)

    @staticmethod
    def get_file_hash(rel_file_path: str, chainId: str) -> str:
        BUF_SIZE = 65536  # 64k chunks
        sha1 = hashlib.sha1()

        file_path = os.path.join(contracts_storage_dir, rel_file_path)

        if not os.path.exists(file_path):
            raise FileNotFoundError(f"File not found: {file_path}")

        sha1.update(chainId.replace("-", "").encode("utf-8"))
        with open(file_path, "rb") as f:
            while True:
                bz = f.read(BUF_SIZE)
                if not bz:
                    break
                sha1.update(bz)

        return sha1.hexdigest()


# We always run this to start.
Cache.default_contracts_json()
