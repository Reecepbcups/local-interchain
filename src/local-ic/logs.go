package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MainLogs struct {
	StartTime uint64       `json:"start-time"`
	Chains    []LogOutput  `json:"chains"`
	Channels  []IBCChannel `json:"ibc-channels"`
}
type LogOutput struct {
	ChainID     string   `json:"chain-id"`
	ChainName   string   `json:"chain-name"`
	RPCAddress  string   `json:"rpc-address"`
	GRPCAddress string   `json:"grpc-address"`
	IBCPath     []string `json:"ibc-paths"`
}

func WriteRunningChains(configsDir string, bz []byte) {
	filepath := filepath.Join(configsDir, "configs", "logs.json")
	_ = os.WriteFile(filepath, bz, 0644)
}

func DumpChainsInfoToLogs(configDir string, config *Config, chains []ibc.Chain, connections []IBCChannel) (*cosmos.CosmosChain, int) {
	// This may be un-needed.
	var longestTTLChain *cosmos.CosmosChain
	ttlWait := 0

	mainLogs := MainLogs{
		StartTime: uint64(time.Now().Unix()),
		Chains:    []LogOutput{},
		Channels:  connections,
	}

	// Iterate chain config & get the ibc chain's to save data to logs.
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)

		ibcPaths := chain.IBCPaths
		if ibcPaths == nil {
			ibcPaths = []string{}
		}

		// TODO: save another log for relayer info instead?
		log := LogOutput{
			// TODO: Rest API Address?
			ChainID:     chainObj.Config().ChainID,
			ChainName:   chainObj.Config().Name,
			RPCAddress:  chainObj.GetHostRPCAddress(),
			GRPCAddress: chainObj.GetHostGRPCAddress(),
			IBCPath:     ibcPaths,
		}

		if chain.BlocksTTL > ttlWait {
			ttlWait = chain.BlocksTTL
			longestTTLChain = chainObj
		}

		mainLogs.Chains = append(mainLogs.Chains, log)
	}

	bz, _ := json.MarshalIndent(mainLogs, "", "  ")
	WriteRunningChains(configDir, []byte(bz))

	return longestTTLChain, ttlWait
}

// == Zap Logger ==
func getLoggerConfig() zap.Config {
	config := zap.NewDevelopmentConfig()

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config
}

func InitLogger() (*zap.Logger, error) {
	config := getLoggerConfig()
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
