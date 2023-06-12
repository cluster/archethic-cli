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

			curve, err := ellipticCurve.GetCurve()
			cobra.CheckErr(err)
			hashAlgo, err := hashAlgo.GetHashAlgo()
			cobra.CheckErr(err)
			seedBytes, err := archethic.MaybeConvertToHex(seed)
			cobra.CheckErr(err)
			address, err := archethic.DeriveAddress(seedBytes, uint32(index), curve, hashAlgo)
			cobra.CheckErr(err)
			fmt.Println(hex.EncodeToString(address))
		},
	}

	generateAddressCmd.Flags().String("seed", "", "Seed")
	generateAddressCmd.Flags().Int("index", 0, "Index")
	generateAddressCmd.Flags().Var(&hashAlgo, "hash-algorithm", "Hash Algorithm (SHA256|SHA512|SHA3_256|SHA3_512|BLAKE2B)")
	generateAddressCmd.Flags().Var(&ellipticCurve, "elliptic-curve", "Elliptic Curve (ED25519|P256|SECP256K1)")
	return generateAddressCmd
}
