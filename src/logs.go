package main

import "io/ioutil"

type LogOutput struct {
	ChainID     string `json:"chain-id"`
	ChainName   string `json:"chain-name"`
	RPCAddress  string `json:"rpc-address"`
	GRPCAddress string `json:"grpc-address"`
	IBCPath     string `json:"ibc-path"`
}

func WriteRunningChains(bz []byte) {
	_ = ioutil.WriteFile("../configs/logs.json", bz, 0644)
}
