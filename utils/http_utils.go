package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	SIGNATURE_KEY_PATTERN = "Signature"
)

var (
	Client       HTTPClient                                        = &http.Client{}
	GetSignature func(payload interface{}, workerId string) []byte = getSignature
)

type HttpResponse struct {
	Body       []byte
	Headers    http.Header
	StatusCode int
}

func getSignature(payload interface{}, workerId string) []byte {
	parsedPayload, err := json.Marshal(payload)

	if err != nil {
		log.Fatal("Error on marshalling the payload")
	}

	signature, _ := SignMessage(GetPrivateKey(workerId), parsedPayload)

	return signature
}

func AddSignature(workerId string, payload interface{}, headers http.Header) http.Header {
	signature := GetSignature(payload, workerId)
	strSignature := fmt.Sprintf("%v", signature)
	headers.Set(SIGNATURE_KEY_PATTERN, strSignature)
	return headers
}

func Post(workerId string, body interface{}, headers http.Header, endpoint string) (*HttpResponse, error) {
	headers = AddSignature(workerId, body, headers)

	requestBody, err := json.Marshal(body)

	if err != nil {
		log.Fatal("Unable to marshal body")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := Client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return &HttpResponse{Body: respBody, Headers: resp.Header, StatusCode: resp.StatusCode}, nil
}

func Get(workerId string, endpoint string, header http.Header) (*HttpResponse, error) {
	header = AddSignature(workerId, endpoint, header)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header = header

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("The following error occurred on parsing response body: " + err.Error())
		return &HttpResponse{nil, resp.Header, resp.StatusCode}, err
	}

	return &HttpResponse{Body: respBody, Headers: resp.Header, StatusCode: resp.StatusCode}, nil
}

func Put(workerId string, body interface{}, headers http.Header, endpoint string) (*HttpResponse, error) {
	headers = AddSignature(workerId, body, headers)

	requestBody, err := json.Marshal(body)

	if err != nil {
		return nil, errors.New("Unable to marshal body")
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(requestBody))

	if err != nil {
		return nil, err
	}

	req.Header = headers
	resp, err := Client.Do(req)

	if err != nil {
		return nil, errors.New("Unable to reach the server on endpoint: " + endpoint)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.New("Error on parsing the body to byte")
	}

	return &HttpResponse{Body: respBody, Headers: resp.Header, StatusCode: resp.StatusCode}, nil
}
