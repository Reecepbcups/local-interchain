package router

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	ictypes "github.com/reecepbcups/localinterchain/interchain/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"

	"github.com/reecepbcups/localinterchain/interchain/handlers"
)

func NewRouter(ctx context.Context, config *ictypes.Config, vals map[string]*cosmos.ChainNode, relayer ibc.Relayer, eRep ibc.RelayerExecReporter, installDir string) *mux.Router {
	r := mux.NewRouter()

	infoH := handlers.NewInfo(config, installDir)
	r.HandleFunc("/info", infoH.GetInfo).Methods(http.MethodGet)

	endpointsH := handlers.NewEndpoints()
	r.HandleFunc("/", endpointsH.GetEndpoints).Methods(http.MethodGet)

	// TODO: Does this work with the other being just GET?
	actionsH := handlers.NewActions(ctx, vals, relayer, eRep)
	r.HandleFunc("/", actionsH.PostActions).Methods(http.MethodPost)

	uploaderH := handlers.NewUploader(ctx, vals)
	r.HandleFunc("/upload", uploaderH.PostUpload).Methods(http.MethodPost)

	return r
}
