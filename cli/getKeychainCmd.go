package cli

import (
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

func GetKeychainCmd() *cobra.Command {
	getKeychainCmd := &cobra.Command{
		Use:   "get-keychain",
		Short: "Get keychain",
		Run: func(cmd *cobra.Command, args []string) {
			accessSeed, _ := cmd.Flags().GetString("access-seed")
			keychain, err := tuiutils.AccessKeychain(endpoint.String(), accessSeed)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Services name | Derivation path")
			for k, service := range keychain.Services {
				fmt.Printf("%s | %s\n", k, service.DerivationPath)
			}
		},
	}
	getKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	getKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	return getKeychainCmd
}
