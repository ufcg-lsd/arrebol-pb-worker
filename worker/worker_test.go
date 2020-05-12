package worker

import(
	"encoding/json"
	"github.com/ufcg-lsd/arrebol-pb-worker/utils"
	"io/ioutil"
	"testing"
)

func TestLoadWorker(t *testing.T) {
 	//setup
	genJsonConf()

	workerInstance := LoadWorker()

	if workerInstance.Id != "1023" {
		t.Errorf("worker id should be 1023 but is " + workerInstance.Id)
	}
}


func genJsonConf()  {
	pbWorkerTestInstance := PBWorker{
		ServerEndPoint: "http://localhost:8000/v1",
		Vcpu:           "1",
		Ram:            "2",
		Image:          "ubuntu",
		Address:        "10.11.19.9",
		Token:          "test-token",
		Id:             "1023",
		QueueId:        "0932",
	}
	file, _ := json.MarshalIndent(pbWorkerTestInstance, "", " ")

	_ = ioutil.WriteFile(utils.GetPrjPath()+"worker/worker-conf.json", file, 0644)
}