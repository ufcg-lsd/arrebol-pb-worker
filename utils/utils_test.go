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


func TestGetPublicKey(t *testing.T) {
	//setup
	setup()
	GenAccessKeys(WorkerId)

	//exercise
	publicKey := GetPublicKey(WorkerId)

	//verification
	if publicKey == nil {
		t.Errorf("Error on retrieving created public key")
	}
}

func TestGetPrivateKey(t *testing.T) {
	//setup
	setup()
	GenAccessKeys(WorkerId)

	//exercise
	privateKey := GetPrivateKey(WorkerId)

	//verification
	if privateKey == nil {
		t.Errorf("Error on retrieving created private key")
	}
}

func TestSignMessage(t *testing.T) {
	//setup
	setup()
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