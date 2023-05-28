package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestLocalChains runs local IBC chain(s) easily.
func TestLocalChains(t *testing.T) {
	config, err := LoadConfig()
	require.NoError(t, err)

	WriteRunningChains([]byte("[]"))

	// ibc-path-name -> index of []cosmos.CosmosChain
	ibcpaths := make(map[string][]int)
	chainSpecs := []*interchaintest.ChainSpec{}

	for idx, cfg := range config.Chains {
		if cfg.Debugging {
			t.Logf("[%d] %v", idx, cfg)
		}

		chainConfig := ibc.ChainConfig{
			Type:                cfg.ChainType,
			Name:                cfg.Name,
			ChainID:             cfg.ChainID,
			Bin:                 cfg.Binary,
			Bech32Prefix:        cfg.Bech32Prefix,
			Denom:               cfg.Denom,
			CoinType:            fmt.Sprintf("%d", cfg.CoinType),
			GasPrices:           cfg.GasPrices,
			GasAdjustment:       cfg.GasAdjustment,
			TrustingPeriod:      cfg.TrustingPeriod,
			NoHostMount:         false,
			ModifyGenesis:       cosmos.ModifyGenesis(cfg.Genesis.Modify),
			ConfigFileOverrides: nil,
			EncodingConfig:      NewEncoding(cfg.EncodingOptions),
		}

		chainConfig.Images = []ibc.DockerImage{{
			Repository: cfg.DockerImage.Repository,
			Version:    cfg.DockerImage.Version,
			UidGid:     cfg.DockerImage.UidGid,
		}}

		chainSpecs = append(chainSpecs, &interchaintest.ChainSpec{
			Name:          cfg.Name,
			Version:       cfg.DockerImage.Version,
			ChainName:     cfg.ChainID,
			ChainConfig:   chainConfig,
			NumValidators: &cfg.NumberVals,
			NumFullNodes:  &cfg.NumberNode,
		})

		if cfg.IBCPath != "" {
			t.Log("IBC Path:", cfg.IBCPath, "Chain:", cfg.Name)
			ibcpaths[cfg.IBCPath] = append(ibcpaths[cfg.IBCPath], idx)
		}
	}

	// ensure that none of ibcpaths values are length > 2
	for k, v := range ibcpaths {
		if len(v) == 1 {
			t.Fatalf("ibc path '%s' has only 1 chain", k)
		}
		if len(v) > 2 {
			t.Fatalf("ibc path '%s' has more than 2 chains", k)
		}
	}

	// Create chain factory for all the chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), chainSpecs)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

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

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	for _, chain := range chains {
		// fmt.Println("adding chain...", chain.Config().Name)
		ic = ic.AddChain(chain)
	}
	ic.AdditionalGenesisWallets = additionalWallets

	// Base setup
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	// setup a relayer if we have IBC paths to use
	if len(ibcpaths) > 0 {
		// relayer
		// Get a relayer instance

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

		r := rf.Build(t, client, network)

		ic = ic.AddRelayer(r, relayerName)

		// Add links between chains
		for path, c := range ibcpaths {
			interLink := interchaintest.InterchainLink{
				Chain1:  nil,
				Chain2:  nil,
				Path:    path,
				Relayer: r,
			}

			// set chain1 & chain2
			for idx, chain := range c {
				if idx == 0 {
					interLink.Chain1 = chains[chain]
				} else {
					interLink.Chain2 = chains[chain]
				}
			}

			ic = ic.AddLink(interLink)
		}
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

	// wait for blocks
	var outputLogs []LogOutput
	var longestTTLChain *cosmos.CosmosChain
	ttlWait := 0
	for idx, chain := range config.Chains {
		chainObj := chains[idx].(*cosmos.CosmosChain)
		t.Logf("\n\n\n\nWaiting for %d blocks on chain %s", chain.BlocksTTL, chainObj.Config().ChainID)

		v := LogOutput{
			// TODO: Rest API Address?
			ChainID:     chainObj.Config().ChainID,
			ChainName:   chainObj.Config().Name,
			RPCAddress:  chainObj.GetHostRPCAddress(),
			GRPCAddress: chainObj.GetHostGRPCAddress(),
			IBCPath:     chain.IBCPath,
		}

		if chain.BlocksTTL > ttlWait {
			ttlWait = chain.BlocksTTL
			longestTTLChain = chainObj
		}

		outputLogs = append(outputLogs, v)
	}

	// dump output logs to file
	bz, _ := json.MarshalIndent(outputLogs, "", "  ")
	WriteRunningChains([]byte(bz))

	// TODO: Way for us to wait for blocks & show the tx logs during this time for each block?
	if err = testutil.WaitForBlocks(ctx, ttlWait, longestTTLChain); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = ic.Close()
		WriteRunningChains([]byte("[]"))
	})
}
