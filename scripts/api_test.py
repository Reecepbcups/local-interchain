"""
This test the rest server to ensures it functions properly.

You could also do in curl
curl http://localhost:8080/ --include --header "Content-Type: application/json" -X POST --data '{"chain_id":"localjuno-1","action":"q","cmd":"bank total"}'
"""


from util_base import API_URL
from util_req import RequestBase, RequestType, send_request
from util_txs import query_tx

CHAIN_ID = "localjuno-1"
bin_base = RequestBase(API_URL, CHAIN_ID, RequestType.BIN)
query_base = RequestBase(API_URL, CHAIN_ID, RequestType.QUERY)

from ref_types.transactions import RequestBuilder


def main():
    bin_test()
    # tx_test()


# Test to ensure the base layer works and returns data properly
def bin_test():
    # TODO: option to toggle printing in send_request for all? (off by default)
    rb = RequestBuilder(API_URL, CHAIN_ID)

    rb.bin("keys list --keyring-backend=test --output=json")
    rb.bin(
        "tx decode ClMKUQobL2Nvc21vcy5nb3YudjFiZXRhMS5Nc2dWb3RlEjIIpwISK2p1bm8xZGM3a2MyZzVrZ2wycmdmZHllZGZ6MDl1YTlwZWo1eDNsODc3ZzcYARJmClAKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECxjGMmYp4MlxxfFWi9x4u+jOleJVde3Cru+HnxAVUJmgSBAoCCH8YNBISCgwKBXVqdW5vEgMyMDQQofwEGkDPE4dCQ4zUh6LIB9wqNXDBx+nMKtg0tEGiIYEH8xlw4H8dDQQStgAe6xFO7I/oYVSWwa2d9qUjs9qyB8r+V0Gy"
    )

    rb.bin("config keyring-backend test")
    rb.bin("config node %RPC%")
    rb.bin("config")

    rb.bin("keys list --output=json")

    rb.query("bank total")
    rb.query("bank balances juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 --output=json")


# Test to ensure Transactions and getting that data returns properly
def tx_test():
    res = send_request(
        bin_base,
        "tx bank send acc0 juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 500ujuno --fees 5000ujuno --node %RPC% --chain_id %CHAIN_ID% --yes --output json",
    )
    print(res)

    tx_data = query_tx(query_base, res)
    print(tx_data)

    res = send_request(
        query_base,
        "bank balances juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0 --output=json",
    )
    print(res)


if __name__ == "__main__":
    main()
