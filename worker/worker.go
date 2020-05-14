package worker

import (
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io"
	"log"
)

type Worker struct {
	Vcpu           string
	Ram            string
	Token          string
	Id             string
	QueueId        string
}

func (w *Worker) Subscribe(serverEndpoint string) {
	httpResponse := utils.SignedPost(w.Id, w, serverEndpoint + "/workers")
	HandleSubscriptionResponse(httpResponse, w)
}

func HandleSubscriptionResponse(response *utils.HttpResponse, w *Worker) {
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

func ParseWorkerConfiguration(reader io.Reader) Worker {
	decoder := json.NewDecoder(reader)
	configuration := Worker{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println("Error on decoding configuration file", err.Error())
	}

	return configuration
}
