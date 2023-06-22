package cli

import (
	"fmt"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

func GetDeleteServiceFromKeychainCmd() *cobra.Command {
	deleteServiceFromKeychainCmd := &cobra.Command{
		Use:   "delete-service-from-keychain",
		Short: "Delete service from keychain",
		Run: func(cmd *cobra.Command, args []string) {
			serviceName, _ := cmd.Flags().GetString("service-name")
			err := validateRequiredFlags(cmd.Flags(), "ssh", "ssh-path", "access-seed", "bip39")
			cobra.CheckErr(err)
			accessSeedBytes, err := tuiutils.GetSeedBytes(cmd.Flags(), "ssh", "ssh-path", "access-seed", "bip39")
			cobra.CheckErr(err)
			feedback, err := tuiutils.RemoveServiceFromKeychain(accessSeedBytes, endpoint.String(), serviceName)
			cobra.CheckErr(err)
			fmt.Println(feedback)
		},
	}
	deleteServiceFromKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	deleteServiceFromKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	deleteServiceFromKeychainCmd.Flags().String("service-name", "", "Service Name")
	deleteServiceFromKeychainCmd.Flags().Bool("ssh", false, "Enable SSH key mode")
	deleteServiceFromKeychainCmd.Flags().String("ssh-path", GetFirstSshKeyDefaultPath(), "Path to ssh key")
	deleteServiceFromKeychainCmd.Flags().Bool("bip39", false, "Enable BIP39 words for seed")
	deleteServiceFromKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh")
	deleteServiceFromKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh-path")
	deleteServiceFromKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "ssh")
	deleteServiceFromKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "ssh-path")
	deleteServiceFromKeychainCmd.MarkFlagsMutuallyExclusive("bip39", "access-seed")
	return deleteServiceFromKeychainCmd
}
