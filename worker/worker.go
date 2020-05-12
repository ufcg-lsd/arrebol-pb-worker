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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type TaskState uint8

const (
	TaskPending TaskState = iota
	TaskRunning
	TaskFinished
	TaskFailed
)

type Task struct {
	Commands       []string
	ReportInterval int64
	State          TaskState
	Progress       int
}

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

func reportReq(w *PBWorker, task *Task) {
	url := w.ServerEndPoint + "/workers/" + w.Id + "/queues/" + w.QueueId + "/tasks"
	requestBody, err := json.Marshal(task)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	req.Header.Set("arrebol-worker-token", w.Token)

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	_, err = client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
}

func (w *PBWorker) getTask() *Task {
	if w.QueueId == "" {
		log.Println("The queue id has not been set yet")
		return nil
	}

	if w.Token == "" {
		log.Println("The token has not been set yet")
		return nil
	}

	url := w.ServerEndPoint + "/workers/" + w.Id + "/queues/" + w.QueueId
	requestBody, err := json.Marshal(&PBWorker{Ram: w.Ram, Vcpu: w.Vcpu, Image: w.Image, Address: w.Address, Id: w.Id})

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(requestBody))
	req.Header.Set("arrebol-worker-token", w.Token)

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	defer resp.Body.Close()

	reqBody, err := ioutil.ReadAll(resp.Body)
	var task Task
	json.Unmarshal(reqBody, &task)

	return &task
}

func getPrivateKey(id string) *rsa.PrivateKey {
	readPrivKey, err := ioutil.ReadFile(getPrjPath() + "worker/keys/" + id + ".priv")
	if err != nil {
		log.Fatal("The private key is not where it should be")
	}

	pemDecodedPrivKey, _ := pem.Decode(readPrivKey)

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(pemDecodedPrivKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing private key1 " + err.Error())
	}

	return rsaPrivateKey
}

func signMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, []byte) {
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

func GetPublicKey(id string) *rsa.PublicKey {
	readPubKey, err := ioutil.ReadFile(getPrjPath() + "worker/keys/" + id + ".pub")
	if err != nil {
		log.Fatal("The private key is not where it should be")
	}

	pemDecodedPubKey, _ := pem.Decode(readPubKey)

	rsaPrivateKey, err := x509.ParsePKCS1PublicKey(pemDecodedPubKey.Bytes)
	if err != nil {
		log.Fatal("Error on parsing public key1 " + err.Error())
	}

	return rsaPrivateKey
}

func verifySignature(key *rsa.PublicKey, hash []byte, signature []byte) bool {
	err := rsa.VerifyPSS(key, crypto.SHA256, hash, signature, nil)
	if err != nil {
		return false
	}
	return true
}

func (w *PBWorker) Subscribe() {
	requestBody, err := json.Marshal(&PBWorker{Ram: w.Ram, Vcpu: w.Vcpu, Image: w.Image, Address: w.Address, Id: w.Id})

	data, hashSum := signMessage(getPrivateKey(w.Id), requestBody)

	payload, _ := json.Marshal(&map[string][]byte{"data": data, "hashSum": hashSum})

	url := w.ServerEndPoint + "/workers"
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	defer resp.Body.Close()

	reqBody, err := ioutil.ReadAll(resp.Body)
	var parsedBody map[string]string
	json.Unmarshal(reqBody, &parsedBody)

	w.Token = parsedBody["arrebol-worker-token"]
	w.QueueId = parsedBody["queue_id"]
}

func getPrjPath() string {
	path_cmd := exec.Command("/bin/sh", "-c", "echo $GOPATH")
	path, _ := path_cmd.Output()
	path_str := strings.TrimSpace(string(path))
	return path_str + "/src/github.com/ufcg-lsd/arrebol-pb-worker/"
}

func LoadWorker() PBWorker {

	log.Println("Starting reading configuration process")
	// it must open the port and make all scripts executable
	file, err := os.Open(getPrjPath()+ "worker/worker-conf.json")
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

	log.Println("Worker: " + configuration.Vcpu)
	log.Println(configuration)

	return configuration
}
