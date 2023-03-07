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

package appwrite

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/pkg/plugins/common"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	Resource Resource
	Action   Action
}

func (a *Connector) ValidateResourceOptions(resourceOpts map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOpts, &a.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate appwrite options
	validate := validator.New()
	if err := validate.Struct(a.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (a *Connector) ValidateActionOptions(actionOpts map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOpts, &a.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate appwrite options
	validate := validator.New()
	if err := validate.Struct(a.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (a *Connector) TestConnection(resourceOpts map[string]interface{}) (common.ConnectionResult, error) {
	// get appwrite database client
	db, err := a.getClientWithOpts(resourceOpts)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test appwrite client
	pong, err := db.Get(a.Resource.DatabaseID)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	if pong.StatusCode != 200 {
		return common.ConnectionResult{Success: false}, errors.New(pong.Result)
	}

	return common.ConnectionResult{Success: true}, nil
}

func (a *Connector) GetMetaInfo(resourceOpts map[string]interface{}) (common.MetaInfoResult, error) {
	// get appwrite database client
	db, err := a.getClientWithOpts(resourceOpts)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get collections
	var EmptyArray = []interface{}{}
	colls, err := db.ListCollections(a.Resource.DatabaseID, EmptyArray, "")
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	if colls.StatusCode != 200 {
		return common.MetaInfoResult{Success: false}, errors.New(colls.Result)
	}

	// get output
	var jsonResp map[string]interface{}
	if err := json.Unmarshal([]byte(colls.Result), &jsonResp); err != nil {
		return common.MetaInfoResult{Success: false}, errors.New("invalid response")
	}
	collections, ok := jsonResp["collections"]
	if !ok {
		return common.MetaInfoResult{Success: false}, errors.New("invalid response")
	}
	collectionsArray, ok := collections.([]map[string]interface{})
	if !ok {
		return common.MetaInfoResult{Success: false}, errors.New("invalid response")
	}

	res := make([]map[string]interface{}, len(collectionsArray))
	for i, v := range collectionsArray {
		res[i] = map[string]interface{}{"id": v["$id"]}
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"collections": res},
	}, nil
}

func (a *Connector) Run(resourceOpts map[string]interface{}, actionOpts map[string]interface{}) (common.RuntimeResult, error) {
	// get appwrite database client
	db, err := a.getClientWithOpts(resourceOpts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format appwrite action
	if err := mapstructure.Decode(actionOpts, &a.Action); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var result common.RuntimeResult
	executor := ActionExecutor{client: db, action: a.Action, database: a.Resource.DatabaseID}
	switch a.Action.Method {
	case LIST_METHOD:
		result, err = executor.ListDocs()
	case CREATE_METHOD:
		result, err = executor.CreateDoc()
	case GET_METHOD:
		result, err = executor.GetDoc()
	case UPDATE_METHOD:
		result, err = executor.UpdateDoc()
	case DELETE_METHOD:
		result, err = executor.DeleteDoc()
	}
	return result, err
}
