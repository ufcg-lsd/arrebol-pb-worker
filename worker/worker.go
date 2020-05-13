package worker

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type PBWorker struct {
	ServerEndPoint string
	Vcpu           string
	Ram            string
	Image          string
	Address        string
	Token          string
	Id             string
	QueueId        string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient = &http.Client{}
)


func GetPrivateKey(id string) *rsa.PrivateKey {
	readPrivKey, err := ioutil.ReadFile(utils.GetPrjPath() + "worker/keys/" + id + ".priv")
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
	readPubKey, err := ioutil.ReadFile(utils.GetPrjPath()  + "worker/keys/" + id + ".pub")
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


func (w *PBWorker) Subscribe() {
	requestBody, err := json.Marshal(&PBWorker{Ram: w.Ram, Vcpu: w.Vcpu, Image: w.Image, Address: w.Address, Id: w.Id, QueueId: w.QueueId})

	data, hashSum := SignMessage(GetPrivateKey(w.Id), requestBody)

	payload, _ := json.Marshal(&map[string][]byte{"data": data, "hashSum": hashSum})

	url := w.ServerEndPoint + "/workers"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := Client.Do(req)
	if err != nil {
		// handle error
		log.Fatal("Unable to subscribe in the server")
		panic(err)
	}

	defer resp.Body.Close()

	reqBody, err := ioutil.ReadAll(resp.Body)
	var parsedBody map[string]string
	json.Unmarshal(reqBody, &parsedBody)

	w.Token = parsedBody["arrebol-worker-token"]
	w.QueueId = parsedBody["queue_id"]
}

func LoadWorker() PBWorker {

	log.Println("Starting reading configuration process")

	file, err := os.Open(utils.GetPrjPath() + "worker/worker-conf.json")
	defer file.Close()
	if err != nil {
		log.Println("Error on opening configuration file", err.Error())
	}

	decoder := json.NewDecoder(file)
	configuration := PBWorker{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Println("Error on decoding configuration file", err.Error())
	}

	return configuration
}
