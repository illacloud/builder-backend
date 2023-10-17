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

package mongodb

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type Connector struct {
	Resource Options
	Action   Query
}

func (m *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format mongodb simple options
	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate simple options
	validate := validator.New()
	if err := validate.Struct(m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate specific options
	if m.Resource.ConfigType == GUI_OPTIONS {
		var mOptions GUIOptions
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
		if err := validate.Struct(mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
	} else if m.Resource.ConfigType == URI_OPTIONS {
		var mOptions URIOptions
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
		if err := validate.Struct(mOptions); err != nil {
			return common.ValidateResult{Valid: false}, err
		}
	}

	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format mongodb query options
	if err := mapstructure.Decode(actionOptions, &m.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate
	validate := validator.New()
	if err := validate.Struct(m.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get mongodb connection
	client, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer client.Disconnect(context.Background())

	// test mongodb connection
	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	return common.ConnectionResult{Success: true}, nil
}

func (m *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {

	return common.MetaInfoResult{
		Success: true,
		Schema:  nil,
	}, nil
}

func (m *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get mongodb connection
	client, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	defer client.Disconnect(context.Background())

	db := ""
	if m.Resource.ConfigType == GUI_OPTIONS {
		var mOptions GUIOptions
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		db = mOptions.DatabaseName
	} else if m.Resource.ConfigType == URI_OPTIONS {
		mOptions := URIOptions{}
		if err := mapstructure.Decode(m.Resource.ConfigContent, &mOptions); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		uri := mOptions.URI
		matchedStrs, err := connstring.Parse(uri)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		db = matchedStrs.Database
	}
	if db == "" {
		db = "test"
	}

	var result common.RuntimeResult
	queryRunner := QueryRunner{client: client, query: m.Action, db: db}
	switch m.Action.ActionType {
	case "aggregate":
		result, err = queryRunner.aggregate()
	case "bulkWrite":
		result, err = queryRunner.bulkWrite()
	case "count":
		result, err = queryRunner.count()
	case "deleteMany":
		result, err = queryRunner.deleteMany()
	case "deleteOne":
		result, err = queryRunner.deleteOne()
	case "distinct":
		result, err = queryRunner.distinct()
	case "find":
		result, err = queryRunner.find()
	case "findOne":
		result, err = queryRunner.findOne()
	case "findOneAndUpdate":
		result, err = queryRunner.findOneAndUpdate()
	case "insertOne":
		result, err = queryRunner.insertOne()
	case "insertMany":
		result, err = queryRunner.insertMany()
	case "listCollections":
		result, err = queryRunner.listCollections()
	case "updateMany":
		result, err = queryRunner.updateMany()
	case "updateOne":
		result, err = queryRunner.updateOne()
	case "command":
		result, err = queryRunner.command()
	}

	return result, err
}
