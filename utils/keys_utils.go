package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
)

const (
	KeysPathKey = "KEYS_PATH"
)

func GenAccessKeys(id string) {
	keyspath := os.Getenv(KeysPathKey)
	privateKeyPath := keyspath + "/" + id + ".priv"
	publicKeyPath := keyspath + "/" + id + ".pub"
	bitSize := 4096

	privateKey, err := GeneratePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes, publicKeyBytes := encodeKeysToPem(privateKey, &privateKey.PublicKey)

	err = saveKey(privateKeyBytes, privateKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = saveKey(publicKeyBytes, publicKeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetPrivateKey(id string) *rsa.PrivateKey {
	keyName := "/" + id + ".priv"
	decodedKey := decodeKey(keyName)

	rsaKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing private key " + err.Error())
	}

	return rsaKey
}


func decodeKey(keyName string)  *pem.Block {
	keyspath := os.Getenv(KeysPathKey)
	keyContent, err := ioutil.ReadFile(keyspath + keyName)

	if err != nil {
		log.Fatal("The private key is not where it should be")
	}

	decodedKey, rest := pem.Decode(keyContent)

	if len(rest) > 0 {
		log.Fatal("Error on decoding private key; the rest is not empty.")
	}

	return decodedKey
}

func SignMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, []byte) {
	messageHash := sha256.New()
	writtenBytesCounter, err := messageHash.Write(message)

	if err != nil {
		panic(err)
	}

	if writtenBytesCounter != len(message) {
		log.Fatal("The message has not been entirely written in the message hash.")
	}

	msgHashSum := messageHash.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, msgHashSum, nil)
	if err != nil {
		panic(err)
	}

	return signature, msgHashSum
}

func VerifySignature(key *rsa.PublicKey, hash []byte, signature []byte) bool {
	err := rsa.VerifyPSS(key, crypto.SHA256, hash, signature, nil)
	if err != nil {
		return false
	}
	return true
}

func GetPublicKey(id string) *rsa.PublicKey {
	keyName := "/" + id + ".pub"
	decodedKey := decodeKey(keyName)

	rsaKey, err := x509.ParsePKCS1PublicKey(decodedKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing public key " + err.Error())
	}

	return rsaKey
}

// GeneratePrivateKey creates a RSA Private Key of specified byte size
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
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

func encodeKeysToPem(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) ([]byte, []byte) {
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	publicDER := x509.MarshalPKCS1PublicKey(publicKey)
	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	publicBlock := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   publicDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	publicPEM := pem.EncodeToMemory(&publicBlock)

	return privatePEM, publicPEM
}

func saveKey(keyBytes []byte, filePath string) error {
	err := ioutil.WriteFile(filePath, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", filePath)
	return nil
}
