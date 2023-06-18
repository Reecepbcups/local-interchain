package interchain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	types "github.com/reecepbcups/localinterchain/interchain/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func WriteRunningChains(configsDir string, bz []byte) {
	filepath := filepath.Join(configsDir, "configs", "logs.json")
	_ = os.WriteFile(filepath, bz, 0644)
}

func DumpChainsInfoToLogs(configDir string, config *types.Config, chains []ibc.Chain, connections []types.IBCChannel) (*cosmos.CosmosChain, int) {
	// This may be un-needed.
	var longestTTLChain *cosmos.CosmosChain
	ttlWait := 0

	mainLogs := types.MainLogs{
		StartTime: uint64(time.Now().Unix()),
		Chains:    []types.LogOutput{},
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
		log := types.LogOutput{
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
