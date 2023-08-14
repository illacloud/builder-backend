package illaresourcemanagerbackendsdk

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/illa-resource-manager-backend/src/utils/config"
	"github.com/illacloud/illa-resource-manager-backend/src/utils/tokenvalidator"
)

const (
	BASEURL = "http://127.0.0.1:8008/api/v1"
	// api route part
	RUN_AI_AGENT_API = "/api/v1/products/%s/%d/fork"
)

const (
	PRODUCT_TYPE_AIAGENTS = "aiAgents"
	PRODUCT_TYPE_APPS     = "apps"
	PRODUCT_TYPE_HUBS     = "hubs"
)

type IllaMarketplaceRestAPI struct {
	Config    *config.Config
	Validator *tokenvalidator.RequestTokenValidator
	Debug     bool `json:"-"`
}

func NewIllaMarketplaceRestAPI() (*IllaMarketplaceRestAPI, error) {
	requestTokenValidator := tokenvalidator.NewRequestTokenValidator()
	return &IllaMarketplaceRestAPI{
		Config:    config.GetInstance(),
		Validator: requestTokenValidator,
	}, nil
}

func (r *IllaMarketplaceRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaMarketplaceRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaMarketplaceRestAPI) RunAiAgentByID(aiAgentID string, productID int) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(string(productID))).
		Post(r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(FORK_COUNTER_API, productType, productID))
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.ForkCounter()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK || resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return err
		}
		return errors.New(resp.String())
	}
	return nil
}
