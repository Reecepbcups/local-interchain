package main

import (
	"context"

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
