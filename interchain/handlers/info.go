package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	types "github.com/reecepbcups/localinterchain/interchain/types"
)

type info struct {
	// ctx    context.Context
	Config     *types.Config
	InstallDir string
}

func NewInfo(cfg *types.Config, installDir string) *info {
	return &info{
		Config:     cfg,
		InstallDir: installDir,
	}
}

type GetInfo struct {
	Logs   types.MainLogs `json:"logs"`
	Chains []types.Chain  `json:"chains"`
	Relay  types.Relayer  `json:"relayer"`
}

func (i *info) GetInfo(w http.ResponseWriter, r *http.Request) {
	fp := filepath.Join(i.InstallDir, "configs", "logs.json")

	bz, err := os.ReadFile(fp)
	if err != nil {
		WriteError(w, err)
		return
	}

	var logs types.MainLogs
	if err := json.Unmarshal(bz, &logs); err != nil {
		WriteError(w, err)
		return
	}

	info := GetInfo{
		Logs:   logs,
		Chains: i.Config.Chains,
		Relay:  i.Config.Relayer,
	}

	jsonRes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		WriteError(w, err)
		return
	}

	Write(w, jsonRes)
}
