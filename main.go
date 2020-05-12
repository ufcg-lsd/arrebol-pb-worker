package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/worker"
	"log"
	"net/http"
	"os"
)

func setup(serverEndPoint string, workerId string) {
	log.Println("Starting to gen rsa key pair with workerid: " + workerId)

	worker.Gen(workerId)
	log.Println("Sending pub key to the server")
	url := serverEndPoint + "/workers/publicKey"
	requestBody, err := json.Marshal(&map[string]*rsa.PublicKey{"key": worker.GetPublicKey(workerId)})

	log.Println("url: " + url)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	log.Println(resp.Body)
	defer resp.Body.Close()
	log.Println("done")
}

func isTokenValid(token string) bool {
	return true
}

func main() {
	switch len(os.Args) {
	case 2:
		workerImpl := os.Args[1]
		println(workerImpl)
	default:
		defaultWorker()
	}
}

func defaultWorker() {
	workerInstance := worker.LoadWorker()
	log.Println(workerInstance)
	setup(workerInstance.ServerEndPoint, workerInstance.Id)
	workerInstance.Subscribe()
	for {
		if !isTokenValid(workerInstance.Token) {
			workerInstance.Subscribe()
		}
		//task := worker.getTask()
		// worker.execTask(task)
	}
}
