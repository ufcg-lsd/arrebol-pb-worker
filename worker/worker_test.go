package worker

import (
	"bytes"
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

var (
	workerTestInstance = Worker{
		Vcpu:    1,
		Ram:     3,
		Token:   "test-token",
		Id:      "1023",
		QueueId: "0932",
	}
)

type MockedClient struct {
}

var (
	GetDo func() (*http.Response, error)
)

func (c *MockedClient) Do(req *http.Request) (*http.Response, error) {
	return GetDo()
}

func TestParseWorkerConfiguration(t *testing.T) {
	testingWorkerAsByte, err := json.Marshal(workerTestInstance)

	if err != nil {

		t.Errorf("Error on bytefying test worker")

	}

	parsedWorker := ParseWorkerConfiguration(bytes.NewReader(testingWorkerAsByte))

	if parsedWorker != workerTestInstance {
		t.Errorf("The parsed worked is different from the expected one")
	}
}

func TestHandleSubscriptionResponse(t *testing.T) {
	//setup
	body := make(map[string]string)
	body["arrebol-worker-token"] = "test-token"
	body["QueueId"] = "192038"

	bodyAsByte, _ := json.Marshal(body)

	ParseToken = func(tokenStr string) (map[string]interface{}, error) {
		return map[string]interface{}{"QueueId": "192038"}, nil
	}

	//exercise
	HandleJoinResponse(&utils.HttpResponse{Body: bodyAsByte, StatusCode: 201}, &workerTestInstance)

	//verification
	if workerTestInstance.QueueId != "192038" {
		t.Errorf("QueueId is not the expected one")
	}

	if workerTestInstance.Token != "test-token" {
		t.Errorf("The token is not the expected one")
	}
}

func TestWorker_GetTask(t *testing.T) {
	//setup
	task := make(map[string]string)
	task["Id"] = "1"

	byteTask, err := json.Marshal(&task)

	if err != nil {
		log.Fatal("Error on marshalling test task")
	}

	body := ioutil.NopCloser(bytes.NewReader(byteTask))

	GetDo = func() (*http.Response, error) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     nil,
			Body:       body,
		}
		return resp, nil
	}

	utils.Client = &MockedClient{}

	utils.GetSignature = func(payload interface{}, workerId string) []byte {
		fakeSignature, _ := json.Marshal("FAKE-SIGNATURE")
		return fakeSignature
	}

	//exercise
	mockedTask, err := workerTestInstance.GetTask("http://test-server:8000/v1")

	//verify
	if err != nil {
		t.Error("Error on getting task: " + err.Error())
	}

	if mockedTask.Id != "1" {
		t.Error("The task Id is different from the expected one")
	}
}

func TestWorker_GetTaskWithEmptyQueue(t *testing.T) {
	//setup
	workerTestInstance.QueueId = ""

	//exercise
	mockedTask, err := workerTestInstance.GetTask("http://test-server:8000/v1")

	//verify
	if err == nil {
		t.Error("The expected error has not occurred")
	}

	if mockedTask != nil {
		t.Error("The expected error has not occurred")
	}
}
