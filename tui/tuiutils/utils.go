package tuiutils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
)

func GetHashAlgorithmName(h archethic.HashAlgo) string {
	switch h {
	case archethic.SHA256:
		return "SHA256"
	case archethic.SHA512:
		return "SHA512"
	case archethic.SHA3_256:
		return "SHA3_256"
	case archethic.SHA3_512:
		return "SHA3_512"
	case archethic.BLAKE2B:
		return "BLAKE2B"
	}
	panic("Unknown hash algorithm")
}

func GetCurveName(h archethic.Curve) string {
	switch h {
	case archethic.ED25519:
		return "ED25519"
	case archethic.P256:
		return "P256"
	case archethic.SECP256K1:
		return "SECP256K1"
	}
	panic("Unknown curve")
}

func CreateKeychain(url, seed string) (string, string, string, string, error) {
	accessSeed, err := archethic.MaybeConvertToHex(seed)
	if err != nil {
		return "", "", "", "", err
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")

	publicKey, _, err := archethic.DeriveKeypair(accessSeed, 0, archethic.ED25519)
	if err != nil {
		return "", "", "", "", err
	}

	randomSeed := make([]byte, 32)
	rand.Read(randomSeed)

	keychain := archethic.NewKeychain(randomSeed)
	keychain.AddService("uco", "m/650'/0/0", archethic.ED25519, archethic.SHA256)
	keychain.AddAuthorizedPublicKey(publicKey)

	accessAddress, err := archethic.DeriveAddress(accessSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", "", "", "", err
	}
	keychainAddress, err := archethic.DeriveAddress(randomSeed, 1, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", "", "", "", err
	}

	keychainTx, err := archethic.NewKeychainTransaction(keychain, 0)
	if err != nil {
		return "", "", "", "", err
	}
	keychainTx.OriginSign(originPrivateKey)

	client := archethic.NewAPIClient(url)
	accessKeychain, _ := archethic.GetKeychain(accessSeed, *client)
	if accessKeychain != nil {
		err = errors.New("keychain access already exists")
		return "", "", "", "", err
	}

	var error error
	feedback := ""
	keychainSeed := ""
	keychainTransactionAddress := ""
	keychainAccessTransactionAddress := ""
	ts := archethic.NewTransactionSender(client)
	ts.AddOnRequiredConfirmation(func(nbConf int) {
		feedback += "\nKeychain's transaction confirmed."

		keychainSeed = hex.EncodeToString(randomSeed)
		keychainTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, keychainAddress)

		accessTx, err := archethic.NewAccessTransaction(accessSeed, keychainAddress)
		if err != nil {
			error = err
			feedback = err.Error()
		}
		accessTx.OriginSign(originPrivateKey)
		ts2 := archethic.NewTransactionSender(client)
		ts2.AddOnRequiredConfirmation(func(nbConf int) {
			feedback += "\nKeychain access transaction confirmed."
			ts2.Unsubscribe("confirmation")
			keychainAccessTransactionAddress = fmt.Sprintf("%s/explorer/transaction/%x", url, accessAddress)
		})
		ts2.AddOnError(func(senderContext, message string) {
			feedback += fmt.Sprintf("\nAccess transaction error: %s", message)
			ts.Unsubscribe("error")
		})
		ts2.SendTransaction(accessTx, 100, 60)
		ts.Unsubscribe("confirmation")
	})
	ts.AddOnError(func(senderContext, message string) {
		feedback += fmt.Sprintf("Keychain transaction error: %s", message)
		ts.Unsubscribe("error")
		error = errors.New(message)
	})
	ts.SendTransaction(keychainTx, 100, 60)
	return feedback, keychainSeed, keychainTransactionAddress, keychainAccessTransactionAddress, error
}

func AccessKeychain(endpoint, seed string) (*archethic.Keychain, error) {
	seedByte, err := archethic.MaybeConvertToHex(seed)
	if err != nil {
		return nil, err
	}
	client := archethic.NewAPIClient(endpoint)
	return archethic.GetKeychain(seedByte, *client)
}

func AddServiceToKeychain(accessSeed []byte, endpoint string, serviceName string, serviceDerivationPath string) (string, error) {
	return updateKeychain(accessSeed, endpoint, func(keychain *archethic.Keychain) {
		keychain.AddService(serviceName, serviceDerivationPath, archethic.ED25519, archethic.SHA256)
	})
}

func RemoveServiceFromKeychain(accessSeed []byte, endpoint string, serviceName string) (string, error) {
	return updateKeychain(accessSeed, endpoint, func(keychain *archethic.Keychain) {
		keychain.RemoveService(serviceName)
	})
}

func updateKeychain(accessSeed []byte, endpoint string, updateFunc func(*archethic.Keychain)) (string, error) {
	client := *archethic.NewAPIClient(endpoint)
	keychain, err := archethic.GetKeychain(accessSeed, client)
	if err != nil {
		return "", err
	}
	updateFunc(keychain)

	keychainGenesisAddress, err := archethic.DeriveAddress(keychain.Seed, 0, archethic.ED25519, archethic.SHA256)
	if err != nil {
		return "", err
	}
	addressHex := hex.EncodeToString(keychainGenesisAddress)
	transactionChainIndex := client.GetLastTransactionIndex(addressHex)
	transaction, err := archethic.NewKeychainTransaction(keychain, uint32(transactionChainIndex))
	if err != nil {
		return "", err
	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
	transaction.OriginSign(originPrivateKey)

	var returnedError error
	var returnedFeedback = ""
	returnedError = nil

	ts := archethic.NewTransactionSender(&client)
	ts.AddOnRequiredConfirmation(func(nbConf int) {
		returnedFeedback = "\nKeychain's transaction confirmed."
	})
	ts.AddOnError(func(senderContext, message string) {
		returnedError = errors.New(message)
		ts.Unsubscribe("error")
	})
	ts.SendTransaction(transaction, 100, 60)

	return returnedFeedback, returnedError
}

func SendTransaction(transaction archethic.TransactionBuilder, secretKey []byte, curve archethic.Curve, serviceMode bool, endpoint string, transactionIndex int, serviceName string, storageNouncePublicKey string, seedStr string) (string, error) {
	feedback := ""
	seed, err := archethic.MaybeConvertToHex(seedStr)
	if err != nil {
		return "", err
	}
	if len(transaction.Data.Code) > 0 {
		ownershipIndex := -1
		for i, ownership := range transaction.Data.Ownerships {
			decryptSecret, err := archethic.AesDecrypt(ownership.Secret, secretKey)
			if err != nil {
				return "", err
			}
			decodedSecret := string(decryptSecret)

			if reflect.DeepEqual(decodedSecret, string(seed)) {
				ownershipIndex = i
				break
			}
		}

		if ownershipIndex == -1 {
			return "", errors.New("you need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract")
		} else {
			authorizedKeyIndex := -1
			for i, authKey := range transaction.Data.Ownerships[ownershipIndex].AuthorizedKeys {
				if reflect.DeepEqual(strings.ToUpper(hex.EncodeToString(authKey.PublicKey)), storageNouncePublicKey) {
					authorizedKeyIndex = i
					break
				}
			}

			if authorizedKeyIndex == -1 {
				return "", errors.New("you need to create an ownership with the transaction seed as secret and authorize node public key to let nodes generate new transaction from your smart contract")
			}
		}
	}

	client := archethic.NewAPIClient(endpoint)

	if serviceMode {
		err = buildKeychainTransaction(seed, client, transaction, serviceName)
		if err != nil {
			return "", err
		}
	} else {
		transaction.Build(seed, uint32(transactionIndex), curve, archethic.SHA256)

	}
	originPrivateKey, _ := hex.DecodeString("01019280BDB84B8F8AEDBA205FE3552689964A5626EE2C60AA10E3BF22A91A036009")
	transaction.OriginSign(originPrivateKey)

	ts := archethic.NewTransactionSender(client)
	ts.AddOnSent(func() {
		feedback = "Transaction sent: " + endpoint + "/explorer/transaction/" + strings.ToUpper(hex.EncodeToString(transaction.Address))
	})

	ts.AddOnError(func(sender, message string) {
		feedback = "Transaction error: " + message
	})

	ts.SendTransaction(&transaction, 100, 60)
	return feedback, nil
}

func buildKeychainTransaction(seed []byte, client *archethic.APIClient, transaction archethic.TransactionBuilder, serviceName string) error {
	keychain, err := archethic.GetKeychain(seed, *client)
	if err != nil {
		return err
	}

	transaction.Version = uint32(keychain.Version)

	genesisAddress, err := keychain.DeriveAddress(serviceName, 0)
	if err != nil {
		return err
	}

	index := client.GetLastTransactionIndex(hex.EncodeToString(genesisAddress))

	err = keychain.BuildTransaction(&transaction, serviceName, uint8(index))
	if err != nil {
		return err
	}
	return nil
}
