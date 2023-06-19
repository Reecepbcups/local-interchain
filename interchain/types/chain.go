package types

import (
	"math"

	"github.com/go-playground/validator"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type Chain struct {
	// ibc chain config (optional)
	ChainType            string `json:"chain-type" validate:"min=1"`
	CoinType             int    `json:"coin-type" validate:"gt=0"`
	Binary               string `json:"binary" validate:"min=1"`
	Bech32Prefix         string `json:"bech32-prefix" validate:"min=1"`
	Denom                string `json:"denom" validate:"min=1"`
	TrustingPeriod       string `json:"trusting-period"`
	Debugging            bool   `json:"debugging"`
	UseNewGenesisCommand bool   `json:"use-new-genesis-command"`

	// Required
	Name    string `json:"name" validate:"min=1"`
	ChainID string `json:"chain-id" validate:"min=3"`

	DockerImage DockerImage `json:"docker-image" validate:"url"`

	GasPrices     string   `json:"gas-prices"`
	GasAdjustment float64  `json:"gas-adjustment"`
	NumberVals    int      `json:"number-vals" validate:"gte=1"`
	NumberNode    int      `json:"number-node"`
	BlocksTTL     int      `json:"blocks-ttl"`
	IBCPaths      []string `json:"ibc-paths"`
	Genesis       Genesis  `json:"genesis"`
}

func (chain *Chain) Validate() error {
	validate := validator.New()
	return validate.Struct(chain)
}

func (chain *Chain) SetChainDefaults() {
	if chain.BlocksTTL <= 0 {
		chain.BlocksTTL = math.MaxInt32
	}

	if chain.ChainType == "" {
		chain.ChainType = "cosmos"
	}

	if chain.CoinType == 0 {
		chain.CoinType = 118
	}

	if chain.DockerImage.UidGid == "" {
		chain.DockerImage.UidGid = "1025:1025"
	}

	if chain.NumberVals == 0 {
		chain.NumberVals = 1
	}

	if chain.TrustingPeriod == "" {
		chain.TrustingPeriod = "112h"
	}

	if chain.IBCPaths == nil {
		chain.IBCPaths = []string{}
	}

	// Genesis
	if chain.Genesis.StartupCommands == nil {
		chain.Genesis.StartupCommands = []string{}
	}
	if chain.Genesis.Accounts == nil {
		chain.Genesis.Accounts = []GenesisAccount{}
	}
	if chain.Genesis.Modify == nil {
		chain.Genesis.Modify = []cosmos.GenesisKV{}
	}

	// TODO: Error here instead?
	if chain.Binary == "" {
		panic("'binary' is required in your config for " + chain.ChainID)
	}
	if chain.Denom == "" {
		panic("'denom' is required in your config for " + chain.ChainID)
	}
	if chain.Bech32Prefix == "" {
		panic("'bech32-prefix' is required in your config for " + chain.ChainID)
	}
}
