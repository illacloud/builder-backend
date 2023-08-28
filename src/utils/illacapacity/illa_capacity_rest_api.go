package illacapacity

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
	START_TRANSACTION_API             = "/transactionControl/teams/%d/new"
	COMMIT_TRANSACTION_API            = "/transactionControl/%s/commit"
	CANCEL_TRANSACTION_API            = "/transactionControl/%s/cancel"
	UPDATE_CAPACITY_WITH_NEGATIVE_API = "/transactionSerialControl/capacitiesWithNegativeValue"
	TEST_CAPACITY                     = "/transactionSerialControl/capacities/test"
)

type IllaCapacityRestAPI struct {
	Config    *config.Config
	Validator *tokenvalidator.RequestTokenValidator
}

func NewIllaCapacityRestAPI() (*IllaCapacityRestAPI, error) {
	requestTokenValidator := tokenvalidator.NewRequestTokenValidator()
	return &IllaCapacityRestAPI{
		Config:    config.GetInstance(),
		Validator: requestTokenValidator,
	}, nil
}

func (r *IllaCapacityRestAPI) StartTransaction(teamID int) (string, error) {
	teamIDInString := strconv.Itoa(teamID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(teamIDInString)).
		Get(r.Config.IllaCapacityInternalRestAPI + fmt.Sprintf(START_TRANSACTION_API, teamID))
	log.Printf("[IllaCapacityRestAPI.StartTransaction()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa capacity failed")
		}
		return "", errors.New("start transaction failed")
	}
	// decode target team drive record id
	transactionResponse := NewStartTransactionResponse()
	errInUnmarshal := json.Unmarshal([]byte(resp.String()), transactionResponse)
	if errInUnmarshal != nil {
		return "", errInUnmarshal
	}
	return transactionResponse.ExportTXUIDInString(), nil
}

func (r *IllaCapacityRestAPI) CommitTransaction(transactionUID string) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(transactionUID)).
		Patch(r.Config.IllaCapacityInternalRestAPI + fmt.Sprintf(COMMIT_TRANSACTION_API, transactionUID))
	log.Printf("[IllaCapacityRestAPI.CommitTransaction()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return errors.New("request illa capacity failed")
		}
		return errors.New("commit transaction failed")
	}
	return nil
}

func (r *IllaCapacityRestAPI) CancelTransaction(transactionUID string) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(transactionUID)).
		Patch(r.Config.IllaCapacityInternalRestAPI + fmt.Sprintf(CANCEL_TRANSACTION_API, transactionUID))
	log.Printf("[IllaCapacityRestAPI.CancelTransaction()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return errors.New("request illa capacity failed")
		}
		return errors.New("cancel transaction failed")
	}
	return nil
}

func (r *IllaCapacityRestAPI) UpdateCapacityWithNegativeValue(updateCapacityRequest *UpdateCapacityRequest) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateTokenBySliceParam(updateCapacityRequest.ExportForRequestTokenValidator())).
		SetBody(updateCapacityRequest).
		Post(r.Config.IllaCapacityInternalRestAPI + UPDATE_CAPACITY_WITH_NEGATIVE_API)
	log.Printf("[IllaCapacityRestAPI.StartTransaction()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return errors.New("request illa capacity failed")
		}
		return errors.New("update capacity failed")
	}
	return nil
}

func (r *IllaCapacityRestAPI) TestCapacity(updateCapacityRequest *UpdateCapacityRequest) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateTokenBySliceParam(updateCapacityRequest.ExportForRequestTokenValidator())).
		SetBody(updateCapacityRequest).
		Post(r.Config.IllaCapacityInternalRestAPI + TEST_CAPACITY)
	log.Printf("[IllaCapacityRestAPI.TestCapacity()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return errors.New("request illa capacity failed")
		}
		return errors.New("update capacity failed")
	}
	return nil
}
