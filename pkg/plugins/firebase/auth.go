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

	"github.com/illa-family/builder-backend/pkg/plugins/common"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/iterator"
)

const (
	AUTH_UID_OP    = "uid"
	AUTH_EMAIL_OP  = "email"
	AUTH_PHOME_OP  = "phone"
	AUTH_CREATE_OP = "create"
	AUTH_UPDATE_OP = "update"
	AUTH_DELETE_OP = "delete"
	AUTH_LIST_OP   = "list"
)

type AuthOperationRunner struct {
	client    *firebase.App
	operation string
	options   map[string]interface{}
}

type AuthBaseOptions struct {
	filter string `validate:"required"`
}

type AuthCreateOptions struct {
	object UserObject `validate:"required"`
}

type AuthUpdateOptions struct {
	uid    string     `validate:"required"`
	object UserObject `validate:"required"`
}

type AuthListOptions struct {
	number int
	token  string
}

type UserObject struct {
	uid           string
	email         string
	emailVerified bool
	phoneNumber   string
	password      string
	displayName   string
	photoURL      string
	disabled      bool
}

func (a *AuthOperationRunner) run() (common.RuntimeResult, error) {
	var result common.RuntimeResult
	var err error
	switch a.operation {
	case AUTH_UID_OP, AUTH_EMAIL_OP, AUTH_PHOME_OP:
		result, err = a.query()
	case AUTH_CREATE_OP:
		result, err = a.create()
	case AUTH_UPDATE_OP:
		result, err = a.update()
	case AUTH_DELETE_OP:
		result, err = a.delete()
	case AUTH_LIST_OP:
		result, err = a.list()
	default:
		result.Success = false
		err = errors.New("unsupported operation")
	}
	return result, err
}

func (a *AuthOperationRunner) query() (common.RuntimeResult, error) {
	var queryOptions AuthBaseOptions
	if err := mapstructure.Decode(a.options, &queryOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase User Management `query` action options
	validate := validator.New()
	if err := validate.Struct(queryOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build query action
	ctx := context.TODO()
	client, err := a.client.Auth(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// run different of operations
	var user *auth.UserRecord
	switch a.operation {
	case AUTH_UID_OP:
		user, err = client.GetUser(ctx, queryOptions.filter)
	case AUTH_EMAIL_OP:
		user, err = client.GetUserByEmail(ctx, queryOptions.filter)
	case AUTH_PHOME_OP:
		user, err = client.GetUserByPhoneNumber(ctx, queryOptions.filter)
	default:
		err = errors.New("unsupported operation")
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"user": user}}}, err
}

func (a *AuthOperationRunner) create() (common.RuntimeResult, error) {
	var createOptions AuthCreateOptions
	if err := mapstructure.Decode(a.options, &createOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase User Management `create user` action options
	validate := validator.New()
	if err := validate.Struct(createOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build create action
	ctx := context.TODO()
	client, err := a.client.Auth(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	params := &auth.UserToCreate{}
	if createOptions.object.uid != "" {
		params.UID(createOptions.object.uid)
	}
	params.
		Email(createOptions.object.email).
		EmailVerified(createOptions.object.emailVerified).
		PhoneNumber(createOptions.object.phoneNumber).
		Password(createOptions.object.password).
		DisplayName(createOptions.object.displayName).
		PhotoURL(createOptions.object.photoURL).
		Disabled(createOptions.object.disabled)

	user, err := client.CreateUser(ctx, params)

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"user": user}}}, err
}

func (a *AuthOperationRunner) update() (common.RuntimeResult, error) {
	var updateOptions AuthUpdateOptions
	if err := mapstructure.Decode(a.options, &updateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase User Management `update user` action options
	validate := validator.New()
	if err := validate.Struct(updateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build update action
	ctx := context.TODO()
	client, err := a.client.Auth(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	params := &auth.UserToUpdate{}
	params.
		Email(updateOptions.object.email).
		EmailVerified(updateOptions.object.emailVerified).
		PhoneNumber(updateOptions.object.phoneNumber).
		Password(updateOptions.object.password).
		DisplayName(updateOptions.object.displayName).
		PhotoURL(updateOptions.object.photoURL).
		Disabled(updateOptions.object.disabled)

	user, err := client.UpdateUser(ctx, updateOptions.uid, params)

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"user": user}}}, err
}

func (a *AuthOperationRunner) delete() (common.RuntimeResult, error) {
	var deleteOptions AuthBaseOptions
	if err := mapstructure.Decode(a.options, &deleteOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase User Management `delete user` action options
	validate := validator.New()
	if err := validate.Struct(deleteOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build delete action
	ctx := context.TODO()
	client, err := a.client.Auth(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	err = client.DeleteUser(ctx, deleteOptions.filter)

	return common.RuntimeResult{Success: true}, err
}

func (a *AuthOperationRunner) list() (common.RuntimeResult, error) {
	var listOptions AuthListOptions
	if err := mapstructure.Decode(a.options, &listOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase User Management `list users` action options
	validate := validator.New()
	if err := validate.Struct(listOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build list action
	ctx := context.TODO()
	client, err := a.client.Auth(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	pageSize := 1000
	iter := client.Users(ctx, listOptions.token)
	if listOptions.number > 0 {
		pageSize = listOptions.number
	}
	pager := iterator.NewPager(iter, pageSize, "")
	var users []*auth.ExportedUserRecord
	nextPageToken, err := pager.NextPage(&users)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{
		Success: true,
		Rows: []map[string]interface{}{
			{"users": users},
			{"nextPageToken": nextPageToken},
		},
	}, nil
}
