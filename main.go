package main

import (
	"github.com/archethic-foundation/archethic-cli/cli"
	"github.com/archethic-foundation/archethic-cli/tui"
	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "archethic-cli",
	Short: "Archethic CLI",
	Run: func(cmd *cobra.Command, args []string) {
		var privateKey []byte
		ssh, _ := cmd.Flags().GetBool("ssh")
		isSshPathSet := cmd.Flag("ssh-path").Changed
		isSshEnabled := ssh || isSshPathSet
		if isSshEnabled {
			var err error
			privateKey, err = tuiutils.GetSeedBytes(cmd.Flags(), "ssh", "ssh-path", "")
			cobra.CheckErr(err)
		}
		tui.StartTea(privateKey)
	},
}

func main() {
	generateAddressCmd := cli.GetGenerateAddressCmd()
	sendTransactionCmd := cli.GetSendTransactionCmd()
	getTransactionFeeCmd := cli.GetGetTransactionFeeCmd()
	createKeychainCmd := cli.GetCreateKeychainCmd()
	getKeychainCmd := cli.GetKeychainCmd()
	addServiceToKeychainCmd := cli.GetAddServiceToKeychainCmd()
	deleteServiceFromKeychainCmd := cli.GetDeleteServiceFromKeychainCmd()

	rootCmd.AddCommand(generateAddressCmd)
	rootCmd.AddCommand(sendTransactionCmd)
	rootCmd.AddCommand(getTransactionFeeCmd)
	rootCmd.AddCommand(createKeychainCmd)
	rootCmd.AddCommand(getKeychainCmd)
	rootCmd.AddCommand(addServiceToKeychainCmd)
	rootCmd.AddCommand(deleteServiceFromKeychainCmd)

	rootCmd.Flags().Bool("ssh", false, "Enable SSH key mode")
	rootCmd.Flags().String("ssh-path", cli.GetFirstSshKeyDefaultPath(), "Path to ssh key")

	err := rootCmd.Execute()
	cobra.CheckErr(err)

}
