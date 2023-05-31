package main

import (
	"context"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

// Used for quick script testing

// go run main.go rest_server.go
func main() {
	config, _ := LoadConfig("chains.json")
	fmt.Println(config)

	ctx := context.Background()
	StartNonBlockingServer(ctx, config, make(map[string]*cosmos.ChainNode))
}
