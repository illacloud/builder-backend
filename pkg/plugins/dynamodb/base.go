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

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mitchellh/mapstructure"
)

const (
	QUERY_METHOD       = "query"
	SCAN_METHOD        = "scan"
	PUT_ITEM_METHOD    = "putItem"
	GET_ITEM_METHOD    = "getItem"
	UPDATE_ITEM_METHOD = "updateItem"
	DELETE_ITEM_METHOD = "deleteItem"
)

func (d *Connector) getClientWithOptions(resourceOptions map[string]interface{}) (*dynamodb.Client, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &d.ResourceOpts); err != nil {
		return nil, err
	}

	// format the parameters for the session you want to create.
	creds := credentials.NewStaticCredentialsProvider(d.ResourceOpts.AccessKeyID, d.ResourceOpts.SecretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(d.ResourceOpts.Region), config.WithCredentialsProvider(creds))
	if err != nil {
		return nil, err
	}

	// Using the Config value, create the DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	return client, nil
}
