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
)

func Gen(id string) {
	keyspath := GetPrjPath() + "worker/keys"
	savePrivateFileTo := keyspath + "/" + id + ".priv"
	savePublicFileTo := keyspath + "/" + id + ".pub"
	bitSize := 4096

	privateKey, err := GeneratePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes := generatePublicKey(&privateKey.PublicKey)

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

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
	readPrivKey, err := ioutil.ReadFile(GetPrjPath() + "worker/keys/" + id + ".priv")
	if err != nil {
		log.Fatal("The private key is not where it should be")
	}

	pemDecodedPrivKey, _ := pem.Decode(readPrivKey)

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(pemDecodedPrivKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing private key " + err.Error())
	}

	return rsaPrivateKey
}

func SignMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, []byte) {
	messageHash := sha256.New()
	_, err := messageHash.Write(message)
	if err != nil {
		panic(err)
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
	readPubKey, err := ioutil.ReadFile(GetPrjPath()  + "worker/keys/" + id + ".pub")
	if err != nil {
		log.Fatal("The public key is not where it should be")
	}

	pemDecodedPubKey, _ := pem.Decode(readPubKey)

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

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePublicKeyToPEM(publicKey *rsa.PublicKey) []byte {
	// Get ASN.1 DER format
	publicDER := x509.MarshalPKCS1PublicKey(publicKey)

	// pem.Block
	publicBlock := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   publicDER,
	}

	// Private key in PEM format
	publicPEM := pem.EncodeToMemory(&publicBlock)

	return publicPEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) []byte {
	return encodePublicKeyToPEM(privatekey)

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
