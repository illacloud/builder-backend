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
	"encoding/json"
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (d *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &d.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Amazon DynamoDB resource options
	validate := validator.New()
	if err := validate.Struct(d.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (d *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &d.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Amazon DynamoDB action options
	validate := validator.New()
	if err := validate.Struct(d.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (d *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get dynamodb client
	svc, err := d.getClientWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test dynamodb client connection
	if _, err := svc.ListTables(context.TODO(), nil); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (d *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get dynamodb client
	svc, err := d.getClientWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get dynamodb tables
	resp, err := svc.ListTables(context.TODO(), nil)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"tables": resp.TableNames},
	}, nil
}

func (d *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get dynamodb client
	svc, err := d.getClientWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &d.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	if d.ActionOpts.UseJson {
		var res map[string]interface{}
		if err := json.Unmarshal([]byte(d.ActionOpts.Parameters), &res); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		d.ActionOpts.StructParams = res
	}

	// start a default context
	ctx, cancel := context.WithTimeout(context.TODO(), common.DEFAULT_QUERY_AND_EXEC_TIMEOUT)
	defer cancel()

	// switch based on different method
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	switch d.ActionOpts.Method {
	case QUERY_METHOD:
		in, err := buildQueryInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		out, err := svc.Query(ctx, in)
		if err != nil {
			return res, err
		}
		rows := make([]map[string]interface{}, len(out.Items))
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &rows); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = rows
	case SCAN_METHOD:
		in, err := buildScanInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		out, err := svc.Scan(ctx, in)
		if err != nil {
			return res, err
		}
		rows := make([]map[string]interface{}, len(out.Items))
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &rows); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = rows
	case PUT_ITEM_METHOD:
		in, err := buildPutItemInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		if _, err := svc.PutItem(ctx, in); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = append(res.Rows, map[string]interface{}{"message": "put item successfully"})
	case GET_ITEM_METHOD:
		in, err := buildGetItemInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		out, err := svc.GetItem(ctx, in)
		if err != nil {
			return res, err
		}
		m := make(map[string]interface{})
		if err := attributevalue.UnmarshalMap(out.Item, &m); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = append(res.Rows, m)
	case UPDATE_ITEM_METHOD:
		in, err := buildUpdateItemInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		if _, err := svc.UpdateItem(ctx, in); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = append(res.Rows, map[string]interface{}{"message": "update item successfully"})
	case DELETE_ITEM_METHOD:
		in, err := buildDeleteItemInput(d.ActionOpts.Table, d.ActionOpts.StructParams)
		if err != nil {
			return res, err
		}
		if _, err := svc.DeleteItem(ctx, in); err != nil {
			return res, err
		}
		res.Success = true
		res.Rows = append(res.Rows, map[string]interface{}{"message": "delete item successfully"})
	default:
		return res, errors.New("unsupported dynamodb method")
	}

	return res, nil
}

func buildQueryInput(table string, params map[string]interface{}) (*dynamodb.QueryInput, error) {
	var queryParams QueryParams
	if err := mapstructure.Decode(params, &queryParams); err != nil {
		return nil, err
	}

	res := &dynamodb.QueryInput{
		TableName:              aws.String(table),
		KeyConditionExpression: aws.String(queryParams.KeyConditionExpression),
	}
	if queryParams.IndexName != "" {
		res.IndexName = aws.String(queryParams.IndexName)
	}
	if queryParams.ProjectionExpression != "" {
		res.ProjectionExpression = aws.String(queryParams.ProjectionExpression)
	}
	if queryParams.FilterExpression != "" {
		res.FilterExpression = aws.String(queryParams.FilterExpression)
	}
	if len(queryParams.ExpressionAttributeValues) != 0 {
		expressionAttributeValues, err := attributevalue.MarshalMap(queryParams.ExpressionAttributeValues)
		if err != nil {
			return nil, err
		}
		res.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(queryParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = queryParams.ExpressionAttributeNames
	}
	if queryParams.Limit > 0 {
		res.Limit = aws.Int32(queryParams.Limit)
	}
	if queryParams.Select != "" {
		res.Select = types.Select(queryParams.Select)
	}

	return res, nil
}

func buildScanInput(table string, params map[string]interface{}) (*dynamodb.ScanInput, error) {
	var scanParams ScanParams
	if err := mapstructure.Decode(params, &scanParams); err != nil {
		return nil, err
	}

	res := &dynamodb.ScanInput{
		TableName: aws.String(table),
	}
	if scanParams.IndexName != "" {
		res.IndexName = aws.String(scanParams.IndexName)
	}
	if scanParams.ProjectionExpression != "" {
		res.ProjectionExpression = aws.String(scanParams.ProjectionExpression)
	}
	if scanParams.FilterExpression != "" {
		res.FilterExpression = aws.String(scanParams.FilterExpression)
	}
	if len(scanParams.ExpressionAttributeValues) != 0 {
		expressionAttributeValues, err := attributevalue.MarshalMap(scanParams.ExpressionAttributeValues)
		if err != nil {
			return nil, err
		}
		res.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(scanParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = scanParams.ExpressionAttributeNames
	}
	if scanParams.Limit > 0 {
		res.Limit = aws.Int32(scanParams.Limit)
	}
	if scanParams.Select != "" {
		res.Select = types.Select(scanParams.Select)
	}

	return res, nil
}

func buildPutItemInput(table string, params map[string]interface{}) (*dynamodb.PutItemInput, error) {
	var putItemParams PutItemParams
	if err := mapstructure.Decode(params, &putItemParams); err != nil {
		return nil, err
	}

	item, err := attributevalue.MarshalMap(putItemParams.Item)
	if err != nil {
		return nil, err
	}

	res := &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	}
	if putItemParams.ConditionExpression != "" {
		res.ConditionExpression = aws.String(putItemParams.ConditionExpression)
	}
	if len(putItemParams.ExpressionAttributeValues) != 0 {
		expressionAttributeValues, err := attributevalue.MarshalMap(putItemParams.ExpressionAttributeValues)
		if err != nil {
			return nil, err
		}
		res.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(putItemParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = putItemParams.ExpressionAttributeNames
	}

	return res, nil
}

func buildGetItemInput(table string, params map[string]interface{}) (*dynamodb.GetItemInput, error) {
	var getItemParams GetItemParams
	if err := mapstructure.Decode(params, &getItemParams); err != nil {
		return nil, err
	}

	itemKey, err := attributevalue.MarshalMap(getItemParams.Key)
	if err != nil {
		return nil, err
	}
	res := &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       itemKey,
	}
	if len(getItemParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = getItemParams.ExpressionAttributeNames
	}
	if getItemParams.ProjectionExpression != "" {
		res.ProjectionExpression = aws.String(getItemParams.ProjectionExpression)
	}

	return res, nil
}

func buildUpdateItemInput(table string, params map[string]interface{}) (*dynamodb.UpdateItemInput, error) {
	var updateItemParams UpdateItemParams
	if err := mapstructure.Decode(params, &updateItemParams); err != nil {
		return nil, err
	}

	itemKey, err := attributevalue.MarshalMap(updateItemParams.Key)
	if err != nil {
		return nil, err
	}
	res := &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key:       itemKey,
	}
	if updateItemParams.ConditionExpression != "" {
		res.ConditionExpression = aws.String(updateItemParams.ConditionExpression)
	}
	if updateItemParams.UpdateExpression != "" {
		res.UpdateExpression = aws.String(updateItemParams.UpdateExpression)
	}
	if len(updateItemParams.ExpressionAttributeValues) != 0 {
		expressionAttributeValues, err := attributevalue.MarshalMap(updateItemParams.ExpressionAttributeValues)
		if err != nil {
			return nil, err
		}
		res.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(updateItemParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = updateItemParams.ExpressionAttributeNames
	}

	return res, nil
}

func buildDeleteItemInput(table string, params map[string]interface{}) (*dynamodb.DeleteItemInput, error) {
	var deleteItemParams DeleteItemParams
	if err := mapstructure.Decode(params, &deleteItemParams); err != nil {
		return nil, err
	}

	itemKey, err := attributevalue.MarshalMap(deleteItemParams.Key)
	if err != nil {
		return nil, err
	}
	res := &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key:       itemKey,
	}
	if deleteItemParams.ConditionExpression != "" {
		res.ConditionExpression = aws.String(deleteItemParams.ConditionExpression)
	}
	if len(deleteItemParams.ExpressionAttributeValues) != 0 {
		expressionAttributeValues, err := attributevalue.MarshalMap(deleteItemParams.ExpressionAttributeValues)
		if err != nil {
			return nil, err
		}
		res.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(deleteItemParams.ExpressionAttributeNames) != 0 {
		res.ExpressionAttributeNames = deleteItemParams.ExpressionAttributeNames
	}

	return res, nil
}
