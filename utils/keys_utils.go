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

func Gen(id string) {
	keyspath := os.Getenv(KeysPathKey)
	savePrivateFileTo := keyspath + "/" + id + ".priv"
	savePublicFileTo := keyspath + "/" + id + ".pub"
	bitSize := 4096

	privateKey, err := GeneratePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes, publicKeyBytes := encodeKeysToPem(privateKey, &privateKey.PublicKey)

	err = writeKeyToFile(privateKeyBytes, savePrivateFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = writeKeyToFile([]byte(publicKeyBytes), savePublicFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetPrivateKey(id string) *rsa.PrivateKey {
	keyspath := os.Getenv(KeysPathKey)
	readPrivKey, err := ioutil.ReadFile(keyspath + "/" + id + ".priv")

	if err != nil {
		log.Fatal("The private key is not where it should be")
	}

	pemDecodedPrivKey, rest := pem.Decode(readPrivKey)

	if len(rest) > 0 {
		log.Fatal("Error on decoding private key; the rest is not empty.")
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(pemDecodedPrivKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing private key " + err.Error())
	}

	return rsaPrivateKey
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
	keyspath := os.Getenv(KeysPathKey)
	readPubKey, err := ioutil.ReadFile(keyspath  + "/" + id + ".pub")
	if err != nil {
		log.Fatal("The public key is not where it should be")
	}

	pemDecodedPubKey, rest := pem.Decode(readPubKey)

	if len(rest) > 0 {
		log.Fatal("Error on decoding public key; the rest is not empty.")
	}

	rsaPrivateKey, err := x509.ParsePKCS1PublicKey(pemDecodedPubKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing public key " + err.Error())
	}

	return rsaPrivateKey
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

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", saveFileTo)
	return nil
}
