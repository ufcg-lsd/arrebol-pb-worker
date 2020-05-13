package worker

import (
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"log"
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

func (w *PBWorker) Subscribe() {
	httpResponse := utils.Post(w.Id, &PBWorker{Ram: w.Ram, Vcpu: w.Vcpu, Image: w.Image,
		Address: w.Address, Id: w.Id, QueueId: w.QueueId}, w.ServerEndPoint + "/workers")

	if httpResponse.StatusCode != 201 {
		log.Fatal("The work could not be subscribed")
	}

	var parsedBody map[string]string
	json.Unmarshal(httpResponse.Body, &parsedBody)

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
