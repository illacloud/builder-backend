// Copyright 2022 The ILLA Authors.
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

package elasticsearch

import (
	"context"
	"errors"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (e *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &e.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate elasticsearch options
	validate := validator.New()
	if err := validate.Struct(e.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (e *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &e.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate elasticsearch options
	validate := validator.New()
	if err := validate.Struct(e.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (e *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get es connection
	esClient, err := e.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test es connection
	pingReq := esapi.PingRequest{
		Pretty: true,
		Human:  true,
	}
	pingRes, err := pingReq.Do(context.TODO(), esClient)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer pingRes.Body.Close()

	return common.ConnectionResult{Success: true}, nil
}

func (e *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{
		Success: true,
		Schema:  nil,
	}, nil
}

func (e *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get mysql connection
	esClient, err := e.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get elasticsearch connection")
	}

	// format es operation
	if err := mapstructure.Decode(actionOptions, &e.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var result common.RuntimeResult
	operationRunner := OperationRunner{client: esClient, operation: e.ActionOpts}
	switch e.ActionOpts.Operation {
	case SEARCH_OPERATION:
		result, err = operationRunner.search()
	case INSERT_OPERATION:
		result, err = operationRunner.insert()
	case GET_OPERATION:
		result, err = operationRunner.get()
	case UPDATE_OPERATION:
		result, err = operationRunner.update()
	case DELETE_OPERATION:
		result, err = operationRunner.delete()
	default:
		result.Success = false
		err = errors.New("unsupported elasticsearch operation")
	}

	return result, err
}
