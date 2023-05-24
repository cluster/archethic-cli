package main

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/archethic-foundation/archethic-cli/tui"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	args := os.Args[1:]
	var pvKeyBytes []byte
	if len(args) > 0 && args[0] != "" {
		privateKeyPath := args[0]
		// Read the private key file
		privateBytes, err := ioutil.ReadFile(privateKeyPath)
		if err != nil {
			log.Fatalf("Failed to load private key: %v", err)
		}

		pvKey, err := ssh.ParseRawPrivateKey(privateBytes)

		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			passphrase := promptPassphrase()
			pvKey, err = ssh.ParseRawPrivateKeyWithPassphrase(privateBytes, []byte(passphrase))
			if err != nil {
				log.Fatalf("Failed to parse private key: %v", err)
			}
		}

		switch pvKey := pvKey.(type) {
		case *rsa.PrivateKey:
			pvKeyBytes = pvKey.D.Bytes()
		case *ecdsa.PrivateKey:
			pvKeyBytes = pvKey.D.Bytes()
		case *dsa.PrivateKey:
			pvKeyBytes = pvKey.X.Bytes()
		default:
			log.Fatalf("Only RSA, ECDSA and DSA keys are supported, got %T", pvKey)
		}

	}

	tui.StartTea(pvKeyBytes)
}

func promptPassphrase() string {
	fmt.Print("Enter passphrase: ")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read passphrase: %v", err)
	}
	fmt.Println()
	return string(passphrase)
}
