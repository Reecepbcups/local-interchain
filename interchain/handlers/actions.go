package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/reecepbcups/localinterchain/interchain/util"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type actions struct {
	ctx  context.Context
	vals map[string]*cosmos.ChainNode
}

func NewActions(ctx context.Context, vals map[string]*cosmos.ChainNode) *actions {
	return &actions{
		ctx:  ctx,
		vals: vals,
	}
}

type ActionHandler struct {
	ChainId string `json:"chain_id"`
	Action  string `json:"action"`
	Cmd     string `json:"cmd"`
}

func (a *actions) PostActions(w http.ResponseWriter, r *http.Request) {
	var ah ActionHandler
	err := json.NewDecoder(r.Body).Decode(&ah)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	chainId := ah.ChainId
	action := ah.Action
	if _, ok := a.vals[chainId]; !ok {
		util.Write(w, []byte(fmt.Sprintf(`{"error":"chain_id '%s' not found. Chains %v"}`, chainId, a.vals[chainId])))
		return
	}

	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%RPC%", fmt.Sprintf("tcp://%s:26657", a.vals[chainId].HostName()))
	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%CHAIN_ID%", ah.ChainId)
	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%HOME%", a.vals[chainId].HomeDir())

	cmd := strings.Split(ah.Cmd, " ")

	// Output can only ever be 1 thing. So we check which is set, then se the output to the user.
	var output string
	var stdout, stderr []byte

	switch action {
	case "q", "query":
		stdout, stderr, err = (a.vals[chainId]).ExecQuery(a.ctx, cmd...)
	case "b", "bin", "binary":
		stdout, stderr, err = (a.vals[chainId]).ExecBin(a.ctx, cmd...)
	case "e", "exec", "execute":
		stdout, stderr, err = (a.vals[chainId]).Exec(a.ctx, cmd, []string{})
	}

	if len(stdout) > 0 {
		output = string(stdout)
	} else if len(stderr) > 0 {
		output = string(stderr)
	} else if err == nil {
		output = "{}"
	} else {
		output = fmt.Sprintf(`%s`, err)
	}

	// Send the response
	util.Write(w, []byte(output))
}
