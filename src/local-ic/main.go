package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"go.uber.org/zap"
)

// Makefile values
var (
	Version, InstallDirectory string
)

// TestLocalChains runs local IBC chain(s) easily.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Logger for ICTest functions only.
	logger, err := InitLogger()
	if err != nil {
		panic(err)
	}

	configDir := getDirectory()
	chainCfgFile := os.Getenv("CONFIG")
	config, err := LoadConfig(configDir, chainCfgFile)
	if err != nil {
		panic(err)
	}

	WriteRunningChains(configDir, []byte("{}"))

	// ibc-path-name -> index of []cosmos.CosmosChain
	ibcpaths := make(map[string][]int)
	chainSpecs := []*interchaintest.ChainSpec{}

	for idx, cfg := range config.Chains {
		_, chainSpec := CreateChainConfigs(cfg)
		chainSpecs = append(chainSpecs, chainSpec)

		if len(cfg.IBCPaths) > 0 {
			for _, path := range cfg.IBCPaths {
				ibcpaths[path] = append(ibcpaths[path], idx)
			}
		}
	}

	if err := VerifyIBCPaths(ibcpaths); err != nil {
		log.Fatal("VerifyIBCPaths", err)
	}

	// Create chain factory for all the chains
	cf := interchaintest.NewBuiltinChainFactory(logger, chainSpecs)

	// Get chains from the chain factory
	name := "ic-" + strings.ReplaceAll(chainCfgFile, ".json", "")
	chains, err := cf.Chains(name)
	if err != nil {
		log.Fatal("cf.Chains", err)
	}

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}
	ic.AdditionalGenesisWallets = SetupGenesisWallets(config, chains)

	fakeT := FakeTesting{
		name: name,
	}

	// Base setup
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(&fakeT)

	client, network := interchaintest.DockerSetup(fakeT)

	// setup a relayer if we have IBC paths to use.
	var relayer ibc.Relayer
	if len(ibcpaths) > 0 {
		rlyCfg := config.Relayer

		relayerType, relayerName := ibc.CosmosRly, "relay"
		rf := interchaintest.NewBuiltinRelayerFactory(
			relayerType,
			logger,
			interchaintestrelayer.CustomDockerImage(
				rlyCfg.DockerImage.Repository,
				rlyCfg.DockerImage.Version,
				rlyCfg.DockerImage.UidGid,
			),
			interchaintestrelayer.StartupFlags(rlyCfg.StartupFlags...),
		)

		// This also just needs the name.
		relayer = rf.Build(fakeT, client, network)
		ic = ic.AddRelayer(relayer, relayerName)

		// Add links between chains
		LinkIBCPaths(ibcpaths, chains, ic, relayer)
	}

	// Build all chains & begin.
	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         name,
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	if err != nil {
		logger.Fatal("ic.Build", zap.Error(err))
	}

	// keys as well?
	vals := make(map[string]*cosmos.ChainNode)
	for _, chain := range chains {
		if cosmosChain, ok := chain.(*cosmos.CosmosChain); ok {
			chainID := cosmosChain.Config().ChainID
			vals[chainID] = cosmosChain.Validators[0]
		}
	}

	// Starts a non blocking REST server to take action on the chain.
	// TODO: kill this later & cleanup all docker containers. (maybe add a /kill-switch endpoint?)
	go StartNonBlockingServer(ctx, config, vals, configDir)

	AddGenesisKeysToKeyring(ctx, config, chains)

	// run commands for each server after startup. Iterate chain configs
	PostStartupCommands(ctx, config, chains)

	connections := GetChannelConnections(ctx, ibcpaths, chains, ic, relayer, eRep)

	// Save to logs.json file for runtime chain information.
	longestTTLChain, ttlWait := DumpChainsInfoToLogs(configDir, config, chains, connections)

	// TODO: Way for us to wait for blocks & show the tx logs during this time for each block?
	log.Println("Waiting for blocks", ttlWait, longestTTLChain.Config().ChainID)

	// Do with context? https://github.com/cosmos/relayer/blob/main/cmd/start.go#L161

	if err = testutil.WaitForBlocks(ctx, ttlWait, longestTTLChain); err != nil {
		log.Fatal("testutil.WaitForBlocks", err)
	}
}

func getDirectory() string {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir != "" {
		return configDir
	}

	return InstallDirectory
}
