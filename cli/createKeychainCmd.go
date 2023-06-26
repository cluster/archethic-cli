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
			err := validateRequiredFlags(cmd.Flags(), "ssh", "ssh-path", "access-seed", "mnemonic")
			cobra.CheckErr(err)
			accessSeedBytes, err := tuiutils.GetSeedBytes(cmd.Flags(), "ssh", "ssh-path", "access-seed", "mnemonic")
			cobra.CheckErr(err)

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
	createKeychainCmd.Flags().Bool("ssh", false, "Enable SSH key mode")
	createKeychainCmd.Flags().String("ssh-path", GetFirstSshKeyDefaultPath(), "Path to ssh key")
	createKeychainCmd.Flags().Bool("mnemonic", false, "Enable mnemonic words for seed")
	createKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh")
	createKeychainCmd.MarkFlagsMutuallyExclusive("access-seed", "ssh-path")
	createKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "ssh")
	createKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "ssh-path")
	createKeychainCmd.MarkFlagsMutuallyExclusive("mnemonic", "access-seed")
	return createKeychainCmd
}
