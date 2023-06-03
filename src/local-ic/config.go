package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/reecepbcups/localinterchain/src/util"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

type Config struct {
	Chains  []Chain    `json:"chains"`
	Relayer Relayer    `json:"relayer"`
	Server  RestServer `json:"server"`
}

type Chain struct {
	// ibc chain config (optional)
	ChainType            string   `json:"chain-type"`
	CoinType             int      `json:"coin-type"`
	Binary               string   `json:"binary"`
	Bech32Prefix         string   `json:"bech32-prefix"`
	Denom                string   `json:"denom"`
	TrustingPeriod       string   `json:"trusting-period"`
	Debugging            bool     `json:"debugging"`
	EncodingOptions      []string `json:"encoding-options"`
	UseNewGenesisCommand bool     `json:"use-new-genesis-command"`

	// Required
	Name    string `json:"name"`
	ChainID string `json:"chain-id"`

	DockerImage DockerImage `json:"docker-image"`

	GasPrices     string   `json:"gas-prices"`
	GasAdjustment float64  `json:"gas-adjustment"`
	NumberVals    int      `json:"number-vals"`
	NumberNode    int      `json:"number-node"`
	BlocksTTL     int      `json:"blocks-ttl"`
	IBCPaths      []string `json:"ibc-paths"`
	Genesis       Genesis  `json:"genesis"`
}

type Relayer struct {
	DockerImage  DockerImage `json:"docker-image"`
	StartupFlags []string    `json:"startup-flags"`
}

type RestServer struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type DockerImage struct {
	Repository string `json:"repository"`
	Version    string `json:"version"`
	UidGid     string `json:"uid-gid"`
}

type Genesis struct {
	// Only apart of my fork for now.
	Modify []cosmos.GenesisKV `json:"modify"` // 'key' & 'val' in the config.

	Accounts []struct {
		Name     string `json:"name"`
		Amount   string `json:"amount"`
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	} `json:"accounts"`

	// A list of commands which run after chains are good to go.
	// May need to move out of genesis into its own section? Seems silly though.
	StartupCommands []string `json:"startup-commands"`
}

func loadConfig(config *Config, filepath string) (*Config, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadConfig(configDirectory, chainCfgFile string) (*Config, error) {
	var config *Config

	configFile := "chains.json"
	if chainCfgFile != "" {
		configFile = chainCfgFile
	}

	configsDir := filepath.Join(configDirectory, "configs")

	cfgFilePath := filepath.Join(configsDir, configFile)
	relayerFilePath := filepath.Join(configsDir, "relayer.json")
	serverFilePath := filepath.Join(configsDir, "server.json")

	config, err := loadConfig(config, cfgFilePath)
	if err != nil {
		return nil, err
	}
	config, _ = loadConfig(config, relayerFilePath)
	config, _ = loadConfig(config, serverFilePath)

	fmt.Printf("Loaded %v\n", config)

	chains := config.Chains
	relayer := config.Relayer

	for i := range chains {
		chain := chains[i]
		chain.setChainDefaults()
		// Replace all string instances of %DENOM% with the chain's denom.
		// Even in nested structs, slices/arrays, etc.
		util.ReplaceStringValues(&chain, "%DENOM%", chain.Denom)
		util.ReplaceStringValues(&chain, "%BIN%", chain.Binary)
		util.ReplaceStringValues(&chain, "%CHAIN_ID%", chain.ChainID)

		chains[i] = chain
	}

	config.Relayer = relayer.setRelayerDefaults()

	return config, nil
}

func CreateChainConfigs(cfg Chain) (ibc.ChainConfig, *interchaintest.ChainSpec) {
	chainCfg := ibc.ChainConfig{
		Type:                   cfg.ChainType,
		Name:                   cfg.Name,
		ChainID:                cfg.ChainID,
		Bin:                    cfg.Binary,
		Bech32Prefix:           cfg.Bech32Prefix,
		Denom:                  cfg.Denom,
		CoinType:               fmt.Sprintf("%d", cfg.CoinType),
		GasPrices:              cfg.GasPrices,
		GasAdjustment:          cfg.GasAdjustment,
		TrustingPeriod:         cfg.TrustingPeriod,
		NoHostMount:            false,
		ModifyGenesis:          cosmos.ModifyGenesis(cfg.Genesis.Modify),
		ConfigFileOverrides:    nil,
		EncodingConfig:         NewEncoding(cfg.EncodingOptions),
		UsingNewGenesisCommand: cfg.UseNewGenesisCommand,
	}

	if cfg.DockerImage.Version == "" {
		panic("DockerImage.Version is required in your config")
	}

	if cfg.DockerImage.Repository != "" {
		chainCfg.Images = []ibc.DockerImage{
			{
				Repository: cfg.DockerImage.Repository,
				Version:    cfg.DockerImage.Version,
				UidGid:     cfg.DockerImage.UidGid,
			},
		}
	}

	chainSpecs := &interchaintest.ChainSpec{
		Name:          cfg.Name,
		Version:       cfg.DockerImage.Version,
		ChainName:     cfg.ChainID,
		ChainConfig:   chainCfg,
		NumValidators: &cfg.NumberVals,
		NumFullNodes:  &cfg.NumberNode,
	}

	return chainCfg, chainSpecs
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
		panic("'binary' is required in your config for " + chain.ChainID)
	}
	if chain.Denom == "" {
		panic("'denom' is required in your config for " + chain.ChainID)
	}
	if chain.Bech32Prefix == "" {
		panic("'bech32-prefix' is required in your config for " + chain.ChainID)
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
