package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var chainsCmd = &cobra.Command{
	Use:   "chains [config.json]",
	Short: "List all current chains or outputs a current config information",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		chainsDir := path.Join(GetDirectory(), "chains")

		if len(args) == 0 {
			allChains, err := os.ReadDir(chainsDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, chain := range allChains {
				fmt.Println(chain.Name())
			}
		} else {
			config := args[0]
			filePath := path.Join(chainsDir, config)

			fc, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(string(fc))
		}
	},
}

func init() {
	rootCmd.AddCommand(chainsCmd)
}
