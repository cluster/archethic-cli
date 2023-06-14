package cli

import (
	"encoding/json"
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
)

func GetKeychainCmd() *cobra.Command {
	getKeychainCmd := &cobra.Command{
		Use:   "get-keychain",
		Short: "Get keychain",
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

			keychain, err := tuiutils.AccessKeychain(endpoint.String(), accessSeedBytes)
			cobra.CheckErr(err)
			jsonServices, err := json.Marshal(keychain.Services)
			cobra.CheckErr(err)
			fmt.Printf("%s\n", jsonServices)
		},
	}
	getKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	getKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	getKeychainCmd.Flags().String("ssh", "", "Path to ssh key")
	return getKeychainCmd
}
