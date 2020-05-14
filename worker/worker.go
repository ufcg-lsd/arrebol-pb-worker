package worker

import (
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io"
	"log"
	"os"
)

const (
	ConfFilePathKey = "CONF_FILE"
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

func (w *PBWorker) Subscribe() {
	httpResponse := utils.SignedPost(w.Id, &PBWorker{Ram: w.Ram, Vcpu: w.Vcpu, Image: w.Image,
		Address: w.Address, Id: w.Id, QueueId: w.QueueId}, w.ServerEndPoint + "/workers")

	HandleSubscriptionResponse(httpResponse, w)
}

func HandleSubscriptionResponse(response *utils.HttpResponse, w *PBWorker) {
	if response.StatusCode != 201 {
		log.Fatal("The work could not be subscribed")
	}

	var parsedBody map[string]string
	err := json.Unmarshal(response.Body, &parsedBody)

	if err != nil {
		log.Fatal("Unable to parse the response body")
	}

	token, ok := parsedBody["arrebol-worker-token"]

	if !ok {
		log.Fatal("The token is not in the response body")
	}

	queueId, ok := parsedBody["queue_id"]

	if !ok {
		log.Fatal("The queue_id is not in the response body")
	}

	w.Token = token
	w.QueueId = queueId
}

func LoadWorker() PBWorker {
	log.Println("Starting reading configuration process")

	file, err := os.Open(os.Getenv(ConfFilePathKey))

	if err != nil {
		log.Fatal("Error on opening configuration file", err.Error())
	}

	defer file.Close()

	return ParseWorkerConfiguration(file)
}

func ParseWorkerConfiguration(reader io.Reader) PBWorker{
	decoder := json.NewDecoder(reader)
	configuration := PBWorker{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println("Error on decoding configuration file", err.Error())
	}

	return configuration
}
