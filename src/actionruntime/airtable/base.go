// Copyright 2023 Illa Soft, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package airtable

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
)

func (a *Connector) ListRecords() (common.RuntimeResult, error) {
	// format `list` method config
	var listConfig ListConfig
	if err := mapstructure.Decode(a.Action.Config, &listConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `list` method config
	if listConfig.PageSize <= 0 {
		listConfig.PageSize = 100
	}
	if listConfig.CellFormat == STRING_CELL_FORMAT {
		if listConfig.TimeZone == "" && listConfig.UserLocale == "" {
			return common.RuntimeResult{Success: false}, errors.New("missing timezone or user locale")
		}
	} else {
		listConfig.CellFormat = JSON_CELL_FORMAT
	}

	// build `List Records` request body
	listReqBody := make(map[string]interface{})
	if len(listConfig.Fields) > 0 {
		listReqBody["fields"] = listConfig.Fields
	}
	if listConfig.FilterByFormula != "" {
		listReqBody["filterByFormula"] = listConfig.FilterByFormula
	}
	if listConfig.MaxRecords > -1 {
		listReqBody["maxRecords"] = listConfig.MaxRecords
	}
	if listConfig.PageSize > -1 {
		listReqBody["pageSize"] = listConfig.PageSize
	}
	sortObjs := make([]map[string]string, 0)
	for _, sortObj := range listConfig.Sort {
		if sortObj.Field != "" {
			sortMap := map[string]string{
				"field":     sortObj.Field,
				"direction": sortObj.Direction,
			}
			sortObjs = append(sortObjs, sortMap)
		}
	}
	listReqBody["sort"] = sortObjs
	if listConfig.View != "" {
		listReqBody["view"] = listConfig.View
	}
	if listConfig.CellFormat != "" {
		listReqBody["cellFormat"] = listConfig.CellFormat
	}
	if listConfig.TimeZone != "" {
		listReqBody["timeZone"] = listConfig.TimeZone
	}
	if listConfig.UserLocale != "" {
		listReqBody["userLocale"] = listConfig.UserLocale
	}
	if listConfig.Offset != "" {
		listReqBody["offset"] = listConfig.Offset
	}

	// call `List Records` method
	restyClient := resty.New()
	listReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		listReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		listReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := listReq.SetBody(listReqBody).
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Post(AIRTABLE_API + "/listRecords")

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) GetRecord() (common.RuntimeResult, error) {
	// format `get` method config
	var getConfig GetConfig
	if err := mapstructure.Decode(a.Action.Config, &getConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `get` method config
	validate := validator.New()
	if err := validate.Struct(getConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// call `Get Record` method
	restyClient := resty.New()
	getReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		getReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		getReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := getReq.
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Get(AIRTABLE_API + "/" + getConfig.RecordID)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) CreateRecords() (common.RuntimeResult, error) {
	// format `create` method config
	var createConfig CreateConfig
	if err := mapstructure.Decode(a.Action.Config, &createConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `create` method config
	validate := validator.New()
	if err := validate.Struct(createConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build `Create Records` request body
	createReqBody := make(map[string]interface{}, 1)
	if len(createConfig.Records) == 1 {
		createReqBody = createConfig.Records[0]
	} else if len(createConfig.Records) > 1 {
		records := make([]map[string]interface{}, 0)
		for _, record := range createConfig.Records {
			records = append(records, record)
		}
		createReqBody["records"] = records
	}

	// call `Create Records` method
	restyClient := resty.New()
	createReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		createReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		createReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := createReq.SetBody(createReqBody).
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Post(AIRTABLE_API)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) UpdateMultipleRecords() (common.RuntimeResult, error) {
	// format `bulkUpdate` method config
	var bulkUpdateConfig BulkUpdateConfig
	if err := mapstructure.Decode(a.Action.Config, &bulkUpdateConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `bulkUpdate` method config
	validate := validator.New()
	if err := validate.Struct(bulkUpdateConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build `Update Multiple Records` request body
	bulkUpdateReqBody := make(map[string]interface{}, 1)
	bulkUpdateReqBody["records"] = bulkUpdateConfig.Records

	// call `Update Multiple Records` method
	restyClient := resty.New()
	bulkUpdateReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		bulkUpdateReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		bulkUpdateReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := bulkUpdateReq.SetBody(bulkUpdateReqBody).
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Patch(AIRTABLE_API)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) UpdateRecord() (common.RuntimeResult, error) {
	// format `update` method config
	var updateConfig UpdateConfig
	if err := mapstructure.Decode(a.Action.Config, &updateConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `update` method config
	validate := validator.New()
	if err := validate.Struct(updateConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build `Update Multiple Records` request body
	updateReqBody := make(map[string]interface{})
	updateReqBody = updateConfig.Record

	// call `Update Multiple Records` method
	restyClient := resty.New()
	updateReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		updateReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		updateReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := updateReq.SetBody(updateReqBody).
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Patch(AIRTABLE_API + "/" + updateConfig.RecordID)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) DeleteMultipleRecords() (common.RuntimeResult, error) {
	// format `bulkDelete` method config
	var bulkDeleteConfig BulkDeleteConfig
	if err := mapstructure.Decode(a.Action.Config, &bulkDeleteConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `bulkDelete` method config
	validate := validator.New()
	if err := validate.Struct(bulkDeleteConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// call `Delete Multiple Records` method
	deleteIds := make([]string, 0, len(bulkDeleteConfig.RecordIDs))
	for _, ids := range bulkDeleteConfig.RecordIDs {
		deleteIds = append(deleteIds, fmt.Sprintf("records=%s", ids))
	}
	deleteIdsQueryParams := "?" + strings.Join(deleteIds, "&")
	restyClient := resty.New()
	deleteReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		deleteReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		deleteReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := deleteReq.
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Delete(AIRTABLE_API + "/" + deleteIdsQueryParams)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func (a *Connector) DeleteRecord() (common.RuntimeResult, error) {
	// format `delete` method config
	var deleteConfig DeleteConfig
	if err := mapstructure.Decode(a.Action.Config, &deleteConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate `delete` method config
	validate := validator.New()
	if err := validate.Struct(deleteConfig); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// call `Delete Record` method
	restyClient := resty.New()
	deleteReq := restyClient.R().SetHeader("Content-Type", "application/json")
	if a.Resource.AuthenticationType == API_KEY_AUTHENTICATION {
		deleteReq.SetAuthToken(a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION])
	} else if a.Resource.AuthenticationType == PERSONAL_TOKEN_AUTHENTICATION {
		deleteReq.SetAuthToken(a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION])
	}
	resp, errRun := deleteReq.
		SetPathParams(map[string]string{
			"baseId":    a.Action.BaseConfig.BaseID,
			"tableName": a.Action.BaseConfig.TableName,
		}).
		Delete(AIRTABLE_API + "/" + deleteConfig.RecordID)

	// handle response
	if resp.StatusCode() != http.StatusOK {
		if errRun != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Request to Airtable failed: " + errRun.Error()}}}, nil
		}
		respMap, errParseError := parseAirtableResponse(resp.String())
		if errParseError != nil {
			return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseError.Error()}}}, nil
		}
		respMap["status"] = resp.Status()
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
	}
	respMap, errParseResp := parseAirtableResponse(resp.String())
	if errParseResp != nil {
		return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"message": "Parse Airtable response error: " + errParseResp.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{respMap}}, nil
}

func parseAirtableResponse(responseString string) (map[string]interface{}, error) {
	var respMap map[string]interface{}
	if err := json.Unmarshal([]byte(responseString), &respMap); err != nil {
		return nil, err
	}
	return respMap, nil
}
