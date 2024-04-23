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

package tencentcos

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

func (s *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &s.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate cos options
	validate := validator.New()
	if err := validate.Struct(s.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &s.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate cos options
	validate := validator.New()
	if err := validate.Struct(s.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get cos client
	cosClient, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test cos client
	ok, err := cosClient.Bucket.IsExist(context.Background())
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	} else if !ok {
		return common.ConnectionResult{Success: false}, errors.New("bucket does not exists")
	}
	return common.ConnectionResult{Success: true}, nil
}

func (s *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get cos client
	cosClient, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// test cos client
	ok, err := cosClient.Bucket.IsExist(context.Background())
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	} else if !ok {
		return common.MetaInfoResult{Success: false}, errors.New("bucket does not exists")
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"buckets": []string{s.ResourceOpts.BucketName}},
	}, nil
}

func (s *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// init cos client
	cosClient, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get tencent cos client")
	}

	// format cos action
	if err := mapstructure.Decode(actionOptions, &s.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var result common.RuntimeResult
	commandExecutor := CommandExecutor{client: cosClient, command: s.ActionOpts, resource: s.ResourceOpts}
	switch s.ActionOpts.Commands {
	case LIST_COMMAND:
		result, err = commandExecutor.listObjects()
	case GET_DOWNLOAD_URL_COMMAND:
		result, err = commandExecutor.getPresignedURL()
	}

	return result, err
}
