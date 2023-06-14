package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	MakeFileInstallDirectory string
)

var rootCmd = &cobra.Command{
	Use:   "local-ic",
	Short: "Your local IBC interchain of nodes program",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func GetDirectory() string {
	installDir := os.Getenv("INSTALL_DIR")
	if installDir != "" {
		return installDir
	}

	return MakeFileInstallDirectory
}
