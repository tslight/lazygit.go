package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func GetSSHPubKey() []byte {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}

	savePrivateFileTo := fmt.Sprintf("%s/.ssh/id_rsa", homeDir)
	savePublicFileTo := fmt.Sprintf("%s/.ssh/id_rsa.pub", homeDir)

	if _, err := os.Stat(savePublicFileTo); err == nil {
		log.Printf("Found %s, so using that and not generating new keys.", savePublicFileTo)
		publicKeyBytes, err := os.ReadFile(savePublicFileTo)
		if err != nil {
			log.Fatal(err.Error())
		}
		return publicKeyBytes
	}

	return writeNewKeys(savePrivateFileTo, savePublicFileTo)
}

func writeNewKeys(privateKeyPath, publicKeyPath string) []byte {
	bitSize := 4096
	err := os.MkdirAll(path.Dir(privateKeyPath), os.ModePerm)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	err = writeKeyToFile(privateKeyBytes, privateKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = writeKeyToFile([]byte(publicKeyBytes), publicKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	return publicKeyBytes
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", saveFileTo)
	return nil
}
