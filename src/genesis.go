package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

func AddGenesisKeysToKeyring(ctx context.Context, config *MainConfig, chains []ibc.Chain) {
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		for _, acc := range chain.Genesis.Accounts {
			chainObj.RecoverKey(ctx, acc.Name, acc.Mnemonic)
		}
	}
}

func PostStartupCommands(ctx context.Context, t *testing.T, config *MainConfig, chains []ibc.Chain) {
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		for _, cmd := range chain.Genesis.StartupCommands {
			t.Log("\nRunning startup command on", chainObj.Config().ChainID, "-->", fmt.Sprintf("`%s`", cmd))

			cmd = strings.ReplaceAll(cmd, "%HOME%", chainObj.Validators[0].HomeDir())
			cmd = strings.ReplaceAll(cmd, "%CHAIN_ID%", chainObj.Config().ChainID)

			stdout, stderr, err := chainObj.Exec(ctx, strings.Split(cmd, " "), []string{})

			output := stdout
			if len(output) == 0 {
				output = stderr
			} else if err != nil {
				t.Fatalf("Error running startup command: %s\n%s", cmd, output)
			}

			t.Log(string(output))
		}
	}
}

func SetupGenesisWallets(config *MainConfig, chains []ibc.Chain) map[ibc.Chain][]ibc.WalletAmount {
	// iterate all chains chain's configs & setup accounts
	additionalWallets := make(map[ibc.Chain][]ibc.WalletAmount)
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		for _, acc := range chain.Genesis.Accounts {
			amount, err := sdk.ParseCoinsNormalized(acc.Amount)
			if err != nil {
				panic(err)
			}

			for _, coin := range amount {
				additionalWallets[chainObj] = append(additionalWallets[chainObj], ibc.WalletAmount{
					Address: acc.Address,
					Amount:  coin.Amount.Int64(),
					Denom:   coin.Denom,
				})
			}
		}
	}
	return additionalWallets
}
