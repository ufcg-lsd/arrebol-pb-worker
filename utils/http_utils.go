package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient = &http.Client{}
)

type HttpResponse struct {
	Body []byte
	Headers http.Header
	StatusCode int
}

// It send an http post to the endpoint signing the body with the worker's private key
func Post(workerId string, body interface{}, endpoint string) *HttpResponse{
	requestBody, err := json.Marshal(body)
	data, hashSum := SignMessage(GetPrivateKey(workerId), requestBody)

	payload, _ := json.Marshal(&map[string][]byte{"data": data, "hashSum": hashSum})

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payload))

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := Client.Do(req)
	if err != nil {
		// handle error
		log.Fatal("Unable to reach the server on endpoint: " + endpoint)
		panic(err)
	}

	resposeBody, _ := ioutil.ReadAll(resp.Body)

	return &HttpResponse{Body: resposeBody, Headers: resp.Header, StatusCode: resp.StatusCode}
}
