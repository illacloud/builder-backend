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

package s3

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

	// validate s3 options
	validate := validator.New()
	if err := validate.Struct(s.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate s3 ACL
	if s.ResourceOpts.ACL != "" {
		if _, ok := ACLs[s.ResourceOpts.ACL]; !ok {
			return common.ValidateResult{Valid: false}, errors.New("invalid ACL")
		}
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &s.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate s3 options
	validate := validator.New()
	if err := validate.Struct(s.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get s3 client
	s3Client, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test s3 client
	if _, err := s3Client.ListBuckets(context.TODO(), nil); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (s *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get s3 client
	s3Client, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get s3 bucket
	buckets, err := s3Client.ListBuckets(context.TODO(), nil)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"buckets": buckets.Buckets},
	}, nil
}

func (s *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get s3 client
	s3Client, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get s3 client")
	}

	// format s3 action
	if err := mapstructure.Decode(actionOptions, &s.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var result common.RuntimeResult
	commandExecutor := CommandExecutor{client: s3Client, command: s.ActionOpts, bucket: s.ResourceOpts.BucketName}
	switch s.ActionOpts.Commands {
	case LIST_COMMAND:
		result, err = commandExecutor.listObjects(s.ResourceOpts.Region)
	case READ_COMMAND:
		result, err = commandExecutor.readAnObject(s.ResourceOpts.Region)
	case DOWNLOAD_COMMAND:
		result, err = commandExecutor.downloadAnObject(s.ResourceOpts.Region)
	case DELETE_COMMAND:
		result, err = commandExecutor.deleteAnObject()
	case BATCH_DELETE_COMMAND:
		result, err = commandExecutor.deleteMultipleObjects()
	case UPLOAD_COMMAND:
		result, err = commandExecutor.uploadAnObject(s.ResourceOpts.ACL)
	case BATCH_UPLOAD_COMMAND:
		result, err = commandExecutor.uploadMultipleObjects(s.ResourceOpts.ACL)
	}

	return result, err
}
