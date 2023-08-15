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

type IllaMarketplaceRestAPI struct {
	Config *config.Config
	Debug  bool `json:"-"`
}

func NewIllaMarketplaceRestAPI() (*IllaMarketplaceRestAPI, error) {
	return &IllaMarketplaceRestAPI{
		Config: config.GetInstance(),
	}, nil
}

func (r *IllaMarketplaceRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaMarketplaceRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaMarketplaceRestAPI) RunAiAgentByID(teamID string, aiAgentID string, authorizationToken string, req map[string]interface{}) (*RunAIAgentResult, error) {
	reqInstance, errInNewReq := NewRunAIAgentRequest(req)
	if errInNewReq != nil {
		return nil, errInNewReq
	}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", authorizationToken).
		SetBody(reqInstance).
		Post(r.Config.GetIllaResourceManagerRestAPI() + fmt.Sprintf(RUN_AI_AGENT_API, teamID, aiAgentID))
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.ForkCounter()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK || resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return nil, err
		}
		return nil, errors.New(resp.String())
	}

	runAIAgentResult := NewRunAIAgentResult()
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &runAIAgentResult)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	return runAIAgentResult, nil
}
