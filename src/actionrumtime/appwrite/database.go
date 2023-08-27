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

package appwrite

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/illacloud/appwrite-sdk-go/appwrite"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
)

type ActionExecutor struct {
	client   *appwrite.Databases
	action   Action
	database string
}

func (a *ActionExecutor) ListDocs() (common.RuntimeResult, error) {
	var listOpts ListOpts
	if err := mapstructure.Decode(a.action.Opts, &listOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate opts
	if listOpts.CollectionID == "" {
		return common.RuntimeResult{}, errors.New("collectionID is required")
	}

	// build queries
	queriesArray := make([]interface{}, 0)
	for _, filter := range listOpts.Filter {
		if filter.Attribute == "" {
			continue
		}
		query := fmt.Sprintf("%s(%s, %s)", filter.Operator, filter.Attribute, filter.Value)
		queriesArray = append(queriesArray, query)
	}
	for _, order := range listOpts.OrderBy {
		if order.Attribute == "" {
			continue
		}
		if order.Value == "asc" {
			order := fmt.Sprintf("orderAsc(%s)", order.Attribute)
			queriesArray = append(queriesArray, order)
		} else if order.Value == "desc" {
			order := fmt.Sprintf("orderDesc(%s)", order.Attribute)
			queriesArray = append(queriesArray, order)
		}
	}
	limit := fmt.Sprintf("limit(%d)", listOpts.Limit)
	queriesArray = append(queriesArray, limit)

	// call ListDocuments
	listRes, err := a.client.ListDocuments(a.database, listOpts.CollectionID, queriesArray)
	if err != nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": err.Error(),
				"success": false,
				"result":  "",
			}},
		}, nil
	}
	if listRes == nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while getting the documents.",
				"success": false,
				"result":  "Unknown error",
			}},
		}, nil
	}
	if listRes.StatusCode != 200 {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while getting the documents.",
				"success": false,
				"result":  listRes.Result,
			}},
		}, nil
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal([]byte(listRes.Result), &res); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	if vs, ok := res["documents"].([]interface{}); ok {
		for _, v := range vs {
			if m, ok := v.(map[string]interface{}); ok {
				modifyMapKeysWithPattern(m, "$", "")
			}
		}
	}

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{0: res},
	}, nil
}

func (a *ActionExecutor) CreateDoc() (common.RuntimeResult, error) {
	var createOpts WithDataOpts
	if err := mapstructure.Decode(a.action.Opts, &createOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate opts
	if createOpts.CollectionID == "" {
		return common.RuntimeResult{}, errors.New("collectionID is required")
	}
	if createOpts.DocumentID == "" {
		return common.RuntimeResult{Success: false}, errors.New("documentID is required")
	}

	var emptyArray = []interface{}{}
	createRes, err := a.client.CreateDocument(a.database, createOpts.CollectionID, createOpts.DocumentID, createOpts.Data, emptyArray)
	if err != nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": err.Error(),
				"success": false,
				"result":  "",
			}},
		}, nil
	}
	if createRes == nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while creating the documents.",
				"success": false,
				"result":  "Unknown error",
			}},
		}, nil
	}
	if createRes.StatusCode != 201 {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while creating the document.",
				"success": false,
				"result":  createRes.Result,
			}},
		}, nil
	}

	return common.RuntimeResult{
		Success: true,
		Rows: []map[string]interface{}{0: {
			"message": "Document created successfully.",
			"success": true,
			"result":  createRes.Result,
		}},
	}, nil
}

func (a *ActionExecutor) GetDoc() (common.RuntimeResult, error) {
	var getOpts BaseOpts
	if err := mapstructure.Decode(a.action.Opts, &getOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate opts
	if getOpts.CollectionID == "" {
		return common.RuntimeResult{}, errors.New("collectionID is required")
	}
	if getOpts.DocumentID == "" {
		return common.RuntimeResult{Success: false}, errors.New("documentID is required")
	}

	getRes, err := a.client.GetDocument(a.database, getOpts.CollectionID, getOpts.DocumentID)
	if err != nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": err.Error(),
				"success": false,
				"result":  "",
			}},
		}, nil
	}
	if getRes == nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while getting the document.",
				"success": false,
				"result":  "Unknown error",
			}},
		}, nil
	}
	if getRes.StatusCode != 200 {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while getting the document.",
				"success": false,
				"result":  getRes.Result,
			}},
		}, nil
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal([]byte(getRes.Result), &res); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	modifyMapKeysWithPattern(res, "$", "")

	return common.RuntimeResult{
		Success: true,
		Rows:    []map[string]interface{}{0: res},
	}, nil
}

func (a *ActionExecutor) UpdateDoc() (common.RuntimeResult, error) {
	var updateOpts WithDataOpts
	if err := mapstructure.Decode(a.action.Opts, &updateOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate opts
	if updateOpts.CollectionID == "" {
		return common.RuntimeResult{}, errors.New("collectionID is required")
	}
	if updateOpts.DocumentID == "" {
		return common.RuntimeResult{Success: false}, errors.New("documentID is required")
	}

	var emptyArray = []interface{}{}
	updateRes, err := a.client.UpdateDocument(a.database, updateOpts.CollectionID, updateOpts.DocumentID, updateOpts.Data, emptyArray)
	if err != nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": err.Error(),
				"success": false,
				"result":  "",
			}},
		}, nil
	}
	if updateRes == nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while updating the document.",
				"success": false,
				"result":  "Unknown error",
			}},
		}, nil
	}
	if updateRes.StatusCode != 200 {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while updating the document.",
				"success": false,
				"result":  updateRes.Result,
			}},
		}, nil
	}

	return common.RuntimeResult{
		Success: true,
		Rows: []map[string]interface{}{0: {
			"message": "Document updated successfully.",
			"success": true,
			"result":  updateRes.Result,
		}},
	}, nil
}

func (a *ActionExecutor) DeleteDoc() (common.RuntimeResult, error) {
	var deleteOpts BaseOpts
	if err := mapstructure.Decode(a.action.Opts, &deleteOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// validate opts
	if deleteOpts.CollectionID == "" {
		return common.RuntimeResult{}, errors.New("collectionID is required")
	}
	if deleteOpts.DocumentID == "" {
		return common.RuntimeResult{Success: false}, errors.New("documentID is required")
	}

	deleteRes, err := a.client.DeleteDocument(a.database, deleteOpts.CollectionID, deleteOpts.DocumentID)
	if err != nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": err.Error(),
				"success": false,
				"result":  "",
			}},
		}, nil
	}
	if deleteRes == nil {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while deleting the document.",
				"success": false,
				"result":  "Unknown error",
			}},
		}, nil
	}
	if deleteRes.StatusCode != 204 {
		return common.RuntimeResult{Success: false,
			Rows: []map[string]interface{}{0: {
				"message": "An error occurred while deleting the document.",
				"success": false,
				"result":  deleteRes.Result,
			}},
		}, nil
	}

	return common.RuntimeResult{
		Success: true,
		Rows: []map[string]interface{}{0: {
			"message": "Document deleted successfully.",
			"success": true,
			"result":  deleteRes.Result,
		}},
	}, nil
}
