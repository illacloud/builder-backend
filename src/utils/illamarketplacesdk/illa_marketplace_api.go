package illamarketplacesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"
)

const (
	BASEURL = "http://127.0.0.1:9001/api/v1"
	// api route part
	FORK_COUNTER_API                             = "/products/%s/%d/fork"
	RUN_COUNTER_API                              = "/products/%s/%d/run"
	DELETE_PRODUCT                               = "/products/%s/%d"
	UPDATE_PRODUCTS                              = "/products/%s/%d"
	DELETE_TEAM_ALL_PRODUCTS                     = "/products/byTeam/%d"
	PUBLISH_AI_AGENT_TO_MARKETPLACE_INTERNAL_API = "/products/byTeam/%d/aiAgents/%d/byUser/%d"
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

func NewIllaMarketplaceRestAPI() *IllaMarketplaceRestAPI {
	requestTokenValidator := tokenvalidator.NewRequestTokenValidator()
	return &IllaMarketplaceRestAPI{
		Config:    config.GetInstance(),
		Validator: requestTokenValidator,
	}
}

func (r *IllaMarketplaceRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaMarketplaceRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaMarketplaceRestAPI) ForkCounter(productType string, productID int) error {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil
	}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(fmt.Sprintf("%d", productID))).
		Post(r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(FORK_COUNTER_API, productType, productID))
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.ForkCounter()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return err
		}
		return errors.New(resp.String())
	}
	return nil
}

func (r *IllaMarketplaceRestAPI) DeleteTeamAllProducts(teamID int) error {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil
	}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(fmt.Sprintf("%d", teamID))).
		Delete(r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(DELETE_TEAM_ALL_PRODUCTS, teamID))
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.DeleteTeamAllProducts()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return err
		}
		return errors.New(resp.String())
	}
	return nil
}

func (r *IllaMarketplaceRestAPI) DeleteProduct(productType string, productID int) error {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil
	}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(fmt.Sprintf("%d", productID))).
		Delete(r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(DELETE_PRODUCT, productType, productID))
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.DeleteProduct()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return err
		}
		return errors.New(resp.String())
	}
	return nil
}

func (r *IllaMarketplaceRestAPI) UpdateProduct(productType string, productID int, product interface{}) error {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil
	}
	b, err := json.Marshal(product)
	if err != nil {
		return err
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(fmt.Sprintf("%d", productID), string(b))).
		SetBody(product).
		Put(r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(UPDATE_PRODUCTS, productType, productID))
	log.Printf("[IllaMarketplaceRestAPI.UpdateProduct()]  response: %+v, err: %+v", resp, err)
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.UpdateProduct()]  response: %+v, err: %+v", resp, err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		if err != nil {
			return err
		}
		return errors.New(resp.String())
	}
	return nil
}

func (r *IllaMarketplaceRestAPI) PublishAIAgentToMarketplace(aiAgentID int, teamID int, userID int) error {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil
	}
	client := resty.New()
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	uri := r.Config.IllaMarketplaceInternalRestAPI + fmt.Sprintf(PUBLISH_AI_AGENT_TO_MARKETPLACE_INTERNAL_API, teamID, aiAgentID, userID)
	resp, errInPatch := client.R().
		SetHeader("Request-Token", tokenValidator.GenerateValidateToken(strconv.Itoa(teamID), strconv.Itoa(aiAgentID), strconv.Itoa(userID))).
		Post(uri)
	if r.Debug {
		log.Printf("[IllaMarketplaceRestAPI.PublishAIAgentToMarketplace()]  uri: %+v \n", uri)
		log.Printf("[IllaMarketplaceRestAPI.PublishAIAgentToMarketplace()]  response: %+v, err: %+v \n", resp, errInPatch)
		log.Printf("[IllaMarketplaceRestAPI.PublishAIAgentToMarketplace()]  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInPatch != nil {
		return errInPatch
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return errors.New(resp.String())
	}

	return nil
}
