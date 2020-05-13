package worker

import (
	"bytes"
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"testing"
)

var (
	pbWorkerTestInstance = PBWorker{
		ServerEndPoint: "http://localhost:8000/v1",
		Vcpu:           "1",
		Ram:            "3",
		Image:          "ubuntu",
		Address:        "10.11.19.9",
		Token:          "test-token",
		Id:             "1023",
		QueueId:        "0932",
	}
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func TestLoadWorker(t *testing.T) {
 	//setup
	setup()

	//exercise
	workerInstance := LoadWorker()

	//verification
	if workerInstance != pbWorkerTestInstance {
		t.Errorf("The loaded worker is different from the expected one")
	}

	//cleanup
	cleanup()
}

func TestPBWorker_Subscribe(t *testing.T) {
	//setup
	setup()
	utils.Client = &MockClient{}

	body := make(map[string]string)
	body["arrebol-worker-token"] = "test-token"
	body["queue_id"] = "192038"

	bodyAsByte, _ := json.Marshal(body)
	parsedBody := ioutil.NopCloser(bytes.NewReader(bodyAsByte))

	GetDoFunc =  func(*http.Request) (*http.Response, error) {
		return &http.Response{Body: parsedBody, StatusCode: 201}, nil
	}

	//exercise
	pbWorkerTestInstance.Subscribe()

	//verification
	if pbWorkerTestInstance.QueueId != "192038" {
		t.Errorf("QueueId is not the expected one")
	}

	if pbWorkerTestInstance.Token != "test-token" {
		t.Errorf("The token is not the expected one")
	}

	//cleanup
	cleanup()
}

//Test utils functions
func setup()  {
	utils.Gen("1023")

	file, _ := json.MarshalIndent(pbWorkerTestInstance, "", " ")

	exec.Command("bash", "-c", "mv " + utils.GetPrjPath() + "worker/worker-conf.json " +
		utils.GetPrjPath() +"worker/worker-orginal-conf.json")

	_ = ioutil.WriteFile(utils.GetPrjPath()+"worker/worker-conf.json", file, 0644)
}

func cleanup() {
	exec.Command("bash", "-c", "mv " + utils.GetPrjPath() + "worker/worker-original-conf.json " +
		utils.GetPrjPath() +"worker/worker-conf.json")

	keysPath := utils.GetPrjPath() + "worker/keys/*"
	cmd := exec.Command("bash", "-c", "rm " + keysPath)
	log.Print(cmd.Output())
}