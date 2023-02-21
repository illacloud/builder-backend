package repository

import (
	"encoding/json"
	"errors"
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

func GenerateSQL(m *GenerateSQLPeripheralRequest, req *GenerateSQLRequest) (*GenerateSQLFeedback, error) {
	payload := m.Export()
	client := resty.New()
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
