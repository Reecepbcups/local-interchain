# Local Interchain

A simple way to config and run IBC local chain testing environments using [Strangelove's InterchainTest](https://github.com/strangelove-ventures/interchaintest)

**This will eventually get phased out and brought into the ICTest repo directly.** Just doing it here for ease of creation & for a future simple public archive.

## Installing

**Install on Mac / Linux**
```bash
git clone https://github.com/Reecepbcups/local-interchain.git 

cd local-interchain

# NOTE: your binary will link back to this location of where you install.
# If you rename the folder or move it, you need to `make install` the binary again.
make install
```
**Install on Windows**

Follow [this guide](WINDOWS.md) to setup the Windows OS environment for installing Local Interchain.

## Running

- *(optional)* Edit `./configs/relayer.json`
- Copy: `cp ./configs/chains.json ./configs/mytest1_ignored.json`
- Run: `CONFIG=mytest1_ignored.json local-ic`
- Change directory `INSTALL_DIR=/root/example/local-interchain CONFIG=myother_ignored.json local-ic`

*(Not using 'CONFIG' will default to `./configs/base.json`. You can have as many chain configurations as you desire. Ending the file with `_ignored.json` will ignore it from git)*

---

## Endpoints

- (POST) localhost:8080/
- (POST) localhost:8080/upload
- (GET)  localhost:8080/info


## Helpful Tips

- Auto complete: edit ~/.bashrc or ~/.zshrc and add `source <(local-ic completion bash)` or `source <(local-ic completion zsh)`.
    (fish & windows powershell is also supported)

- After starting the chain(s), you can read the `./configs/logs.json` file to get useful information. This includes the chain's id, name, RPC address, and more.

```json
// ./configs/logs.json
[
  {
    "chain-id": "exampleid-1",
    "chain-name": "example",
    "rpc-address": "http://localhost:38829",
    "grpc-address": "localhost:34917",
    "ibc-paths": []
  }
]
```

- 'ibc-path' should only be set if you are using 2+ chains. If you are using 1 chain, you can leave it blank.

- You can use `%DENOM%` anywhere in the chain's config to use the `denom` line. This is useful for gas prices, Genesis accounts, etc.

- Configuration's have set defaults. The minimum you need to run a single chain is the following

```json
{
    "name": "juno",            
    "chain-id": "localjuno-2",
    "denom": "ujuno",
    "docker-image": {        
        "version": "v14.1.0"
    },    
    "gas-prices": "0%DENOM%",
    "gas-adjustment": 2.0,           
}
```

---

## Base Chain Template

Here is a base chain template with every feature the configuration accepts. Accounts have extra data to make it simpler for scripting and read from the file directly.

```json
{
    "name": "juno",            
    "chain-id": "localjuno-1",
    "denom": "ujuno",
    "binary": "junod",
    "bech32-prefix": "juno",
    "docker-image": {
        "repository": "ghcr.io/cosmoscontracts/juno-e2e",
        "version": "v14.1.0",
        "uid-gid": "1000:1000"
    },
    "chain-type": "cosmos",
    "coin-type": 118,
    "trusting-period": "112h",
    "gas-prices": "0%DENOM%",
    "gas-adjustment": 2.0,
    "number-vals": 1,
    "number-node": 0,
    "blocks-ttl": -1,
    "use-new-genesis-command": false,
    "ibc-paths": ["juno-ibc-1"],
    "debugging": true,
    "encoding-options": ["juno"],
    "genesis": {
        "modify": [
            {
                "key": "app_state.gov.voting_params.voting_period",
                "val": "15s"
            },
            {
                "key": "app_state.gov.deposit_params.max_deposit_period",
                "val": "15s"
            },
            {
                "key": "app_state.gov.deposit_params.min_deposit.0.denom",
                "val": "ujuno"
            }
        ],     
        "accounts": [
            {
                "name": "acc0",
                "address": "juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk",
                "amount": "10000000%DENOM%",
                "mnemonic": "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
            }
        ]                
    }
},
```
