package main

import (
	"encoding/json"
	"io/ioutil"
	"math"

	"github.com/reecepbcups/localinterchain/src/util"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type MainConfig struct {
	Chains  []Chain `json:"chains"`
	Relayer Relayer `json:"relayer"`
}

type Chain struct {
	// ibc chain config (optional)
	ChainType      string `json:"chain-type"`
	CoinType       int    `json:"coin-type"`
	Binary         string `json:"binary"`
	Bech32Prefix   string `json:"bech32-prefix"`
	Denom          string `json:"denom"`
	TrustingPeriod string `json:"trusting-period"`
	Debugging      bool   `json:"debugging"`

	// Required
	Name            string   `json:"name"`
	ChainID         string   `json:"chain-id"`
	EncodingOptions []string `json:"encoding-options"`

	DockerImage DockerImage `json:"docker-image"`

	GasPrices     string  `json:"gas-prices"`
	GasAdjustment float64 `json:"gas-adjustment"`
	NumberVals    int     `json:"number-vals"`
	NumberNode    int     `json:"number-node"`
	BlocksTTL     int     `json:"blocks-ttl"`
	IBCPath       string  `json:"ibc-path"`
	Genesis       Genesis `json:"genesis"`
}

type Relayer struct {
	DockerImage  DockerImage `json:"docker-image"`
	StartupFlags []string    `json:"startup-flags"`
}

type DockerImage struct {
	Repository string `json:"repository"`
	Version    string `json:"version"`
	UidGid     string `json:"uid-gid"`
}

type Genesis struct {
	// Only apart of my fork for now.
	Modify []cosmos.GenesisKV `json:"modify"` // 'key' & 'val' in the config
	// Modify []struct {
	// 	Key   string `json:"key"`
	// 	Value string `json:"value"`
	// } `json:"modify"`
	Accounts []struct {
		Name     string `json:"name"`
		Amount   string `json:"amount"`
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	} `json:"accounts"`
}

func loadConfig(config *MainConfig, filepath string) (*MainConfig, error) {
	// Load Chains
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadConfig() (*MainConfig, error) {
	var config *MainConfig
	config, _ = loadConfig(config, "../configs/chains.json")
	config, _ = loadConfig(config, "../configs/relayer.json")

	chains := config.Chains
	relayer := config.Relayer

	for i := range chains {
		chain := chains[i]
		chain.setChainDefaults()
		// Replace all string instances of %DENOM% with the chain's denom.
		// Even in nested structs, slices/arrays, etc.
		util.ReplaceStringValues(&chain, "%DENOM%", chain.Denom)
		chains[i] = chain
	}

	config.Relayer = relayer.setRelayerDefaults()

	return config, nil
}

func (chain *Chain) setChainDefaults() {
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

	// TODO: Error here instead?
	if chain.Binary == "" {
		chain.Binary = "junod"
	}
	if chain.Denom == "" {
		chain.Denom = "ujuno"
	}
	if chain.Bech32Prefix == "" {
		chain.Bech32Prefix = "juno"
	}
}

func (r Relayer) setRelayerDefaults() Relayer {
	if r.DockerImage.Repository == "" {
		r.DockerImage.Repository = "ghcr.io/cosmos/relayer"
	}

	if r.DockerImage.Version == "" {
		r.DockerImage.Version = "latest"
	}

	if r.DockerImage.UidGid == "" {
		r.DockerImage.UidGid = "100:1000"
	}

	return r
}
