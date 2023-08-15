package illaresourcemanagerbackendsdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/internal/util/config"
)

const (
	BASEURL = "http://127.0.0.1:8008/api/v1"
	// api route part
	RUN_AI_AGENT_API = "/api/v1/teams/%s/aiAgent/%s/run"
)

const (
	PRODUCT_TYPE_AIAGENTS = "aiAgents"
	PRODUCT_TYPE_APPS     = "apps"
	PRODUCT_TYPE_HUBS     = "hubs"
)

type IllaResourceManagerRestAPI struct {
	Config *config.Config
	Debug  bool `json:"-"`
}

func NewIllaResourceManagerRestAPI() (*IllaResourceManagerRestAPI, error) {
	return &IllaResourceManagerRestAPI{
		Config: config.GetInstance(),
	}, nil
}

func (r *IllaResourceManagerRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaResourceManagerRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaResourceManagerRestAPI) RunAiAgent(req map[string]interface{}) (*RunAIAgentResult, error) {
	reqInstance, errInNewReq := NewRunAIAgentRequest(req)
	if errInNewReq != nil {
		return nil, errInNewReq
	}
	client := resty.New()
	uri := r.Config.GetIllaResourceManagerRestAPI() + fmt.Sprintf(RUN_AI_AGENT_API, reqInstance.ExportTeamID(), reqInstance.ExportAIAgentID())
	resp, errInPost := client.R().
		SetHeader("Authorization", reqInstance.ExportAuthorization()).
		SetBody(reqInstance).
		Post(uri)
	if r.Debug {
		log.Printf("[IllaResourceManagerRestAPI.RunAiAgent()]  uri: %+v \n", uri)
		log.Printf("[IllaResourceManagerRestAPI.RunAiAgent()]  response: %+v, err: %+v \n", resp, errInPost)
		log.Printf("[IllaResourceManagerRestAPI.RunAiAgent()]  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInPost != nil {
		return nil, errInPost
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, errors.New(resp.String())
	}

	runAIAgentResult := NewRunAIAgentResult()
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &runAIAgentResult)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	return runAIAgentResult, nil
}
