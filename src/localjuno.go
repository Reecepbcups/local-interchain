package main

import (
	"context"
	"os"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestLocalChains runs local IBC chain(s) easily.

func getLoggerConfig() zap.Config {
	config := zap.NewDevelopmentConfig()

	// Customize the configuration according to your needs
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config
}

func initLogger() (*zap.Logger, error) {
	config := getLoggerConfig()
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// remove testing, use go context from here. then put into main()
// https://github.com/cosmos/relayer/blob/main/cmd/start.go#L161
func LocalChains() {

	// create a ctx with a cancel func on completion of a sigkill
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := initLogger()
	if err != nil {
		panic(err)
	}

	chainCfgFile := os.Getenv("CONFIG")
	config, err := LoadConfig(chainCfgFile)
	if err != nil {
		panic(err)
	}

	WriteRunningChains([]byte("{}"))

	// ibc-path-name -> index of []cosmos.CosmosChain
	ibcpaths := make(map[string][]int)
	chainSpecs := []*interchaintest.ChainSpec{}

	for idx, cfg := range config.Chains {
		// if cfg.Debugging {
		// 	t.Logf("[%d] %v", idx, cfg)
		// }

		_, chainSpec := CreateChainConfigs(cfg)
		chainSpecs = append(chainSpecs, chainSpec)

		if len(cfg.IBCPaths) > 0 {
			// t.Log("IBC Path:", cfg.IBCPaths, "Chain:", cfg.Name)

			for _, path := range cfg.IBCPaths {
				ibcpaths[path] = append(ibcpaths[path], idx)
			}
		}
	}

	if err := VerifyIBCPaths(ibcpaths); err != nil {
		logger.Fatal("VerifyIBCPaths", zap.Error(err))
	}

	// Create chain factory for all the chains
	cf := interchaintest.NewBuiltinChainFactory(logger, chainSpecs)

	// Get chains from the chain factory
	name := "LocalChains" + chainCfgFile
	chains, err := cf.Chains(name)
	if err != nil {
		logger.Fatal("cf.Chains", zap.Error(err))
	}

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}
	ic.AdditionalGenesisWallets = SetupGenesisWallets(config, chains)

	fakeTest := FakeT{name: name}

	fakeTesting2 := FakeTesting{
		name: name,
	}

	// Base setup
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(&fakeTest) // this could just impl the test name. no need to require all of t here.

	client, network := interchaintest.DockerSetup(fakeTesting2)

	// setup a relayer if we have IBC paths to use, then use a relayer
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
		relayer = rf.Build(fakeTesting2, client, network)
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
	// TODO: kill this later & cleanup all docker containers. (maybe add a /kill-switch endpoint.)
	go StartNonBlockingServer(ctx, config, vals)

	AddGenesisKeysToKeyring(ctx, config, chains)

	// run commands for each server after startup. Iterate chain configs
	PostStartupCommands(ctx, logger, config, chains)

	connections := GetChannelConnections(ctx, ibcpaths, chains, ic, relayer, eRep)

	// Save to logs.json file for runtime chain information.
	longestTTLChain, ttlWait := DumpChainsInfoToLogs(logger, config, chains, connections)

	// TODO: Way for us to wait for blocks & show the tx logs during this time for each block?
	logger.Info("Waiting for blocks", zap.Int("blocks", ttlWait), zap.String("chain", longestTTLChain.Config().ChainID))

	if err = testutil.WaitForBlocks(ctx, ttlWait, longestTTLChain); err != nil {
		logger.Fatal("testutil.WaitForBlocks", zap.Error(err))
	}

	// t.Cleanup(func() {
	// 	_ = ic.Close()
	// 	WriteRunningChains([]byte("{}"))
	// })

	// wait for ctx to be done, and if so run close()
	<-ctx.Done()
	_ = ic.Close()
	WriteRunningChains([]byte("{}"))

}
