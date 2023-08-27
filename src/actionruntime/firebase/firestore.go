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
	"strings"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

const (
	FS_QUERY_FS_OP   = "query_fs"
	FS_INSERT_DOC_OP = "insert_doc"
	FS_UPDATE_DOC_OP = "update_doc"
	FS_GET_DOC_OP    = "get_doc"
	FS_DELETE_DOC_OP = "delete_doc"
	FS_GET_COLLS_OP  = "get_colls"
	FS_QUERY_COLL_OP = "query_coll"
)

type FirestoreOperationRunner struct {
	client    *firebase.App
	operation string
	options   map[string]interface{}
}

type FSQueryOptions struct {
	Collection     string `validate:"required"`
	CollectionType string `validate:"oneof=select input"`
	Where          []QueryCondition
	Limit          int
	OrderBy        string
	OrderDirection string
	StartAt        SimpleCursor `validate:"required"`
	EndAt          SimpleCursor `validate:"required"`
}

type QueryCondition struct {
	Field     string
	Condition string
	Value     interface{}
}

type SimpleCursor struct {
	Trigger bool
	Value   interface{}
}

type FSDocValueOptions struct {
	Collection string `validate:"required"`
	ID         string
	Value      map[string]interface{} `validate:"required"`
}

type FSDocIDOptions struct {
	Collection string `validate:"required"`
	ID         string `validate:"required"`
}

type FSGetCollsOptions struct {
	Parent string
}

func (f *FirestoreOperationRunner) run() (common.RuntimeResult, error) {
	var result common.RuntimeResult
	var err error
	switch f.operation {
	case FS_QUERY_FS_OP:
		result, err = f.queryFirestore()
	case FS_INSERT_DOC_OP:
		result, err = f.insertDoc()
	case FS_UPDATE_DOC_OP:
		result, err = f.updateDoc()
	case FS_GET_DOC_OP:
		result, err = f.getDocByID()
	case FS_DELETE_DOC_OP:
		result, err = f.deleteDoc()
	case FS_GET_COLLS_OP:
		result, err = f.getCollections()
	case FS_QUERY_COLL_OP:
		result, err = f.queryCollectionGroup()
	default:
		result.Success = false
		err = errors.New("unsupported operation")
	}
	return result, err
}

func (f *FirestoreOperationRunner) queryFirestore() (common.RuntimeResult, error) {
	var queryFSOptions FSQueryOptions
	if err := mapstructure.Decode(f.options, &queryFSOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `query firestore` action options
	validate := validator.New()
	if err := validate.Struct(queryFSOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build query firestore action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	collRef := client.Collection(queryFSOptions.Collection)
	query := collRef.Query

	conditionN := len(queryFSOptions.Where)
	for i := 0; i < conditionN; i++ {
		if queryFSOptions.Where[i].Field == "" {
			break
		}
		query = query.Where(queryFSOptions.Where[i].Field, queryFSOptions.Where[i].Condition, queryFSOptions.Where[i].Value)
	}

	if queryFSOptions.Limit > 0 {
		query = query.Limit(queryFSOptions.Limit)
	}

	if queryFSOptions.OrderBy != "" {
		direct := firestore.Asc
		if queryFSOptions.OrderDirection == "desc" {
			direct = firestore.Desc
		}
		query = query.OrderBy(queryFSOptions.OrderBy, direct)
	}

	if queryFSOptions.StartAt.Trigger {
		query = query.EndAt(queryFSOptions.StartAt.Value)
	}

	if queryFSOptions.EndAt.Trigger {
		query = query.EndAt(queryFSOptions.EndAt.Value)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	res := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		res = append(res, doc.Data())
	}

	return common.RuntimeResult{Success: true, Rows: res}, err
}

func (f *FirestoreOperationRunner) insertDoc() (common.RuntimeResult, error) {
	var insertDocOptions FSDocValueOptions
	if err := mapstructure.Decode(f.options, &insertDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `insert document` action options
	validate := validator.New()
	if err := validate.Struct(insertDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build insert document action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if insertDocOptions.ID != "" {
		_, err = client.Collection(insertDocOptions.Collection).Doc(insertDocOptions.ID).Set(ctx, insertDocOptions.Value)
	} else {
		_, _, err = client.Collection(insertDocOptions.Collection).Add(ctx, insertDocOptions.Value)
	}

	return common.RuntimeResult{Success: true}, err
}

func (f *FirestoreOperationRunner) updateDoc() (common.RuntimeResult, error) {
	var updateDocOptions FSDocValueOptions
	if err := mapstructure.Decode(f.options, &updateDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `update document` action options
	validate := validator.New()
	if err := validate.Struct(updateDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build update document action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if updateDocOptions.ID == "" {
		return common.RuntimeResult{Success: false}, errors.New("document id required")
	}

	_, err = client.Collection(updateDocOptions.Collection).Doc(updateDocOptions.ID).Set(ctx, updateDocOptions.Value, firestore.MergeAll)

	return common.RuntimeResult{Success: true}, err
}

func (f *FirestoreOperationRunner) getDocByID() (common.RuntimeResult, error) {
	var getDocByIDOptions FSDocIDOptions
	if err := mapstructure.Decode(f.options, &getDocByIDOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `get document by id` action options
	validate := validator.New()
	if err := validate.Struct(getDocByIDOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build get document by id action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	doc, err := client.Collection(getDocByIDOptions.Collection).Doc(getDocByIDOptions.ID).Get(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{doc.Data()}}, nil
}

func (f *FirestoreOperationRunner) deleteDoc() (common.RuntimeResult, error) {
	var deleteDocOptions FSDocIDOptions
	if err := mapstructure.Decode(f.options, &deleteDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `delete document` action options
	validate := validator.New()
	if err := validate.Struct(deleteDocOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build delete document action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	_, err = client.Collection(deleteDocOptions.Collection).Doc(deleteDocOptions.ID).Delete(ctx)

	return common.RuntimeResult{Success: true}, nil
}

func (f *FirestoreOperationRunner) getCollections() (common.RuntimeResult, error) {
	var getCollsOptions FSGetCollsOptions
	if err := mapstructure.Decode(f.options, &getCollsOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `get collections` action options
	validate := validator.New()
	if err := validate.Struct(getCollsOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build get collections action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var collsIter *firestore.CollectionIterator

	if getCollsOptions.Parent != "" {
		documentPath := ""
		if getCollsOptions.Parent[0] == '/' {
			documentPath = getCollsOptions.Parent[1:]
		} else {
			documentPath = getCollsOptions.Parent
		}
		docPaths := strings.Split(documentPath, "/")
		collsIter = client.Collection(docPaths[0]).Doc(documentPath[len(docPaths[0]):]).Collections(ctx)
	} else {
		collsIter = client.Collections(ctx)
	}

	colls, err := collsIter.GetAll()
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	res := make([]string, 0, len(colls))
	for _, coll := range colls {
		res = append(res, coll.ID)
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"collections": res}}}, nil
}

func (f *FirestoreOperationRunner) queryCollectionGroup() (common.RuntimeResult, error) {
	var queryCGOptions FSQueryOptions
	if err := mapstructure.Decode(f.options, &queryCGOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate Firebase Firestore `query collection group` action options
	validate := validator.New()
	if err := validate.Struct(queryCGOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// build query collection group action
	ctx := context.TODO()
	client, err := f.client.Firestore(ctx)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	collRef := client.CollectionGroup(queryCGOptions.Collection)
	query := collRef.Query

	conditionN := len(queryCGOptions.Where)
	for i := 0; i < conditionN; i++ {
		if queryCGOptions.Where[i].Field == "" {
			break
		}
		query = query.Where(queryCGOptions.Where[i].Field, queryCGOptions.Where[i].Condition, queryCGOptions.Where[i].Value)
	}

	if queryCGOptions.Limit > 0 {
		query = query.Limit(queryCGOptions.Limit)
	}

	if queryCGOptions.OrderBy != "" {
		direct := firestore.Asc
		if queryCGOptions.OrderDirection == "desc" {
			direct = firestore.Desc
		}
		query = query.OrderBy(queryCGOptions.OrderBy, direct)
	}

	if queryCGOptions.StartAt.Trigger {
		query = query.StartAt(queryCGOptions.StartAt.Value)
	}

	if queryCGOptions.EndAt.Trigger {
		query = query.EndAt(queryCGOptions.EndAt.Value)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	res := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		res = append(res, doc.Data())
	}

	return common.RuntimeResult{Success: true, Rows: res}, err
}
