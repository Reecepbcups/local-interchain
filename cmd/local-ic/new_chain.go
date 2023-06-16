package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	interchain "github.com/reecepbcups/localinterchain/interchain"
	"github.com/spf13/cobra"
)

var reader = bufio.NewReader(os.Stdin)

var newChainCmd = &cobra.Command{
	Use:     "new-chain <name>",
	Aliases: []string{"new", "new-config"},
	Short:   "Create a new chain config",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := strings.CutSuffix(args[0], ".json")
		filePath := path.Join(GetDirectory(), "chains", fmt.Sprintf("%s.json", name))

		// while loop to allow for IBC conncetions to work as expected. Else set IBC as []string{}

		// text, _ := os.ReadFile(filePath)
		// if len(text) > 0 {
		// 	value := getOrDefault(fmt.Sprintf("File %s already exist at this location, override?", name), "false").(string)
		// 	if res, _ := strconv.ParseBool(value); res == false {
		// 		panic(fmt.Sprintf("File %s already exist", filePath))
		// 	}
		// }

		var config interchain.Config
		var chains []interchain.Chain

		for i := 1; i < 1000; i++ {
			fmt.Printf("\n===== Creating new chain #%d =====\n", i)

			c := interchain.Chain{
				// Required
				Name:          getOrDefault("Name", "juno").(string),
				ChainID:       getOrDefault("Chain-ID", "local-1").(string),
				Binary:        getOrDefault("App Binary", "junod").(string),
				Bech32Prefix:  getOrDefault("Bech32 Prefix", "juno").(string),
				GasPrices:     getOrDefault("Gas-Prices (comma seperated)", "0%DENOM%,0other").(string),
				GasAdjustment: getOrDefaultFloat("Gas-Adjustment", 2.5),

				// IBCPaths should be unique chain ids?
				IBCPaths: strings.Split(getOrDefault("IBC Paths", "").(string), ","),
				DockerImage: interchain.DockerImage{
					Repository: getOrDefault("Docker Repo", "ghcr.io/cosmoscontracts/juno-e2e").(string),
					Version:    getOrDefault("Docker Tag/Branch Version", "v15.0.0").(string),
					UidGid:     "1000:1000",
				},

				// maybe if genesis fails, we just recomend to toggle this?
				// SDK v47+ only
				// UseNewGenesisCommand: getOrDefault("Use new genesis command?", false).(bool),
				UseNewGenesisCommand: true,

				// genesis accounts (juno1...:100ujuno,10uatom;)

				// Spam through enter typically
				Denom:          getOrDefault("Denom", "ujuno").(string),
				TrustingPeriod: getOrDefault("Trusting Period", "112h").(string),
				ChainType:      getOrDefault("Chain Type", "cosmos").(string),
				CoinType:       getOrDefault("Coin Type", 118).(int),

				// defaults
				Debugging:  false,
				NumberVals: 1,
				NumberNode: 0,
				BlocksTTL:  -1,
			}
			chains = append(chains, c)

			// break here
			res, err := strconv.ParseBool(getOrDefault("\n\n\n === Add more chains? ===", "false").(string))
			if err != nil || res == false {
				break
			}
		}

		// save c to file in GetD

		config.Chains = chains

		bz, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}

		os.WriteFile(filePath, bz, 777)
	},
}

// use generics to return the type of what defaultValue is?
func getOrDefault(output string, defaultVal any) any {
	fmt.Printf("%s. (Default %v)\n>>> ", output, defaultVal)
	text, err := reader.ReadString('\n')

	if err != nil || text == "\n" {
		// fmt.Printf("Set: %v\n", defaultVal)
		return defaultVal
	}

	return text
}

func getOrDefaultFloat(output string, defaultVal float64) float64 {
	fmt.Printf("%s. (Default %v)\n>>> ", output, defaultVal)
	text, err := reader.ReadString('\n')

	if err != nil || text == "\n" {
		return defaultVal
	}

	text = strings.ReplaceAll(text, "\n", "")
	res, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(err)
	}
	return res
}

//

func init() {
	rootCmd.AddCommand(newChainCmd)
}
