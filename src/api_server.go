package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

// start as `go StartNonBlockingServer()`
func StartNonBlockingServer(ctx context.Context, config *MainConfig, vals map[string]*cosmos.ChainNode) {
	// TODO: Multiple in 1?
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlePostRequest(w, r, ctx, vals)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Only POST requests are allowed"))
		}
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleUploadFile(w, r, ctx, vals)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Only POST requests are allowed"))
		}
	})

	server := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	if err := http.ListenAndServe(server, nil); err != nil {
		log.Default().Println(err)
	}
}

type ActionHandler struct {
	ChainId string `json:"chain-id"`
	Action  string `json:"action"`
	Cmd     string `json:"cmd"`
}

type Uploader struct {
	ChainId  string `json:"chain-id"`
	KeyName  string `json:"key-name"`
	FileName string `json:"file-name"`
}

func handleUploadFile(w http.ResponseWriter, r *http.Request, ctx context.Context, vals map[string]*cosmos.ChainNode) {
	var u Uploader
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error":%s}`, err)))
		return
	}

	log.Printf("Uploader: %+v", u)

	chainId := u.ChainId
	if _, ok := vals[chainId]; !ok {
		w.Write([]byte(fmt.Sprintf(`{"error":"chain-id %s not found"}`, chainId)))
		return
	}

	codeId, err := vals[chainId].StoreContract(ctx, u.KeyName, u.FileName)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error":%s}`, err)))
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"code_id":%s}`, codeId)))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request, ctx context.Context, vals map[string]*cosmos.ChainNode) {
	var ah ActionHandler
	err := json.NewDecoder(r.Body).Decode(&ah)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error":%s}`, err)))
		return
	}

	chainId := ah.ChainId
	action := ah.Action
	if _, ok := vals[chainId]; !ok {
		w.Write([]byte(fmt.Sprintf(`{"error":"chain-id %s not found"}`, chainId)))
		return
	}

	// replace env variables
	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%RPC%", fmt.Sprintf("tcp://%s:26657", vals[chainId].HostName()))
	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%CHAIN_ID%", ah.ChainId)
	ah.Cmd = strings.ReplaceAll(ah.Cmd, "%HOME%", vals[chainId].HomeDir())

	cmd := strings.Split(ah.Cmd, " ")

	// Output can only ever be 1 thing. So we check which is set, then se the output to the user.
	var output string
	var stdout, stderr []byte

	switch action {
	case "q", "query":
		stdout, stderr, err = (vals[chainId]).ExecQuery(ctx, cmd...)
	case "b", "bin", "binary":
		stdout, stderr, err = (vals[chainId]).ExecBin(ctx, cmd...)
	case "e", "exec", "execute":
		stdout, stderr, err = (vals[chainId]).Exec(ctx, cmd, []string{})
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
	w.Write([]byte(output))
}
