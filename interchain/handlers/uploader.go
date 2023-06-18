package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type upload struct {
	ctx  context.Context
	vals map[string]*cosmos.ChainNode
}

type Uploader struct {
	ChainId  string `json:"chain-id"`
	KeyName  string `json:"key-name"`
	FileName string `json:"file-name"`
}

func NewUploader(ctx context.Context, vals map[string]*cosmos.ChainNode) *upload {
	return &upload{
		ctx:  ctx,
		vals: vals,
	}
}

func (up *upload) PostUpload(w http.ResponseWriter, r *http.Request) {
	var upload Uploader
	err := json.NewDecoder(r.Body).Decode(&up)
	if err != nil {
		WriteError(w, err)
		return
	}

	log.Printf("Uploader: %+v", up)

	chainId := upload.ChainId
	if _, ok := up.vals[chainId]; !ok {
		Write(w, []byte(fmt.Sprintf(`{"error":"chain-id %s not found"}`, chainId)))
		return
	}

	codeId, err := up.vals[chainId].StoreContract(up.ctx, upload.KeyName, upload.FileName)

	if err != nil {
		WriteError(w, err)
		return
	}

	Write(w, []byte(fmt.Sprintf(`{"code_id":%s}`, codeId)))
}
