package cli

import (
	"encoding/json"
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
)

func GetCreateKeychainCmd() *cobra.Command {
	createKeychainCmd := &cobra.Command{
		Use:   "create-keychain",
		Short: "Create keychain",
		Run: func(cmd *cobra.Command, args []string) {
			accessSeed, _ := cmd.Flags().GetString("access-seed")
			privateKeyPath, _ := cmd.Flags().GetString("ssh")

			var accessSeedBytes []byte
			if privateKeyPath != "" {
				accessSeedBytes = tuiutils.GetSSHPrivateKey(privateKeyPath)
			} else {
				var err error
				accessSeedBytes, err = archethic.MaybeConvertToHex(accessSeed)
				cobra.CheckErr(err)
			}

			feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, err := tuiutils.CreateKeychain(endpoint.String(), accessSeedBytes)
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
	createKeychainCmd.Flags().String("ssh", "", "Path to ssh key")
	return createKeychainCmd
}
