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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetFiles(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		chainsDir := path.Join(GetDirectory(), "chains")

		if len(args) == 0 {
			files := GetFiles()
			for _, file := range files {
				fmt.Printf("- %s\n", file)
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

func GetFiles() []string {
	chainsDir := path.Join(GetDirectory(), "chains")

	files, err := os.ReadDir(chainsDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames
}

func init() {
	rootCmd.AddCommand(chainsCmd)
}
