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

	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

const (
	DB_QUERY_OP  = "query"
	DB_SET_OP    = "set"
	DB_UPDATE_OP = "update"
	DB_APPEND_OP = "append"
)

type DBOperationRunner struct {
	client    *firebase.App
	operation string
	options   map[string]interface{}
}

type DBOptions struct {
	Ref    string
	Object map[string]interface{}
}

func (d *DBOperationRunner) run() (common.RuntimeResult, error) {
	var result common.RuntimeResult
	var err error
	switch d.operation {
	case DB_QUERY_OP:
		result, err = d.query()
	case DB_SET_OP:
		result, err = d.set()
	case DB_UPDATE_OP:
		result, err = d.update()
	case DB_APPEND_OP:
		result, err = d.append()
	default:
		result.Success = false
		err = errors.New("unsupported operation")
	}
	return result, err
}

func (d *DBOperationRunner) query() (common.RuntimeResult, error) {
	var queryOptions DBOptions
	if err := mapstructure.Decode(d.options, &queryOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase realtime database `query` action options
	validate := validator.New()
	if err := validate.Struct(queryOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build query action
	ctx := context.TODO()
	client, err := d.client.Database(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	ref := client.NewRef(queryOptions.Ref)

	var res interface{}
	if err := ref.Get(ctx, &res); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": res}}}, nil
}

func (d *DBOperationRunner) set() (common.RuntimeResult, error) {
	var setOptions DBOptions
	if err := mapstructure.Decode(d.options, &setOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase realtime database `set` action options
	validate := validator.New()
	if err := validate.Struct(setOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build set action
	ctx := context.TODO()
	client, err := d.client.Database(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	ref := client.NewRef(setOptions.Ref)

	if err := ref.Set(ctx, &setOptions.Object); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true}, nil
}

func (d *DBOperationRunner) update() (common.RuntimeResult, error) {
	var updateOptions DBOptions
	if err := mapstructure.Decode(d.options, &updateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase realtime database `update` action options
	validate := validator.New()
	if err := validate.Struct(updateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build update action
	ctx := context.TODO()
	client, err := d.client.Database(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	ref := client.NewRef(updateOptions.Ref)

	if err := ref.Update(ctx, updateOptions.Object); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true}, nil
}

func (d *DBOperationRunner) append() (common.RuntimeResult, error) {
	var appendOptions DBOptions
	if err := mapstructure.Decode(d.options, &appendOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase realtime database `append` action options
	validate := validator.New()
	if err := validate.Struct(appendOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build append action
	ctx := context.TODO()
	client, err := d.client.Database(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	ref := client.NewRef(appendOptions.Ref)
	newRef, err := ref.Push(ctx, nil)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if err := newRef.Set(ctx, &appendOptions.Object); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true}, nil
}
