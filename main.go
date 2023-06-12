package main

import (
	"github.com/archethic-foundation/archethic-cli/cli"
	"github.com/archethic-foundation/archethic-cli/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "archethic-cli",
	Short: "Archethic CLI",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if any flags are passed
		flagsPassed := false
		cmd.Flags().Visit(func(f *pflag.Flag) {
			flagsPassed = true
		})

		// If no flags are passed, start the TUI
		if !flagsPassed {
			tui.StartTea()
		}
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

	err := rootCmd.Execute()
	cobra.CheckErr(err)

}
