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
func SignedPost(workerId string, body interface{}, endpoint string) *HttpResponse{
	requestBody, err := json.Marshal(body)

	if err != nil {
		log.Fatal("Error on marshalling the request body")
	}

	data, hashSum := SignMessage(GetPrivateKey(workerId), requestBody)


	return Post(&map[string][]byte{"data": data, "hashSum": hashSum}, endpoint)
}

func Post(body interface{}, endpoint string) *HttpResponse{
	requestBody, err := json.Marshal(body)

	if err != nil {
		log.Fatal("Unable to marshal body")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestBody))

	if err != nil {
		log.Fatal(err)
	}

	resp, err := Client.Do(req)

	if err != nil {
		log.Fatal("Unable to reach the server on endpoint: " + endpoint)
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error on parsing the body to byte")
	}

	return &HttpResponse{Body: respBody, Headers: resp.Header, StatusCode: resp.StatusCode}
}
