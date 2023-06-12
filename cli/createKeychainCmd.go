package cli

import (
	"encoding/json"
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

func GetCreateKeychainCmd() *cobra.Command {
	createKeychainCmd := &cobra.Command{
		Use:   "create-keychain",
		Short: "Create keychain",
		Run: func(cmd *cobra.Command, args []string) {
			accessSeed, _ := cmd.Flags().GetString("access-seed")
			feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, err := tuiutils.CreateKeychain(endpoint.String(), accessSeed)
			cobra.CheckErr(err)

			data := map[string]interface{}{
				"feedback":                         feedback,
				"keychainSeed":                     keychainSeed,
				"keychainTransactionAddress":       keychainTransactionAddress,
				"keychainAccessTransactionAddress": keychainAccessTransactionAddress,
			}

			jsonData, err := json.Marshal(data)
			cobra.CheckErr(err)

			fmt.Println(string(jsonData))
		},
	}
	createKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	createKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	return createKeychainCmd
}
