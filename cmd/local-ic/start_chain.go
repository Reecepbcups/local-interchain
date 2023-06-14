package main

import (
	"github.com/spf13/cobra"

	interchain "github.com/reecepbcups/localinterchain/interchain"
)

var startCmd = &cobra.Command{
	Use:     "start <config.json>",
	Aliases: []string{"s"},
	Short:   "Starts up the chain of choice with the config name",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := args[0]
		interchain.StartChain(GetDirectory(), config)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
