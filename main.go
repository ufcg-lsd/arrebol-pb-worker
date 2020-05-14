package main

import (
	"crypto/rsa"
	"github.com/joho/godotenv"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"github.com/ufcg-lsd/arrebol-pb-worker/worker"
	"log"
	"os"
)

const (
	ConfFilePathKey = "CONF_FILE"
	ServerEndpointKey = "SERVER_ENDPOINT"
)

func setup(serverEndPoint string, workerId string) {
	// The setup routine is responsible for generates the rsa key pairs
	// of the worker and to send the public part to the server.
	log.Println("Starting to gen rsa key pair with workerid: " + workerId)

	utils.GenAccessKeys(workerId)

	log.Println("Sending pub key to the server")
	url := serverEndPoint + "/workers/publicKey"
	requestBody := &map[string]*rsa.PublicKey{"key": utils.GetPublicKey(workerId)}
	httpResp := utils.Post(requestBody, url)

	if httpResp.StatusCode != 201 {
		log.Fatal("Unable to send public key to the server")
	}
}

func isTokenValid(token string) bool {
	return true
}

func main() {
	// this main function start the worker following the chosen implementation
	// passed by arg in the cli. The defaultWorker is started if no arg has been received.
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	switch len(os.Args) {
	case 2:
		workerImpl := os.Args[1]
		println(workerImpl)
	default:
		defaultWorker()
	}
}

func defaultWorker() {
	// This is the default work behavior implementation.
	// Its core stands for executing one task at a time.
	log.Println("Starting reading configuration process")
	file, err := os.Open(os.Getenv(ConfFilePathKey))

	if err != nil {
		log.Fatal("Error on opening configuration file", err.Error())
	}

	defer file.Close()

	workerInstance := worker.ParseWorkerConfiguration(file)

	serverEndpoint := os.Getenv(ServerEndpointKey)

	setup(serverEndpoint, workerInstance.Id)

	workerInstance.Subscribe(serverEndpoint)

	for {
		if !isTokenValid(workerInstance.Token) {
			workerInstance.Subscribe(serverEndpoint)
		}
	}
}
