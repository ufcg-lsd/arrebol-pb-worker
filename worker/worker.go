package worker

import (
	"bytes"
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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

type Task struct {
	Commands       []string
	ReportInterval int64
	State          TaskState
	Progress       int
	Image string
	Id string
}

const (
	WorkerNodeAddressKey = "WORKER_NODE_ADDRESS"
)

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

func ParseWorkerConfiguration(reader io.Reader) Worker {
	decoder := json.NewDecoder(reader)
	configuration := Worker{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println("Error on decoding configuration file", err.Error())
	}

	return configuration
}

func (w *Worker) ExecTask(task *Task, serverEndPoint string) {
	address := os.Getenv(WorkerNodeAddressKey)
	client := utils.NewDockerClient(address)
	taskExecutor := &TaskExecutor{Cli: *client}
	endingChannel := make(chan interface{}, 1)
	waitChannel := make(chan interface{}, 1)
	spawnWarner := make(chan interface{}, 1)
	reportChannels := []chan interface{}{endingChannel, waitChannel, spawnWarner}
	go w.reportTask(task, taskExecutor, reportChannels, serverEndPoint)
	err := taskExecutor.Execute(task, spawnWarner)
	log.Println(err)
	log.Println("Wrting in the endingCHannel")
	endingChannel <- "done"
	log.Println("Reading from the wait channel")
	<-waitChannel
	log.Println("Dieing")
}

func (w *Worker) reportTask(task *Task, executor *TaskExecutor, channels []chan interface{},serverEndPoint string) {
	startTime := time.Now().Unix()
	for {
		log.Println("going to start select block")
		select {
		case <-channels[0]:
			task.Progress = 100
			//reportReq(w, task, serverEndPoint)
			channels[1] <- "done"
			return
		default:
			log.Println("default option has been chosen")
			<-channels[2]
			progress, err := executor.Track()

			if err != nil {
				log.Println(err)
			}

			task.Progress = progress

			log.Println("progess: " + string(progress))

			currentTime := time.Now().Unix()
			if currentTime-startTime < task.ReportInterval {
				time.Sleep(5 * time.Second)
				continue
			}

			//reportReq(w, task, serverEndPoint)

			startTime = currentTime
		}
	}
}

func reportReq(w *Worker, task *Task, serverEndPoint string) {
	url := serverEndPoint + "/workers/" + w.Id + "/queues/" + w.QueueId + "/tasks"
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

