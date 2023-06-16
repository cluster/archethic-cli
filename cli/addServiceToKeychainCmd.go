package cli

import (
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
)

func GetAddServiceToKeychainCmd() *cobra.Command {
	addServiceToKeychainCmd := &cobra.Command{
		Use:   "add-service-to-keychain",
		Short: "Add service to keychain",
		Run: func(cmd *cobra.Command, args []string) {
			accessSeed, _ := cmd.Flags().GetString("access-seed")
			serviceName, _ := cmd.Flags().GetString("service-name")
			derivationPath, _ := cmd.Flags().GetString("derivation-path")

			// set default derivation path if not set
			if !cmd.Flag("derivation-path").Changed {
				derivationPath = "m/650'/" + serviceName + "/0"
			}

			accessSeedBytes, err := archethic.MaybeConvertToHex(accessSeed)
			cobra.CheckErr(err)
			feedback, err := tuiutils.AddServiceToKeychain(accessSeedBytes, endpoint.String(), serviceName, derivationPath)
			cobra.CheckErr(err)
			fmt.Println(feedback)
		},
	}

	addServiceToKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	addServiceToKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	addServiceToKeychainCmd.Flags().String("service-name", "", "Service Name")
	addServiceToKeychainCmd.Flags().String("derivation-path", "", "Derivation Path")
	return addServiceToKeychainCmd
}
