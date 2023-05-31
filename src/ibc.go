package main

import (
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

func VerifyIBCPaths(ibcpaths map[string][]int) error {
	for k, v := range ibcpaths {
		if len(v) == 1 {
			return fmt.Errorf("ibc path '%s' has only 1 chain", k)
		}
		if len(v) > 2 {
			return fmt.Errorf("ibc path '%s' has more than 2 chains", k)
		}
	}
	return nil
}

// TODO: Allow for a single chain to IBC between multiple chains
func LinkIBCPaths(ibcpaths map[string][]int, chains []ibc.Chain, ic *interchaintest.Interchain, r ibc.Relayer) {
	for path, c := range ibcpaths {
		interLink := interchaintest.InterchainLink{
			Chain1:  chains[c[0]],
			Chain2:  chains[c[1]],
			Path:    path,
			Relayer: r,
		}

		ic = ic.AddLink(interLink)
	}
}
