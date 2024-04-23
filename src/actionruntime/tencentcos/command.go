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
	"net/http"
	"time"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type CommandExecutor struct {
	client   *cos.Client
	command  Action
	resource Resource
}

func (c *CommandExecutor) listObjects() (common.RuntimeResult, error) {
	// get config
	var listCommandArgs ListCommandArgs
	errInDecode := mapstructure.Decode(c.command.CommandArgs, &listCommandArgs)
	if errInDecode != nil {
		return common.RuntimeResult{Success: false}, errInDecode
	}

	// validate cos list action options
	validate := validator.New()
	errInValidate := validate.Struct(listCommandArgs)
	if errInValidate != nil {
		return common.RuntimeResult{Success: false}, errInValidate
	}

	opt := &cos.BucketGetOptions{
		MaxKeys: listCommandArgs.MaxKeys,
	}

	// get file list
	filesInBucket, _, errInGetBucket := c.client.Bucket.Get(context.Background(), opt)
	if errInGetBucket != nil {
		return common.RuntimeResult{Success: false}, errInGetBucket
	}

	// export file list
	listObjRes := make([]map[string]interface{}, 0)
	for _, fileInstance := range filesInBucket.Contents {
		listObjRes = append(listObjRes, map[string]interface{}{"fileName": fileInstance.Key, "fileSize": fileInstance.Size})
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    listObjRes,
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) getPresignedURL() (common.RuntimeResult, error) {
	var getPresignedURLCommandArgs GetPresignedURLCommandArgs
	errInDecode := mapstructure.Decode(c.command.CommandArgs, &getPresignedURLCommandArgs)
	if errInDecode != nil {
		return common.RuntimeResult{Success: false}, errInDecode
	}

	// validate cos read action options
	validate := validator.New()
	errInValidate := validate.Struct(getPresignedURLCommandArgs)
	if errInValidate != nil {
		return common.RuntimeResult{Success: false}, errInValidate
	}

	// get presigned url
	ctx := context.Background()
	presignedURL, errInGetPresignedURL := c.client.Object.GetPresignedURL(ctx, http.MethodGet, getPresignedURLCommandArgs.FileName, c.resource.AccessKeyID, c.resource.SecretAccessKey, time.Hour, nil)
	if errInGetPresignedURL != nil {
		return common.RuntimeResult{Success: false}, errInGetPresignedURL
	}

	retContent := map[string]interface{}{"fileName": getPresignedURLCommandArgs.FileName, "downloadURL": presignedURL.String()}
	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{retContent},
		Extra:   nil,
	}, nil
}
