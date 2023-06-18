package types

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