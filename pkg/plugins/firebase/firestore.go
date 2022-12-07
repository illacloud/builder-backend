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

	"cloud.google.com/go/firestore"
	"github.com/go-playground/validator/v10"
	"github.com/illa-family/builder-backend/pkg/plugins/common"
	"github.com/mitchellh/mapstructure"

	firebase "firebase.google.com/go/v4"
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
	collection     string `validate:"required"`
	where          [][]interface{}
	limit          int
	orderBy        string
	orderDirection string
	startAt        SimpleCursor `validate:"required"`
	endAt          SimpleCursor `validate:"required"`
}

type SimpleCursor struct {
	trigger bool
	value   interface{}
}

type FSDocValueOptions struct {
	collection string `validate:"required"`
	id         string
	value      map[string]interface{} `validate:"required"`
}

type FSDocIDOptions struct {
	collection string `validate:"required"`
	id         string `validate:"required"`
}

type FSGetCollsOptions struct {
	parent string
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
	collRef := client.Collection(queryFSOptions.collection)
	query := collRef.Query

	conditionN := len(queryFSOptions.where)
	for i := 0; i < conditionN; i++ {
		if len(queryFSOptions.where[i]) != 3 {
			break
		}
		query = query.Where(queryFSOptions.where[i][0].(string), queryFSOptions.where[i][1].(string), queryFSOptions.where[i][2])
	}

	if queryFSOptions.limit > 0 {
		query = query.Limit(queryFSOptions.limit)
	}

	if queryFSOptions.orderBy != "" {
		direct := firestore.Asc
		if queryFSOptions.orderDirection == "desc" {
			direct = firestore.Desc
		}
		query = query.OrderBy(queryFSOptions.orderBy, direct)
	}

	if queryFSOptions.startAt.trigger {
		query = query.EndAt(queryFSOptions.startAt.value)
	}

	if queryFSOptions.endAt.trigger {
		query = query.EndAt(queryFSOptions.endAt.value)
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

	if insertDocOptions.id != "" {
		_, err = client.Collection(insertDocOptions.collection).Doc(insertDocOptions.id).Set(ctx, insertDocOptions.value)
	} else {
		_, _, err = client.Collection(insertDocOptions.collection).Add(ctx, insertDocOptions.value)
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

	if updateDocOptions.id == "" {
		return common.RuntimeResult{Success: false}, errors.New("document id required")
	}

	_, err = client.Collection(updateDocOptions.collection).Doc(updateDocOptions.id).Set(ctx, updateDocOptions.value, firestore.MergeAll)

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

	doc, err := client.Collection(getDocByIDOptions.collection).Doc(getDocByIDOptions.id).Get(ctx)
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

	_, err = client.Collection(deleteDocOptions.collection).Doc(deleteDocOptions.id).Delete(ctx)

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

	if getCollsOptions.parent != "" {
		documentPath := ""
		if getCollsOptions.parent[0] == '/' {
			documentPath = getCollsOptions.parent[1:]
		} else {
			documentPath = getCollsOptions.parent
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
	collRef := client.CollectionGroup(queryCGOptions.collection)
	query := collRef.Query

	conditionN := len(queryCGOptions.where)
	for i := 0; i < conditionN; i++ {
		if len(queryCGOptions.where[i]) != 3 {
			break
		}
		query = query.Where(queryCGOptions.where[i][0].(string), queryCGOptions.where[i][1].(string), queryCGOptions.where[i][2])
	}

	if queryCGOptions.limit > 0 {
		query = query.Limit(queryCGOptions.limit)
	}

	if queryCGOptions.orderBy != "" {
		direct := firestore.Asc
		if queryCGOptions.orderDirection == "desc" {
			direct = firestore.Desc
		}
		query = query.OrderBy(queryCGOptions.orderBy, direct)
	}

	if queryCGOptions.startAt.trigger {
		query = query.EndAt(queryCGOptions.startAt.value)
	}

	if queryCGOptions.endAt.trigger {
		query = query.EndAt(queryCGOptions.endAt.value)
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
