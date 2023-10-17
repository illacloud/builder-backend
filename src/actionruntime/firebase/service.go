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

package firebase

import (
	"context"
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (f *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &f.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate firebase options
	validate := validator.New()
	if err := validate.Struct(f.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (f *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &f.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate firebase options
	validate := validator.New()
	if err := validate.Struct(f.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (f *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get firebase app
	app, err := f.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test connection
	ctx := context.TODO()
	firestoreClient, errF := app.Firestore(ctx)
	_, errA := app.Auth(ctx)
	_, errD := app.Database(ctx)
	if errF != nil && errA != nil && errD != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer firestoreClient.Close()

	return common.ConnectionResult{Success: true}, nil
}

// GetMetaInfo get the collections in firestore
func (f *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get firebase app
	app, err := f.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get firestore client
	ctx := context.TODO()
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer firestoreClient.Close()

	// get collections
	collsIter := firestoreClient.Collections(ctx)
	colls, err := collsIter.GetAll()
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	res := make([]string, 0, len(colls))
	for _, coll := range colls {
		res = append(res, coll.ID)
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"collections": res},
	}, nil
}

func (f *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get firebase app
	app, err := f.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format firebase operation
	if err := mapstructure.Decode(actionOptions, &f.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var result common.RuntimeResult
	switch f.ActionOpts.Service {
	case AUTH_SERVICE:
		operationRunner := &AuthOperationRunner{client: app, operation: f.ActionOpts.Operation, options: f.ActionOpts.Options}
		result, err = operationRunner.run()
	case DATABASE_SERVICE:
		operationRunner := &DBOperationRunner{client: app, operation: f.ActionOpts.Operation, options: f.ActionOpts.Options}
		result, err = operationRunner.run()
	case FIRESTORE_SERVICE:
		operationRunner := &FirestoreOperationRunner{client: app, operation: f.ActionOpts.Operation, options: f.ActionOpts.Options}
		result, err = operationRunner.run()
	default:
		result.Success = false
		err = errors.New("unsupported operation")
	}

	return result, err
}
