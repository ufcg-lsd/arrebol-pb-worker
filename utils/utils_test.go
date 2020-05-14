package utils

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

const (
	WorkerId = "1023"
)

func setup() {
	err := godotenv.Load("../.env")

	if err != nil {
		log.Println("No .env file found")
	}
}

func TestModule(t *testing.T) {
	setup()
	t.Run("TestGetPublicKey", getPublicKey)
	t.Run("TestGetPublicKey", getPrivateKey)
	t.Run("TestGetPublicKey", signMessage)
}

func getPublicKey(t *testing.T) {
	//setup
	GenAccessKeys(WorkerId)

	//exercise
	publicKey := GetPublicKey(WorkerId)

	//verification
	if publicKey == nil {
		t.Errorf("Error on retrieving created public key")
	}
}

func getPrivateKey(t *testing.T) {
	//setup
	GenAccessKeys(WorkerId)

	//exercise
	privateKey := GetPrivateKey(WorkerId)

	//verification
	if privateKey == nil {
		t.Errorf("Error on retrieving created private key")
	}
}

func signMessage(t *testing.T) {
	//setup
	GenAccessKeys(WorkerId)

	mockedData := make(map[string]string)

	marshalledData, err := json.Marshal(mockedData)

	if err != nil {
		t.Errorf("Error on mashalling the mockedData")
	}

	//exercise
	signedWorker, hashSum := SignMessage(GetPrivateKey(WorkerId), marshalledData)

	//verification
	if ! VerifySignature(GetPublicKey(WorkerId), hashSum, signedWorker) {
		t.Errorf("Signature verification doesnt match the specifications")
	}
}