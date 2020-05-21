package worker

import (
	"encoding/json"
	"errors"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io"
	"log"
	"net/http"
)

//It represents each one of the worker's instances that will run on the worker node.
//The informations kept in this struct are important to the
//communication process with the server. While the Vcpu and Ram allow the server
//to choose better which tasks to dispatch to the worker instance, the Token, the Id
//and the QueueId are indispensable to establish the communication.
//Note: the Token is only SET when the worker joins the server.
//The QueueId can be SET during a join or in the worker's conf file.
//The others are set in the conf file.
type Worker struct {
	//The Vcpu available to the worker instance
	Vcpu          float32
	//The Ram available to the worker instance
	Ram            float32
	//The Token that the server has been assigned to the worker
	//so it is able to authenticate in next requests
	Token          string
	//The worker instance id
	Id             string
	//The queue from which the worker must ask for tasks
	QueueId        string
}

type TaskState uint8

const (
	TaskPending TaskState = iota
	TaskRunning
	TaskFinished
	TaskFailed
)

//This struct represents a task, the executable piece of the system.
type Task struct {
	// This array wraps the task's commands. Each positions of the array
	// is a unix command.
	Commands       []string
	// This field stands for the the time interval in which
	// the worker must report the task status for the server.
	// It is measured in seconds.
	ReportInterval int64
	// A very intuitive field, that represents
	// the current situation of the task and
	// ease the understanding between the worker and the server
	State          TaskState
	// It is measured in percentage, meaning
	// how much of the task has been processed yet.
	// For example, if this field hols the value 50,
	// it means that half of the task's commands have been
	// ran.
	Progress       int
	// It is the docker image in which the task must
	// be executed. (e.g library/ubuntu).
	Image string
	// The task Id is another field that aims
	// ease the communication between the worker and the server
	Id string
}

func (w *Worker) Join(serverEndpoint string) {
	httpResponse := utils.SignedPost(w.Id, w, serverEndpoint + "/workers")
	HandleJoinResponse(httpResponse, w)
}

func HandleJoinResponse(response *utils.HttpResponse, w *Worker) {
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

func (w *Worker) GetTask(serverEndPoint string) (*Task, error) {
	log.Println("Starting GetTask routine")

	if w.QueueId == "" {
		return nil, errors.New("The QueueId must be set before getting a task")
	}

	url := serverEndPoint + "/workers/" + w.Id + "/queues/" + w.QueueId + "/tasks"

	headers := http.Header{}
	headers.Set("arrebol-worker-token", w.Token)

	httpResp, err := utils.Get(url, headers)

	if err != nil {
		return nil, errors.New("Error on GET request: " + err.Error())
	}

	respBody := httpResp.Body

	var task Task
	err = json.Unmarshal(respBody, &task)

	if err != nil {
		return nil, errors.New("Error on unmarshalling the task: " + err.Error())
	}

	return &task, nil
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
