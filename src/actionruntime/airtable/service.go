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
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	Resource Resource
	Action   Action
}

func (a *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &a.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Airtable options
	validate := validator.New()
	if err := validate.Struct(a.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	switch a.Resource.AuthenticationType {
	case API_KEY_AUTHENTICATION:
		if _, ok := a.Resource.AuthenticationConfig[API_KEY_AUTHENTICATION]; !ok {
			return common.ValidateResult{Valid: false}, errors.New("missing API key")
		}
	case PERSONAL_TOKEN_AUTHENTICATION:
		if _, ok := a.Resource.AuthenticationConfig[TOKEN_AUTHENTICATION]; !ok {
			return common.ValidateResult{Valid: false}, errors.New("missing Personal Access Token")
		}
	default:
		return common.ValidateResult{Valid: false}, errors.New("invalid parameters")
	}

	return common.ValidateResult{Valid: true}, nil
}

func (a *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &a.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Airtable options
	validate := validator.New()
	if err := validate.Struct(a.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (a *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: true}, nil
}

func (a *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: true}, nil
}

func (a *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &a.Resource); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &a.Action); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// run action based on action method
	var result common.RuntimeResult
	var errRun error
	switch a.Action.Method {
	case LIST_METHOD:
		result, errRun = a.ListRecords()
	case GET_METHOD:
		result, errRun = a.GetRecord()
	case CREATE_METHOD:
		result, errRun = a.CreateRecords()
	case BULKUPDATE_METHOD:
		result, errRun = a.UpdateMultipleRecords()
	case UPDATE_METHOD:
		result, errRun = a.UpdateRecord()
	case BULKDELETE_METHOD:
		result, errRun = a.DeleteMultipleRecords()
	case DELETE_METHOD:
		result, errRun = a.DeleteRecord()
	default:
		errRun = errors.New("invalid action method")
	}

	return result, errRun
}
