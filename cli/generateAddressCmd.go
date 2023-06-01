package cli

import (
	"encoding/hex"
	"fmt"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
)

func GetGenerateAddressCmd() *cobra.Command {
	generateAddressCmd := &cobra.Command{
		Use:   "generate-address",
		Short: "Generate address",
		Run: func(cmd *cobra.Command, args []string) {
			seed, _ := cmd.Flags().GetString("seed")
			index, _ := cmd.Flags().GetInt("index")

			fmt.Println("Generating address...")
			curve, err := ellipticCurve.GetCurve()
			if err != nil {
				fmt.Println(err)
				return
			}
			hashAlgo, err := hashAlgo.GetHashAlgo()
			if err != nil {
				fmt.Println(err)
				return
			}
			seedBytes, err := archethic.MaybeConvertToHex(seed)
			if err != nil {
				fmt.Println(err)
				return
			}
			address, err := archethic.DeriveAddress(seedBytes, uint32(index), curve, hashAlgo)
			if err != nil {
				fmt.Println(err)
				return
			} else {
				fmt.Println("Address:", hex.EncodeToString(address))
			}
		},
	}

	generateAddressCmd.Flags().String("seed", "", "Seed")
	generateAddressCmd.Flags().Int("index", 0, "Index")
	generateAddressCmd.Flags().Var(&hashAlgo, "hash-algorithm", "Hash Algorithm (SHA256|SHA512|SHA3_256|SHA3_512|BLAKE2B)")
	generateAddressCmd.Flags().Var(&ellipticCurve, "elliptic-curve", "Elliptic Curve (ED25519|P256|SECP256K1)")
	return generateAddressCmd
}
