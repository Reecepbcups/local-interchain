package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

type MainLogs struct {
	Chains   []LogOutput  `json:"chains"`
	Channels []IBCChannel `json:"ibc-channels"`
}
type LogOutput struct {
	ChainID     string   `json:"chain-id"`
	ChainName   string   `json:"chain-name"`
	RPCAddress  string   `json:"rpc-address"`
	GRPCAddress string   `json:"grpc-address"`
	IBCPath     []string `json:"ibc-paths"`
}

const filename = "../configs/logs.json"

func WriteRunningChains(bz []byte) {
	_ = ioutil.WriteFile(filename, bz, 0644)
}

func DumpChainsInfoToLogs(t *testing.T, config *MainConfig, chains []ibc.Chain, connections []IBCChannel) (*cosmos.CosmosChain, int) {
	// This may be un-needed.
	var longestTTLChain *cosmos.CosmosChain
	ttlWait := 0

	mainLogs := MainLogs{
		Chains:   []LogOutput{},
		Channels: connections,
	}

	// Iterate chain config & get the ibc chain's to save data to logs.
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		// TODO: save another log for relayer info instead?
		log := LogOutput{
			// TODO: Rest API Address?
			ChainID:     chainObj.Config().ChainID,
			ChainName:   chainObj.Config().Name,
			RPCAddress:  chainObj.GetHostRPCAddress(),
			GRPCAddress: chainObj.GetHostGRPCAddress(),
			IBCPath:     chain.IBCPaths,
		}

		if chain.BlocksTTL > ttlWait {
			ttlWait = chain.BlocksTTL
			longestTTLChain = chainObj
		}

		mainLogs.Chains = append(mainLogs.Chains, log)
	}

	bz, _ := json.MarshalIndent(mainLogs, "", "  ")
	WriteRunningChains([]byte(bz))

	return longestTTLChain, ttlWait
}
