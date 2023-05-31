package main

import (
	"context"
	"os"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestLocalChains runs local IBC chain(s) easily.
func TestLocalChains(t *testing.T) {
	chainCfgFile := os.Getenv("CHAIN_CONFIG")
	config, err := LoadConfig(chainCfgFile)
	require.NoError(t, err)

	WriteRunningChains([]byte("[]"))

	// ibc-path-name -> index of []cosmos.CosmosChain
	ibcpaths := make(map[string][]int)
	chainSpecs := []*interchaintest.ChainSpec{}

	for idx, cfg := range config.Chains {
		if cfg.Debugging {
			t.Logf("[%d] %v", idx, cfg)
		}

		_, chainSpec := CreateChainConfigs(cfg)
		chainSpecs = append(chainSpecs, chainSpec)

		if len(cfg.IBCPaths) > 0 {
			t.Log("IBC Path:", cfg.IBCPaths, "Chain:", cfg.Name)

			for _, path := range cfg.IBCPaths {
				ibcpaths[path] = append(ibcpaths[path], idx)
			}
		}
	}

	if err := VerifyIBCPaths(ibcpaths); err != nil {
		t.Fatal(err)
	}

	// Create chain factory for all the chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), chainSpecs)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}
	ic.AdditionalGenesisWallets = SetupGenesisWallets(config, chains)

	// Base setup
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	// setup a relayer if we have IBC paths to use, then use a relayer
	var relayer ibc.Relayer
	if len(ibcpaths) > 0 {
		rlyCfg := config.Relayer

		relayerType, relayerName := ibc.CosmosRly, "relay"
		rf := interchaintest.NewBuiltinRelayerFactory(
			relayerType,
			zaptest.NewLogger(t),
			interchaintestrelayer.CustomDockerImage(
				rlyCfg.DockerImage.Repository,
				rlyCfg.DockerImage.Version,
				rlyCfg.DockerImage.UidGid,
			),
			interchaintestrelayer.StartupFlags(rlyCfg.StartupFlags...),
		)

		relayer = rf.Build(t, client, network)
		ic = ic.AddRelayer(relayer, relayerName)

		// Add links between chains
		LinkIBCPaths(ibcpaths, chains, ic, relayer)
	}

	// Build all chains & begin.
	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	// keys as well?
	vals := make(map[string]*cosmos.ChainNode)
	for _, chain := range chains {
		if cosmosChain, ok := chain.(*cosmos.CosmosChain); ok {
			chainID := cosmosChain.Config().ChainID
			vals[chainID] = cosmosChain.Validators[0]
		}
	}

	// Starts a non blocking REST server to take action on the chain.
	// TODO: kill this later & cleanup all docker containers. (maybe add a /kill-switch endpoint.)
	go StartNonBlockingServer(ctx, config, vals)

	AddGenesisKeysToKeyring(ctx, config, chains)

	// run commands for each server after startup. Iterate chain configs
	PostStartupCommands(ctx, t, config, chains)

	connections := GetChannelConnections(ctx, ibcpaths, chains, ic, relayer, eRep)

	// Save to logs.json file for runtime chain information.
	longestTTLChain, ttlWait := DumpChainsInfoToLogs(t, config, chains, connections)

	// TODO: Way for us to wait for blocks & show the tx logs during this time for each block?
	t.Logf("\n\nWaiting for %d blocks on chain %s", ttlWait, longestTTLChain.Config().ChainID)

	if err = testutil.WaitForBlocks(ctx, ttlWait, longestTTLChain); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = ic.Close()
		WriteRunningChains([]byte("[]"))
	})
}
