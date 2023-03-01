package illacloudservice

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/illa-builder-backend/src/utils/config"
)

type GenerateSQLFeedback struct {
	Payload string `json:"payload"`
}

func GenerateSQL(m *GenerateSQLPeripheralRequest, req *GenerateSQLRequest) (*GenerateSQLFeedback, error) {
	conf := config.GetInstance()
	// run
	payload := m.Export()
	client := resty.New()
	resp, err := client.R().
		SetBody(payload).
		Post(conf.GetGenerateSQLAPI())
	if resp.StatusCode() != http.StatusOK || err != nil {
		return nil, errors.New("failed to generate SQL")
	}
	res := &GenerateSQLFeedback{}
	json.Unmarshal(resp.Body(), res)
	res.Payload = req.GetActionInString() + res.Payload + ";"
	return res, nil
}
