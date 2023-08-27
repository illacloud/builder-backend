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
	"strings"
	"time"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type CommandExecutor struct {
	client  *s3.Client
	command Action
	bucket  string
}

func (c *CommandExecutor) listObjects(region string) (common.RuntimeResult, error) {
	var listCommandArgs ListCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &listCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 list action options
	validate := validator.New()
	if err := validate.Struct(listCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && listCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if listCommandArgs.BucketName == "" {
		listCommandArgs.BucketName = c.bucket
	}
	if listCommandArgs.MaxKeys == 0 {
		listCommandArgs.MaxKeys = 100
	}

	// build listObjectInput
	params := s3.ListObjectsV2Input{
		Bucket:    &listCommandArgs.BucketName,
		Delimiter: &listCommandArgs.Delimiter,
		Prefix:    &listCommandArgs.Prefix,
		MaxKeys:   listCommandArgs.MaxKeys,
	}

	res, err := c.client.ListObjectsV2(context.TODO(), &params)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	listObjRes := make([]map[string]interface{}, 0, len(res.Contents))
	for _, obj := range res.Contents {
		objRes := map[string]interface{}{"objectKey": *obj.Key}
		expiryDuration := time.Duration(listCommandArgs.Expiry) * time.Minute
		signedURL, _ := presignGetObject(c.client, listCommandArgs.BucketName, *obj.Key,
			expiryDuration)
		if listCommandArgs.SignedURL {
			objRes["signedURL"] = signedURL
			objRes["urlExpiryDate"] = time.Now().UTC().Add(expiryDuration).Format("2006.01.02 15:04:07.000 UTC")
		} else {
			parts := strings.SplitN(signedURL, "?", 2)
			if len(parts) != 0 {
				objRes["url"] = parts[0]
			}
		}
		listObjRes = append(listObjRes, objRes)
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    listObjRes,
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) readAnObject(region string) (common.RuntimeResult, error) {
	var readCommandArgs BaseCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &readCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 read action options
	validate := validator.New()
	if err := validate.Struct(readCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && readCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if readCommandArgs.BucketName == "" {
		readCommandArgs.BucketName = c.bucket
	}

	// build get presigned url
	urlObj := make(map[string]interface{}, 2)
	urlObj["key"] = readCommandArgs.ObjectKey
	expiryDuration := time.Duration(readCommandArgs.Expiry) * time.Minute
	signedURL, err := presignGetObject(c.client, readCommandArgs.BucketName, readCommandArgs.ObjectKey,
		expiryDuration)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	if readCommandArgs.SignedURL {
		urlObj["url"] = signedURL
	} else {
		parts := strings.SplitN(signedURL, "?", 2)
		if len(parts) != 0 {
			urlObj["url"] = parts[0]
		}
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{urlObj},
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) downloadAnObject(region string) (common.RuntimeResult, error) {
	var downloadCommandArgs BaseCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &downloadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 download action options
	validate := validator.New()
	if err := validate.Struct(downloadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && downloadCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if downloadCommandArgs.BucketName == "" {
		downloadCommandArgs.BucketName = c.bucket
	}

	// build get presigned url
	urlObj := make(map[string]interface{}, 2)
	urlObj["key"] = downloadCommandArgs.ObjectKey
	expiryDuration := time.Duration(downloadCommandArgs.Expiry) * time.Minute
	signedURL, err := presignGetObject(c.client, downloadCommandArgs.BucketName, downloadCommandArgs.ObjectKey,
		expiryDuration)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	if downloadCommandArgs.SignedURL {
		urlObj["url"] = signedURL
	} else {
		parts := strings.SplitN(signedURL, "?", 2)
		if len(parts) != 0 {
			urlObj["url"] = parts[0]
		}
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{urlObj},
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) deleteAnObject() (common.RuntimeResult, error) {
	var delete1CommandArgs BaseCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &delete1CommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 delete action options
	validate := validator.New()
	if err := validate.Struct(delete1CommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && delete1CommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if delete1CommandArgs.BucketName == "" {
		delete1CommandArgs.BucketName = c.bucket
	}

	// build DeleteObjectInput
	params := s3.DeleteObjectInput{
		Bucket: &delete1CommandArgs.BucketName,
		Key:    &delete1CommandArgs.ObjectKey,
	}

	res, err := c.client.DeleteObject(context.TODO(), &params)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{{"objectKey": delete1CommandArgs.ObjectKey, "deleteMarker": res.DeleteMarker}},
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) deleteMultipleObjects() (common.RuntimeResult, error) {
	var batchDeleteCommandArgs BatchDeleteCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &batchDeleteCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 batchDelete action options
	validate := validator.New()
	if err := validate.Struct(batchDeleteCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && batchDeleteCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if batchDeleteCommandArgs.BucketName == "" {
		batchDeleteCommandArgs.BucketName = c.bucket
	}

	// run PutObject for BatchUpload
	batchN := len(batchDeleteCommandArgs.ObjectKeyList)
	failedKeys := make([]string, 0, batchN)
	successN := 0
	for i := 0; i < batchN; i++ {
		// build DeleteObjectInput
		params := s3.DeleteObjectInput{
			Bucket: &batchDeleteCommandArgs.BucketName,
			Key:    &batchDeleteCommandArgs.ObjectKeyList[i],
		}

		_, err := c.client.DeleteObject(context.TODO(), &params)
		if err != nil {
			failedKeys = append(failedKeys, batchDeleteCommandArgs.ObjectKeyList[i])
			continue
		}

		successN += 1
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{{"count": batchN, "success": successN, "failure": failedKeys}},
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) uploadAnObject(ACL string) (common.RuntimeResult, error) {
	var uploadCommandArgs UploadCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &uploadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 upload action options
	validate := validator.New()
	if err := validate.Struct(uploadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && uploadCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if uploadCommandArgs.BucketName == "" {
		uploadCommandArgs.BucketName = c.bucket
	}

	// build put presigned url
	expiryDuration := time.Duration(uploadCommandArgs.Expiry) * time.Minute
	signedURL, err := presignPutObject(c.client, uploadCommandArgs.BucketName, uploadCommandArgs.ObjectKey, ACL, expiryDuration)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	urlObj := make(map[string]interface{}, 3)
	urlObj["url"] = signedURL
	urlObj["key"] = uploadCommandArgs.ObjectKey
	urlObj["acl"] = ACL

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{urlObj},
		Extra:   nil,
	}, nil
}

func (c *CommandExecutor) uploadMultipleObjects(ACL string) (common.RuntimeResult, error) {
	var batchUploadCommandArgs BatchUploadCommandArgs
	if err := mapstructure.Decode(c.command.CommandArgs, &batchUploadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate s3 upload action options
	validate := validator.New()
	if err := validate.Struct(batchUploadCommandArgs); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if c.bucket == "" && batchUploadCommandArgs.BucketName == "" {
		return common.RuntimeResult{Success: false}, errors.New("no bucket name")
	}
	if batchUploadCommandArgs.BucketName == "" {
		batchUploadCommandArgs.BucketName = c.bucket
	}
	batchN := len(batchUploadCommandArgs.ObjectKeyList)
	if len(batchUploadCommandArgs.ObjectKeyList) != len(batchUploadCommandArgs.ObjectDataList) {
		return common.RuntimeResult{Success: false}, errors.New("mismatch between object keys and object data")
	}

	// build put presigned urls
	res := make([]map[string]interface{}, 0, batchN)
	for i := 0; i < batchN; i++ {
		expiryDuration := time.Duration(batchUploadCommandArgs.Expiry) * time.Minute
		signedURL, _ := presignPutObject(c.client, batchUploadCommandArgs.BucketName, batchUploadCommandArgs.ObjectKeyList[i], ACL, expiryDuration)
		urlObj := make(map[string]interface{}, 3)
		urlObj["url"] = signedURL
		urlObj["key"] = batchUploadCommandArgs.ObjectKeyList[i]
		urlObj["acl"] = ACL
		res = append(res, urlObj)
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    res,
		Extra:   nil,
	}, nil
}
