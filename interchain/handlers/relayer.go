package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/reecepbcups/localinterchain/interchain/util"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

type relaying struct {
	ctx     context.Context
	vals    map[string]*cosmos.ChainNode
	relayer *ibc.Relayer
	eRep    ibc.RelayerExecReporter
}

func NewRelayerActions(ctx context.Context, vals map[string]*cosmos.ChainNode, relayer *ibc.Relayer, eRep ibc.RelayerExecReporter) *relaying {
	return &relaying{
		ctx:     ctx,
		vals:    vals,
		relayer: relayer,
		eRep:    eRep,
	}
}

type RelayerHandler struct {
	ChainId string `json:"chain-id"`
	Action  string `json:"action"`
	Cmd     string `json:"cmd"`
}

// TODO: Combine with actions.go? ActionHandler & RelayerHandler are the same now.
func (a *relaying) PostRelayerActions(w http.ResponseWriter, r *http.Request) {
	var err error
	var rh RelayerHandler

	err = json.NewDecoder(r.Body).Decode(&rh)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	chainId := rh.ChainId
	action := rh.Action
	if _, ok := a.vals[chainId]; !ok {
		util.Write(w, []byte(fmt.Sprintf(`{"error":"chain-id %s not found"}`, chainId)))
		return
	}

	rh.Cmd = strings.ReplaceAll(rh.Cmd, "%RPC%", fmt.Sprintf("tcp://%s:26657", a.vals[chainId].HostName()))
	rh.Cmd = strings.ReplaceAll(rh.Cmd, "%CHAIN_ID%", rh.ChainId)
	rh.Cmd = strings.ReplaceAll(rh.Cmd, "%HOME%", a.vals[chainId].HomeDir())

	cmd := strings.Split(rh.Cmd, " ")

	// Output can only ever be 1 thing. So we check which is set, then set the output to the user.
	var output string
	var stdout, stderr []byte

	if a.relayer == nil {
		util.Write(w, []byte(`{"error":"relayer not configured for this setup"}`))
		return
	}

	switch action {
	case "e", "exec", "execute":
		if !strings.Contains(rh.Cmd, "--home") {
			// does this ever change for any other relayer?
			cmd = append(cmd, "--home", "/home/relayer")
		}

		res := (*a.relayer).Exec(a.ctx, a.eRep, cmd, []string{})
		stdout = []byte(res.Stdout)
		stderr = []byte(res.Stderr)
		err = res.Err

	case "get_channels", "get-channels", "getChannels":
		res, err := (*a.relayer).GetChannels(a.ctx, a.eRep, chainId)
		if err != nil {
			util.WriteError(w, err)
			return
		}

		r, err := json.Marshal(res)
		if err != nil {
			util.WriteError(w, err)
			return
		}
		stdout = r

	case "stop", "stop-relayer", "stopRelayer", "stop_relayer":
		err = (*a.relayer).StopRelayer(a.ctx, a.eRep)

	case "start", "start-relayer", "startRelayer", "start_relayer":
		paths := strings.FieldsFunc(rh.Cmd, func(c rune) bool {
			return c == ',' || c == ' '
		})
		err = (*a.relayer).StartRelayer(a.ctx, a.eRep, paths...)
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

	util.Write(w, []byte(output))
}
