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
		if cmd.Flags().Changed("ssh") {
			privateKeyPath, _ := cmd.Flags().GetString("ssh")
			privateKey = tuiutils.GetSSHPrivateKey(privateKeyPath)
		}
		tui.StartTea(privateKey)
	},
}

func main() {
	generateAddressCmd := cli.GetGenerateAddressCmd()
	sendTransactionCmd := cli.GetSendTransactionCmd()
	createKeychainCmd := cli.GetCreateKeychainCmd()
	getKeychainCmd := cli.GetKeychainCmd()
	addServiceToKeychainCmd := cli.GetAddServiceToKeychainCmd()
	deleteServiceFromKeychainCmd := cli.GetDeleteServiceFromKeychainCmd()

	rootCmd.AddCommand(generateAddressCmd)
	rootCmd.AddCommand(sendTransactionCmd)
	rootCmd.AddCommand(createKeychainCmd)
	rootCmd.AddCommand(getKeychainCmd)
	rootCmd.AddCommand(addServiceToKeychainCmd)
	rootCmd.AddCommand(deleteServiceFromKeychainCmd)

	rootCmd.Flags().String("ssh", "", "Path to ssh key")

	err := rootCmd.Execute()
	cobra.CheckErr(err)

}
