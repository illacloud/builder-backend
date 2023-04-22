package repository

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	PERIPHERAL_API_BASEURL           = "https://peripheral-api.illasoft.com/v1/"
	PERIPHERAL_API_GENERATE_SQL_PATH = "generateSQL"
	PERIPHERAL_API_ECHO_PATH         = "echo"
	PERIPHERAL_API_GLOBAL_TIMEOUT    = 5 * time.Minute
)

type GenerateSQLFeedback struct {
	Payload string `json:"payload"`
}

type GeneralFeedback struct {
	Payload string `json:"payload"`
}

func GenerateSQL(m *GenerateSQLPeripheralRequest, req *GenerateSQLRequest) (*GenerateSQLFeedback, error) {
	payload := m.Export()
	client := resty.New()
	client.SetTimeout(PERIPHERAL_API_GLOBAL_TIMEOUT)
	resp, err := client.R().
		SetBody(payload).
		Post(PERIPHERAL_API_BASEURL + PERIPHERAL_API_GENERATE_SQL_PATH)
	if resp.StatusCode() != http.StatusOK || err != nil {
		return nil, errors.New("failed to generate SQL")
	}
	res := &GenerateSQLFeedback{}
	json.Unmarshal(resp.Body(), res)
	res.Payload = req.GetActionInString() + res.Payload + ";"
	return res, nil
}

func Echo(req *EchoPeripheralRequest) (*EchoFeedback, error) {
	payload := req.Export()
	client := resty.New()
	client.SetTimeout(PERIPHERAL_API_GLOBAL_TIMEOUT)
	resp, err := client.R().
		SetBody(payload).
		Post(PERIPHERAL_API_BASEURL + PERIPHERAL_API_ECHO_PATH)
	if resp.StatusCode() != http.StatusOK || err != nil {
		return nil, errors.New("failed to generate SQL")
	}
	res := &GeneralFeedback{}
	json.Unmarshal(resp.Body(), res)
	echoFeedback := &EchoFeedback{}
	json.Unmarshal([]byte(res.Payload), echoFeedback)
	if !echoFeedback.Avaliable() {
		return nil, errors.New("unavaliable echo request.")
	}
	return echoFeedback, nil
}
