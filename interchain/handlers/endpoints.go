package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/reecepbcups/localinterchain/interchain/util"
)

type endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

var endpoints = []endpoint{
	{
		Method: "POST",
		Path:   "/",
	},
	{
		Method: "GET",
		Path:   "/info",
	},
	{
		Method: "POST",
		Path:   "/upload",
	},
}

func NewEndpoints() *endpoint {
	return &endpoint{}
}

func (e *endpoint) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	jsonRes, err := json.MarshalIndent(endpoints, "", "  ")
	if err != nil {
		util.WriteError(w, err)
		return
	}

	util.Write(w, jsonRes)
}
