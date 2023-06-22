package cli

import (
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

func GetAddServiceToKeychainCmd() *cobra.Command {
	addServiceToKeychainCmd := &cobra.Command{
		Use:   "add-service-to-keychain",
		Short: "Add service to keychain",
		Run: func(cmd *cobra.Command, args []string) {
			serviceName, _ := cmd.Flags().GetString("service-name")
			derivationPath, _ := cmd.Flags().GetString("derivation-path")
			err := validateRequiredFlags(cmd.Flags(), "ssh", "ssh-path", "access-seed", "bip39")
			cobra.CheckErr(err)
			accessSeedBytes, err := tuiutils.GetSeedBytes(cmd.Flags(), "ssh", "ssh-path", "access-seed", "bip39")
			cobra.CheckErr(err)

			// set default derivation path if not set
			if !cmd.Flag("derivation-path").Changed {
				derivationPath = "m/650'/" + serviceName + "/0"
			}

			feedback, err := tuiutils.AddServiceToKeychain(accessSeedBytes, endpoint.String(), serviceName, derivationPath)
			cobra.CheckErr(err)
			fmt.Println(feedback)
		},
	}

	addServiceToKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	addServiceToKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	addServiceToKeychainCmd.Flags().String("service-name", "", "Service Name")
	addServiceToKeychainCmd.Flags().String("derivation-path", "", "Derivation Path")
	addServiceToKeychainCmd.Flags().Bool("ssh", false, "Enable SSH key mode")
	addServiceToKeychainCmd.Flags().String("ssh-path", GetFirstSshKeyDefaultPath(), "Path to ssh key")
	addServiceToKeychainCmd.Flags().Bool("bip39", false, "Enable BIP39 words for seed")
	addServiceToKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh")
	addServiceToKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh-path")
	addServiceToKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "ssh")
	addServiceToKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "ssh-path")
	addServiceToKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "access-seed")
	return addServiceToKeychainCmd
}
