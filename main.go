package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui"
	"github.com/archethic-foundation/archethic-cli/tui/tuiutils"
	archethic "github.com/archethic-foundation/libgo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

var (
	hashAlgo        = SHA256
	ellipticCurve   = ED25519
	endpoint        = local
	transactionType = TransferType
)

var generateAddressCmd = &cobra.Command{
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

type SendTransactionData struct {
	Endpoint        string            `yaml:"endpoint"`
	AccessSeed      string            `yaml:"access_seed"`
	Index           int               `yaml:"index"`
	EllipticCurve   string            `yaml:"elliptic_curve"`
	TransactionType string            `yaml:"transaction_type"`
	UcoTransfers    map[string]string `yaml:"uco_transfers,omitempty"`
	TokenTransfers  map[string]string `yaml:"token_transfers,omitempty"`
	Recipients      []string          `yaml:"recipients,omitempty"`
	Ownerships      map[string]string `yaml:"ownerships,omitempty"`
	Content         string            `yaml:"content,omitempty"`
	SmartContract   string            `yaml:"smart_contract,omitempty"`
	ServiceName     string            `yaml:"serviceName,omitempty"`
}

type UCOTransfer struct {
	To     string `yaml:"to"`
	Amount int    `yaml:"amount"`
}

type TokenTransfer struct {
	To           string `yaml:"to"`
	Amount       int    `yaml:"amount"`
	TokenAddress string `yaml:"token_address"`
	TokenID      int    `yaml:"token_id"`
}

type Ownership struct {
	Secret            string   `yaml:"secret"`
	AuthorizationKeys []string `yaml:"authorization_keys"`
}

var sendTransactionCmd = &cobra.Command{
	Use:   "send-transaction",
	Short: "Send transaction",
	Run: func(cmd *cobra.Command, args []string) {
		var accessSeed string
		var index int
		var ucoTransfers map[string]string
		var tokenTransfers map[string]string

		var recipients []string
		var ownerships map[string]string
		var content string
		var smartContract string
		var serviceName string

		config, _ := cmd.Flags().GetString("config")
		if config != "" {
			configBytes, err := os.ReadFile(config)
			if err != nil {
				fmt.Println(err)
				return
			}
			var data SendTransactionData
			err = yaml.Unmarshal(configBytes, &data)
			if err != nil {
				fmt.Println(err)
				return
			}
			accessSeed = data.AccessSeed
			index = data.Index
			ucoTransfers = data.UcoTransfers
			tokenTransfers = data.TokenTransfers
			recipients = data.Recipients
			ownerships = data.Ownerships
			content = data.Content
			smartContract = data.SmartContract
			serviceName = data.ServiceName
			endpoint.Set(data.Endpoint)
			ellipticCurve.Set(data.EllipticCurve)
			transactionType.Set(data.TransactionType)

		} else {
			accessSeed, _ = cmd.Flags().GetString("access-seed")
			index, _ = cmd.Flags().GetInt("index")
			ucoTransfers, _ = cmd.Flags().GetStringToString("uco-transfers")
			tokenTransfers, _ = cmd.Flags().GetStringToString("token-transfers")

			recipients, _ = cmd.Flags().GetStringSlice("recipients")
			ownerships, _ = cmd.Flags().GetStringToString("ownerships")
			content, _ = cmd.Flags().GetString("content")
			smartContract, _ = cmd.Flags().GetString("smart-contract")
			serviceName, _ = cmd.Flags().GetString("serviceName")
		}

		secretKey := make([]byte, 32)
		rand.Read(secretKey)

		serviceMode := serviceName != ""

		client := archethic.NewAPIClient(endpoint.String())
		storageNouncePublicKey, err := client.GetStorageNoncePublicKey()
		if err != nil {
			fmt.Println(err)
			return
		}

		txType, err := transactionType.GetTransactionType()
		if err != nil {
			fmt.Println(err)
			return
		}
		transaction := archethic.NewTransaction(txType)

		for to, amount := range ucoTransfers {
			toBytes, err := hex.DecodeString(to)
			if err != nil {
				fmt.Println(err)
				return
			}
			amountInt, err := strconv.ParseUint(amount, 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			transaction.AddUcoTransfer(toBytes, amountInt)
		}

		for to, values := range tokenTransfers {
			toBytes, err := hex.DecodeString(to)
			if err != nil {
				fmt.Println(err)
				return
			}
			value := strings.Split(values, ",")
			amountInt, err := strconv.ParseUint(value[0], 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			tokenAddress, err := hex.DecodeString(value[1])

			if err != nil {
				fmt.Println(err)
				return
			}
			tokenId, err := strconv.ParseInt(value[2], 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			transaction.AddTokenTransfer(toBytes, tokenAddress, amountInt, int(tokenId))
		}

		for _, recipient := range recipients {
			recipientBytes, err := hex.DecodeString(recipient)
			if err != nil {
				fmt.Println(err)
				return
			}
			transaction.AddRecipient(recipientBytes)
		}

		mapSecretOwnership := mapOwnership(ownerships)
		for secret, authorizedKeys := range mapSecretOwnership {
			cipher, err := archethic.AesEncrypt([]byte(secret), secretKey)
			if err != nil {
				fmt.Println(err)
				return
			}
			authorizedKeysResult := make([]archethic.AuthorizedKey, len(authorizedKeys))
			for i, key := range authorizedKeys {
				keyByte, err := hex.DecodeString(key)
				if err != nil {
					fmt.Println(err)
					return
				}
				encrypedSecretKey, err := archethic.EcEncrypt(secretKey, keyByte)
				if err != nil {
					fmt.Println(err)
					return
				}
				authorizedKeysResult[i] = archethic.AuthorizedKey{
					PublicKey:          keyByte,
					EncryptedSecretKey: encrypedSecretKey,
				}
			}
			transaction.AddOwnership(cipher, authorizedKeysResult)
		}

		if content != "" {
			contentBytes, err := os.ReadFile(content)
			if err != nil {
				fmt.Println(err)
				return
			}

			transaction.SetContent(contentBytes)
		}

		if smartContract != "" {
			smartContractBytes, err := os.ReadFile(smartContract)
			if err != nil {
				fmt.Println(err)
				return
			}
			transaction.SetCode(string(smartContractBytes))
		}

		curve, err := ellipticCurve.GetCurve()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Sending transaction...")
		tuiutils.SendTransaction(*transaction, secretKey, curve, serviceMode, endpoint.String(), index, serviceName, storageNouncePublicKey, accessSeed)
	},
}

var createKeychainCmd = &cobra.Command{
	Use:   "create-keychain",
	Short: "Create keychain",
	Run: func(cmd *cobra.Command, args []string) {
		accessSeed, _ := cmd.Flags().GetString("access-seed")

		fmt.Println("Creating keychain...")
		fmt.Println("Endpoint:", endpoint)
		fmt.Println("Access Seed:", accessSeed)
		feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, error := tuiutils.CreateKeychain(endpoint.String(), accessSeed)
		if error != nil {
			fmt.Println(error)
		} else {
			fmt.Println("Feedback:", feedback)
			fmt.Println("Keychain Seed:", keychainSeed)
			fmt.Println("Keychain Transaction Address:", keychainTransactionAddress)
			fmt.Println("Keychain Access Transaction Address:", keychainAccessTransactionAddress)
		}
	},
}

var getKeychainCmd = &cobra.Command{
	Use:   "get-keychain",
	Short: "Get keychain",
	Run: func(cmd *cobra.Command, args []string) {
		accessSeed, _ := cmd.Flags().GetString("access-seed")
		keychain, err := tuiutils.AccessKeychain(endpoint.String(), accessSeed)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Services name | Derivation path")
		for k, service := range keychain.Services {
			fmt.Printf("%s | %s\n", k, service.DerivationPath)
		}
	},
}

var addServiceToKeychainCmd = &cobra.Command{
	Use:   "add-service-to-keychain",
	Short: "Add service to keychain",
	Run: func(cmd *cobra.Command, args []string) {
		accessSeed, _ := cmd.Flags().GetString("access-seed")
		serviceName, _ := cmd.Flags().GetString("service-name")
		derivationPath, _ := cmd.Flags().GetString("derivation-path")

		fmt.Println("Adding service to keychain...")
		accessSeedBytes, err := archethic.MaybeConvertToHex(accessSeed)
		if err != nil {
			fmt.Println(err)
			return
		}
		feedback, err := tuiutils.AddServiceToKeychain(accessSeedBytes, endpoint.String(), serviceName, derivationPath)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(feedback)
		}
	},
}

var deleteServiceFromKeychainCmd = &cobra.Command{
	Use:   "delete-service-from-keychain",
	Short: "Delete service from keychain",
	Run: func(cmd *cobra.Command, args []string) {
		accessSeed, _ := cmd.Flags().GetString("access-seed")
		serviceName, _ := cmd.Flags().GetString("service-name")

		fmt.Println("Deleting service from keychain...")
		accessSeedBytes, err := archethic.MaybeConvertToHex(accessSeed)
		if err != nil {
			fmt.Println(err)
			return
		}
		feedback, err := tuiutils.RemoveServiceFromKeychain(accessSeedBytes, endpoint.String(), serviceName)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(feedback)
		}
	},
}

type EndpointCLI string

const (
	local   EndpointCLI = "local"
	testnet EndpointCLI = "testnet"
	mainnet EndpointCLI = "mainnet"
)

func (e *EndpointCLI) String() string {
	switch *e {
	case local:
		return "http://localhost:4000"
	case testnet:
		return "https://testnet.archethic.net"
	case mainnet:
		return "https://mainnet.archethic.net"
	default:
		_, err := url.Parse(string(*e))
		if err != nil {
			return ""
		}
		return string(*e)
	}
}

func (e *EndpointCLI) Set(value string) error {
	switch value {
	case "local":
		*e = "local"
	case "testnet":
		*e = "testnet"
	case "mainnet":
		*e = "mainnet"
	default:
		_, err := url.Parse(string(*e))
		if err != nil {
			return errors.New("invalid endpoint value")
		}
		*e = EndpointCLI(value)
	}
	return nil
}

func (e *EndpointCLI) Type() string {
	return "EndpointCLI"
}

type HashAlgoCLI uint8

const (
	SHA256   HashAlgoCLI = 0
	SHA512   HashAlgoCLI = 1
	SHA3_256 HashAlgoCLI = 2
	SHA3_512 HashAlgoCLI = 3
	BLAKE2B  HashAlgoCLI = 4
)

func (ha *HashAlgoCLI) String() string {
	switch *ha {
	case SHA256:
		return "SHA256"
	case SHA512:
		return "SHA512"
	case SHA3_256:
		return "SHA3_256"
	case SHA3_512:
		return "SHA3_512"
	case BLAKE2B:
		return "BLAKE2B"
	default:
		return ""
	}
}

func (ha *HashAlgoCLI) Set(value string) error {
	switch value {
	case "SHA256":
		*ha = SHA256
	case "SHA512":
		*ha = SHA512
	case "SHA3_256":
		*ha = SHA3_256
	case "SHA3_512":
		*ha = SHA3_512
	case "BLAKE2B":
		*ha = BLAKE2B
	default:
		return errors.New("invalid HashAlgo value")
	}
	return nil
}

func (ha *HashAlgoCLI) GetHashAlgo() (archethic.HashAlgo, error) {
	switch *ha {
	case SHA256:
		return archethic.SHA256, nil
	case SHA512:
		return archethic.SHA512, nil
	case SHA3_256:
		return archethic.SHA3_256, nil
	case SHA3_512:
		return archethic.SHA3_512, nil
	case BLAKE2B:
		return archethic.BLAKE2B, nil
	default:
		return archethic.SHA256, errors.New("invalid HashAlgo value")
	}
}

func (ha *HashAlgoCLI) Type() string {
	return "HashAlgoCLI"
}

type CurveCLI uint8

const (
	ED25519   CurveCLI = 0
	P256      CurveCLI = 1
	SECP256K1 CurveCLI = 2
)

func (c *CurveCLI) String() string {
	switch *c {
	case ED25519:
		return "ED25519"
	case P256:
		return "P256"
	case SECP256K1:
		return "SECP256K1"
	default:
		return ""
	}
}

func (c *CurveCLI) Set(value string) error {
	switch value {
	case "ED25519":
		*c = ED25519
	case "P256":
		*c = P256
	case "SECP256K1":
		*c = SECP256K1
	default:
		return errors.New("invalid Curve value")
	}
	return nil
}

func (c *CurveCLI) GetCurve() (archethic.Curve, error) {
	switch *c {
	case ED25519:
		return archethic.ED25519, nil
	case P256:
		return archethic.P256, nil
	case SECP256K1:
		return archethic.SECP256K1, nil
	default:
		return archethic.ED25519, errors.New("invalid Curve value")
	}
}

func (c *CurveCLI) Type() string {
	return "CurveCLI"
}

type TransactionTypeCLI uint8

const (
	KeychainAccessType TransactionTypeCLI = 254
	KeychainType       TransactionTypeCLI = 255
	TransferType       TransactionTypeCLI = 253
	HostingType        TransactionTypeCLI = 252
	TokenType          TransactionTypeCLI = 251
	DataType           TransactionTypeCLI = 250
	ContractType       TransactionTypeCLI = 249
	CodeProposalType   TransactionTypeCLI = 5
	CodeApprovalType   TransactionTypeCLI = 6
)

func (tt *TransactionTypeCLI) String() string {
	switch *tt {
	case KeychainAccessType:
		return "keychain_access"
	case KeychainType:
		return "keychain"
	case TransferType:
		return "transfer"
	case HostingType:
		return "hosting"
	case TokenType:
		return "token"
	case DataType:
		return "data"
	case ContractType:
		return "contract"
	case CodeProposalType:
		return "code_proposal"
	case CodeApprovalType:
		return "code_approval"
	default:
		return ""
	}
}

func (tt *TransactionTypeCLI) Set(value string) error {
	switch value {
	case "keychain_access":
		*tt = KeychainAccessType
	case "keychain":
		*tt = KeychainType
	case "transfer":
		*tt = TransferType
	case "hosting":
		*tt = HostingType
	case "token":
		*tt = TokenType
	case "data":
		*tt = DataType
	case "contract":
		*tt = ContractType
	case "code_proposal":
		*tt = CodeProposalType
	case "code_approval":
		*tt = CodeApprovalType
	default:
		return errors.New("invalid TransactionType value")
	}
	return nil
}

func (tt *TransactionTypeCLI) Type() string {
	return "TransactionTypeCLI"
}

func (tt *TransactionTypeCLI) GetTransactionType() (archethic.TransactionType, error) {
	switch *tt {
	case KeychainAccessType:
		return archethic.KeychainAccessType, nil
	case KeychainType:
		return archethic.KeychainType, nil
	case TransferType:
		return archethic.TransferType, nil
	case HostingType:
		return archethic.HostingType, nil
	case TokenType:
		return archethic.TokenType, nil
	case DataType:
		return archethic.DataType, nil
	case ContractType:
		return archethic.ContractType, nil
	case CodeProposalType:
		return archethic.CodeProposalType, nil
	case CodeApprovalType:
		return archethic.CodeApprovalType, nil
	default:
		return archethic.TransferType, errors.New("invalid TransactionType value")
	}
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
func main() {
	generateAddressCmd.Flags().String("seed", "", "Seed")
	generateAddressCmd.Flags().Int("index", 0, "Index")
	generateAddressCmd.Flags().Var(&hashAlgo, "hash-algorithm", "Hash Algorithm (SHA256|SHA512|SHA3_256|SHA3_512|BLAKE2B)")
	generateAddressCmd.Flags().Var(&ellipticCurve, "elliptic-curve", "Elliptic Curve (ED25519|P256|SECP256K1)")

	sendTransactionCmd.Flags().String("config", "", "The file location of the YAML configuration file")
	sendTransactionCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	sendTransactionCmd.Flags().String("access-seed", "", "Access Seed")
	sendTransactionCmd.Flags().Int("index", 0, "Index")
	sendTransactionCmd.Flags().Var(&ellipticCurve, "elliptic-curve", "Elliptic Curve (ED25519|P256|SECP256K1)")
	sendTransactionCmd.Flags().Var(&transactionType, "transaction-type", "Transaction Type (keychain_access|keychain|transfer|hosting|token|data|contract|code_proposal|code_approval)")
	sendTransactionCmd.Flags().StringToString("uco-transfers", map[string]string{}, "UCO Transfers (format: to=amount)")
	sendTransactionCmd.Flags().StringToString("token-transfers", map[string]string{}, "Token Transfers (format: to=amount,token_address,token_id)")
	sendTransactionCmd.Flags().StringSlice("recipients", []string{}, "Recipients")
	sendTransactionCmd.Flags().StringToString("ownerships", map[string]string{}, "Ownerships (format: secret=authorization_key)")
	sendTransactionCmd.Flags().String("content", "", "The file location of the content")
	sendTransactionCmd.Flags().String("smart-contract", "", "The file location containing the smart Contract")
	sendTransactionCmd.Flags().String("serviceName", "", "Service Name (required if creating a transaction for a service)")

	createKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	createKeychainCmd.Flags().String("access-seed", "", "Access Seed")

	getKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	getKeychainCmd.Flags().String("access-seed", "", "Access Seed")

	addServiceToKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	addServiceToKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	addServiceToKeychainCmd.Flags().String("service-name", "", "Service Name")
	addServiceToKeychainCmd.Flags().String("derivation-path", "", "Derivation Path")

	deleteServiceFromKeychainCmd.Flags().Var(&endpoint, "endpoint", "Endpoint (local|testnet|mainnet|[custom url])")
	deleteServiceFromKeychainCmd.Flags().String("access-seed", "", "Access Seed")
	deleteServiceFromKeychainCmd.Flags().String("service-name", "", "Service Name")

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
