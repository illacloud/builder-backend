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

type Resource struct {
	Region          string `validate:"required"`
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

type Action struct {
	Method       string `validate:"required,oneof=query scan putItem getItem updateItem deleteItem"`
	Table        string
	UseJson      bool
	Parameters   string
	StructParams map[string]interface{}
}

type QueryParams struct {
	IndexName                 string
	KeyConditionExpression    string
	ProjectionExpression      string
	FilterExpression          string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
	Limit                     int32
	Select                    string
}

type ScanParams struct {
	IndexName                 string
	ProjectionExpression      string
	FilterExpression          string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
	Limit                     int32
	Select                    string
}

type PutItemParams struct {
	Item                      map[string]interface{}
	ConditionExpression       string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
}

type GetItemParams struct {
	Key                      map[string]interface{}
	ProjectionExpression     string
	ExpressionAttributeNames map[string]string
}

type UpdateItemParams struct {
	Key                       map[string]interface{}
	UpdateExpression          string
	ConditionExpression       string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
}

type DeleteItemParams struct {
	Key                       map[string]interface{}
	ConditionExpression       string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]interface{}
}
