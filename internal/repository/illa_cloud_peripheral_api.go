package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	PERIPHERAL_API_BASEURL           = "https://email.illasoft.com/v1/"
	PERIPHERAL_API_GENERATE_SQL_PATH = "generateSQL"
)

type GenerateSQLFeedback struct {
	Payload string `json:"payload"`
}

func (f *GenerateSQLFeedback) AddSuffix() {
	f.Payload = "SELECT " + f.Payload
}

func GenerateSQL(m *GenerateSQLPeripheralRequest, req *GenerateSQLRequest) (*GenerateSQLFeedback, error) {
	fmt.Printf("%v\n", m)
	payload := m.Export()
	client := resty.New()
	resp, err := client.R().
		SetBody(payload).
		Post(PERIPHERAL_API_BASEURL + PERIPHERAL_API_GENERATE_SQL_PATH)
	if resp.StatusCode() != http.StatusOK || err != nil {
		fmt.Printf("response: %+v, err: %+v", resp, err)
		return nil, errors.New("failed to generate SQL")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	res := &GenerateSQLFeedback{}
	json.Unmarshal(resp.Body(), res)
	res.Payload = req.GetActionInString() + res.Payload + ";"
	return res, nil
}
