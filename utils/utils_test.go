package utils

import (
	"encoding/json"
	"testing"
)

const (
	WorkerId = "1023"
)

func TestGetPublicKey(t *testing.T) {
	//setup
	Gen(WorkerId)

	//exercise
	publicKey := GetPublicKey(WorkerId)

	//verification
	if publicKey == nil {
		t.Errorf("Error on retrieving created public key")
	}
}

func TestGetPrivateKey(t *testing.T) {
	//setup
	Gen(WorkerId)

	//exercise
	privateKey := GetPrivateKey(WorkerId)

	//verification
	if privateKey == nil {
		t.Errorf("Error on retrieving created private key")
	}
}

func TestSignMessage(t *testing.T) {
	//setup
	Gen(WorkerId)

	mockedData := make(map[string]string)

	marshalledData, _ := json.Marshal(mockedData)

	//exercise
	signedWorker, hashSum := SignMessage(GetPrivateKey(WorkerId), marshalledData)

	//verification
	if ! VerifySignature(GetPublicKey(WorkerId), hashSum, signedWorker) {
		t.Errorf("Signature verification doesnt match the specifications")
	}
}