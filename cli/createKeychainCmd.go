package cli

import (
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

			fmt.Println("Creating keychain...")
			fmt.Println("Endpoint:", endpoint)
			fmt.Println("Access Seed:", accessSeed)
			feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, error := tuiutils.CreateKeychain(endpoint.String(), accessSeed)
			if error != nil {
				fmt.Println(error)
			} else {
				fmt.Println("Feedback:", feedback)
				fmt.Println("Keychain Seed:", keychainSeed)
				fmt.Println("Keychain Transaction Address:", keychainTransactionAddress)
				fmt.Println("Keychain Access Transaction Address:", keychainAccessTransactionAddress)
			}
		},
	}
	createKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	createKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	return createKeychainCmd
}
