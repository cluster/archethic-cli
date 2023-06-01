package main

import (
	"fmt"
	"os"

	"github.com/archethic-foundation/archethic-cli/cli"
	"github.com/archethic-foundation/archethic-cli/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "archethic-cli",
	Short: "Archethic CLI",
}

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Terminal User Interface",
	Run: func(cmd *cobra.Command, args []string) {
		tui.StartTea()
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
	rootCmd.AddCommand(uiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
