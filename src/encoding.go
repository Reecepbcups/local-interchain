package main

import (
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	// Encoding Types
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"

	// IBC
	ibccoreclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibccoretypes "github.com/cosmos/ibc-go/v7/modules/core/types"
)

func LegacyJunoEncoding() *sdktestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func NewEncoding(options []string) *sdktestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// This way we can support multiple networks with different modules
	for _, option := range options {
		if option == "juno" {
			// Contains all the sub types as well. maybe use a builder for this.
			feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
			tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
			wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
			ibccoretypes.RegisterInterfaces(cfg.InterfaceRegistry)
			ibccoreclienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
		}

		if option == "wasm" {
			wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
		}

		if option == "ibc" {
			ibccoretypes.RegisterInterfaces(cfg.InterfaceRegistry)
			ibccoreclienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
		}

		// When v47:
		// osmosis
		// stargaze
		// terra
	}

	return &cfg
}
