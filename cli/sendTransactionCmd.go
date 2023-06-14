package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func extractTransactionFromInputFile(config string) (ConfiguredTransaction, error) {
	configBytes, err := os.ReadFile(config)
	if err != nil {
		return ConfiguredTransaction{}, err
	}
	var data SendTransactionData
	err = yaml.Unmarshal(configBytes, &data)
	if err != nil {
		return ConfiguredTransaction{}, err
	}
	endpoint.Set(data.Endpoint)
	ellipticCurve.Set(data.EllipticCurve)
	transactionType.Set(data.TransactionType)
	seedByte, err := archethic.MaybeConvertToHex(data.AccessSeed)
	if err != nil {
		return ConfiguredTransaction{}, err
	}
	return ConfiguredTransaction{
		accessSeed:     seedByte,
		index:          data.Index,
		ucoTransfers:   data.UcoTransfers,
		tokenTransfers: data.TokenTransfers,
		recipients:     data.Recipients,
		ownerships:     data.Ownerships,
		content:        []byte(data.Content),
		smartContract:  data.SmartContract,
		serviceName:    data.ServiceName,
	}, nil

}

func extractTransactionFromInputFlags(cmd *cobra.Command) (ConfiguredTransaction, error) {
	accessSeed, _ := cmd.Flags().GetString("access-seed")
	index, _ := cmd.Flags().GetInt("index")
	serviceName, _ := cmd.Flags().GetString("serviceName")
	recipients, _ := cmd.Flags().GetStringSlice("recipients")

	// extract uco transfers
	ucoTransfersStr, _ := cmd.Flags().GetStringToString("uco-transfer")
	var ucoTransfers []UCOTransfer
	for to, amount := range ucoTransfersStr {
		toBytes, err := hex.DecodeString(to)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
		amountInt, err := strconv.ParseUint(amount, 10, 64)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
		ucoTransfers = append(ucoTransfers, UCOTransfer{
			To:     hex.EncodeToString(toBytes),
			Amount: amountInt,
		})
	}

	// extract token transfers
	tokenTransfersStr, _ := cmd.Flags().GetStringToString("token-transfer")
	var tokenTransfers []TokenTransfer
	for to, values := range tokenTransfersStr {
		value := strings.Split(values, ",")
		amountInt, err := strconv.ParseUint(value[0], 10, 64)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
		tokenId, err := strconv.ParseInt(value[2], 10, 64)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
		tokenTransfers = append(tokenTransfers, TokenTransfer{
			To:           to,
			Amount:       amountInt,
			TokenAddress: value[1],
			TokenID:      int(tokenId),
		})
	}

	// extract ownerships
	ownershipsStr, _ := cmd.Flags().GetStringToString("ownerships")
	var ownerships []Ownership
	mapSecretOwnership := mapOwnership(ownershipsStr)
	for secret, authorizedKeys := range mapSecretOwnership {
		ownerships = append(ownerships, Ownership{
			Secret:         secret,
			AuthorizedKeys: authorizedKeys,
		})
	}

	// extract content
	content, _ := cmd.Flags().GetString("content")
	contentBytes := []byte{}
	var err error
	if content != "" {
		contentBytes, err = os.ReadFile(content)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
	}

	// extract smart contract
	smartContract, _ := cmd.Flags().GetString("smart-contract")
	smartContractStr := ""
	if smartContract != "" {
		smartContractBytes, err := os.ReadFile(smartContract)
		if err != nil {
			return ConfiguredTransaction{}, err
		}
		smartContractStr = string(smartContractBytes)
	}

	privateKeyPath, _ := cmd.Flags().GetString("ssh")

	var accessSeedBytes []byte
	if privateKeyPath != "" {
		accessSeedBytes = tuiutils.GetSSHPrivateKey(privateKeyPath)
	} else {
		var err error
		accessSeedBytes, err = archethic.MaybeConvertToHex(accessSeed)
		cobra.CheckErr(err)
	}

	return ConfiguredTransaction{
		accessSeed:     accessSeedBytes,
		index:          index,
		ucoTransfers:   ucoTransfers,
		tokenTransfers: tokenTransfers,
		recipients:     recipients,
		ownerships:     ownerships,
		content:        contentBytes,
		smartContract:  smartContractStr,
		serviceName:    serviceName,
	}, nil
}

func configureTransaction(configuredTransaction ConfiguredTransaction, txType archethic.TransactionType, secretKey []byte) (*archethic.TransactionBuilder, error) {

	transaction := archethic.NewTransaction(txType)

	// set uco transfers
	for _, ucoTransfer := range configuredTransaction.ucoTransfers {
		toBytes, err := hex.DecodeString(ucoTransfer.To)
		if err != nil {
			return nil, err
		}
		transaction.AddUcoTransfer(toBytes, ucoTransfer.Amount*1e8)
	}

	// set token transfers
	for _, tokenTransfer := range configuredTransaction.tokenTransfers {
		toBytes, err := hex.DecodeString(tokenTransfer.To)
		if err != nil {
			return nil, err
		}

		tokenAddress, err := hex.DecodeString(tokenTransfer.TokenAddress)
		if err != nil {
			return nil, err
		}
		transaction.AddTokenTransfer(toBytes, tokenAddress, tokenTransfer.Amount*1e8, tokenTransfer.TokenID)
	}

	// set recipients
	for _, recipient := range configuredTransaction.recipients {
		recipientBytes, err := hex.DecodeString(recipient)
		if err != nil {
			return nil, err
		}
		transaction.AddRecipient(recipientBytes)
	}

	// set ownerships
	for _, ownership := range configuredTransaction.ownerships {
		cipher, err := archethic.AesEncrypt([]byte(ownership.Secret), secretKey)
		if err != nil {
			return nil, err
		}
		authorizedKeysResult := make([]archethic.AuthorizedKey, len(ownership.AuthorizedKeys))
		for i, key := range ownership.AuthorizedKeys {
			keyByte, err := hex.DecodeString(key)
			if err != nil {
				return nil, err
			}
			encrypedSecretKey, err := archethic.EcEncrypt(secretKey, keyByte)
			if err != nil {
				return nil, err
			}
			authorizedKeysResult[i] = archethic.AuthorizedKey{
				PublicKey:          keyByte,
				EncryptedSecretKey: encrypedSecretKey,
			}
		}
		transaction.AddOwnership(cipher, authorizedKeysResult)
	}

	transaction.SetContent(configuredTransaction.content)
	transaction.SetCode(configuredTransaction.smartContract)

	return transaction, nil
}

func mapOwnership(ownerships map[string]string) map[string][]string {
	result := make(map[string][]string)

	for secret, authorizedKey := range ownerships {
		if _, ok := result[secret]; !ok {
			result[secret] = []string{authorizedKey}
		} else {
			result[secret] = append(result[secret], authorizedKey)
		}
	}

	return result
}

func GetSendTransactionCmd() *cobra.Command {
	sendTransactionCmd := &cobra.Command{
		Use:   "send-transaction",
		Short: "Send transaction",
		Run: func(cmd *cobra.Command, args []string) {
			secretKey := make([]byte, 32)
			rand.Read(secretKey)

			config, _ := cmd.Flags().GetString("config")
			var configuredTransaction ConfiguredTransaction
			var err error
			if config != "" {
				configuredTransaction, err = extractTransactionFromInputFile(config)
				cobra.CheckErr(err)
			} else {
				configuredTransaction, err = extractTransactionFromInputFlags(cmd)
				cobra.CheckErr(err)
			}

			curve, err := ellipticCurve.GetCurve()
			cobra.CheckErr(err)

			txType, err := transactionType.GetTransactionType()
			cobra.CheckErr(err)

			transaction, err := configureTransaction(configuredTransaction, txType, secretKey)
			cobra.CheckErr(err)

			serviceMode := configuredTransaction.serviceName != ""

			client := archethic.NewAPIClient(endpoint.String())

			// if no index is provided and not in serviceMode, get the last transaction index
			if !cmd.Flags().Changed("index") && !serviceMode {
				configuredTransaction.index, err = tuiutils.GetLastTransactionIndex(endpoint.String(), curve, configuredTransaction.accessSeed)
				cobra.CheckErr(err)
			}

			storageNouncePublicKey, err := client.GetStorageNoncePublicKey()
			cobra.CheckErr(err)

			feedback, err := tuiutils.SendTransaction(*transaction, secretKey, curve, serviceMode, endpoint.String(), configuredTransaction.index, configuredTransaction.serviceName, storageNouncePublicKey, configuredTransaction.accessSeed)
			cobra.CheckErr(err)
			fmt.Println(feedback)
		},
	}
	sendTransactionCmd.Flags().String("config", "", "The file location of the YAML configuration file")
	sendTransactionCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	sendTransactionCmd.Flags().String("access-seed", "", "Access Seed")
	sendTransactionCmd.Flags().String("ssh", "", "Path to ssh key")
	sendTransactionCmd.Flags().Int("index", 0, "Index")
	sendTransactionCmd.Flags().Var(&ellipticCurve, "elliptic-curve", "Elliptic Curve (ED25519|P256|SECP256K1)")
	sendTransactionCmd.Flags().Var(&transactionType, "transaction-type", "Transaction Type (keychain_access|keychain|transfer|hosting|token|data|contract|code_proposal|code_approval)")
	sendTransactionCmd.Flags().StringToString("uco-transfer", map[string]string{}, "UCO Transfers (format: to=amount)")
	sendTransactionCmd.Flags().StringToString("token-transfer", map[string]string{}, "Token Transfers (format: to=amount,token_address,token_id)")
	sendTransactionCmd.Flags().StringSlice("recipients", []string{}, "Recipients")
	sendTransactionCmd.Flags().StringToString("ownerships", map[string]string{}, "Ownerships (format: secret=authorization_key)")
	sendTransactionCmd.Flags().String("content", "", "The file location of the content")
	sendTransactionCmd.Flags().String("smart-contract", "", "The file location containing the smart Contract")
	sendTransactionCmd.Flags().String("serviceName", "", "Service Name (required if creating a transaction for a service)")
	return sendTransactionCmd
}
