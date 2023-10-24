package illacloudperipheralapisdk

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
)

const (
	PERIPHERAL_API_GENERATE_SQL_PATH = "generateSQL"
)

type IllaCloudPeriphearalAPI struct {
	Config *config.Config
}

func NewIllaCloudPeriphearalAPI() *IllaCloudPeriphearalAPI {
	return &IllaCloudPeriphearalAPI{
		Config: config.GetInstance(),
	}
}

func (i *IllaCloudPeriphearalAPI) GenerateSQL(m *GenerateSQLPeripheralRequest) (*GenerateSQLFeedback, error) {
	payload := m.Export()
	client := resty.New()
	resp, err := client.R().
		SetBody(payload).
		Post(i.Config.GetIllaPeripheralAPI() + PERIPHERAL_API_GENERATE_SQL_PATH)
	if resp.StatusCode() != http.StatusOK || err != nil {
		return nil, errors.New("failed to generate SQL")
	}
	res := &GenerateSQLFeedback{}
	json.Unmarshal(resp.Body(), res)
	res.Payload = m.SQLAction + " " + res.Payload + ";"
	return res, nil
}
