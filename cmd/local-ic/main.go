package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	MakeFileInstallDirectory string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func GetDirectory() string {
	installDir := os.Getenv("INSTALL_DIR")
	if installDir != "" {
		return installDir
	}

	return MakeFileInstallDirectory
}

var rootCmd = &cobra.Command{
	Use:   "local-ic",
	Short: "Your local IBC interchain of nodes program",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
