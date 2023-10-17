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

package couchdb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	resourceOptions resource
	actionOptions   action
}

func (c *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &c.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate couchdb options
	validate := validator.New()
	if err := validate.Struct(c.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (c *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &c.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate couchdb options
	validate := validator.New()
	if err := validate.Struct(c.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (c *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get couchdb client
	client, err := c.getClient(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	// test couchdb connection
	if _, err := client.Version(context.Background()); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (c *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get couchdb client
	client, err := c.getClient(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get all databases
	dbs, err := client.AllDBs(context.Background())
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"databases": dbs},
	}, nil
}

func (c *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get couchdb client
	client, err := c.getClient(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &c.actionOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// get database
	db := client.DB(c.actionOptions.Database)

	// switch based on different method
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	switch c.actionOptions.Method {
	case LIST_METHOD:
		limitType := reflect.TypeOf(c.actionOptions.Opts["limit"])
		if limitType.Kind() == reflect.Float64 {
			floatTmp := c.actionOptions.Opts["limit"].(float64)
			c.actionOptions.Opts["limit"] = int(floatTmp)
		} else {
			delete(c.actionOptions.Opts, "limit")
		}
		skipType := reflect.TypeOf(c.actionOptions.Opts["skip"])
		if skipType.Kind() == reflect.Float64 {
			floatTmp := c.actionOptions.Opts["skip"].(float64)
			c.actionOptions.Opts["skip"] = int(floatTmp)
		} else {
			delete(c.actionOptions.Opts, "skip")
		}
		c.actionOptions.Opts["include_docs"] = c.actionOptions.Opts["includeDocs"]
		delete(c.actionOptions.Opts, "includeDocs")
		c.actionOptions.Opts["descending_order"] = c.actionOptions.Opts["descendingOrder"]
		delete(c.actionOptions.Opts, "descending_order")
		resSet := db.AllDocs(context.TODO(), c.actionOptions.Opts)
		rows := make([]map[string]interface{}, 0)
		for resSet.Next() {
			item := make(map[string]interface{}, 3)
			item["_id"] = resSet.ID()
			var rev map[string]interface{}
			resSet.ScanValue(&rev)
			item["_rev"] = rev["rev"]
			if v, ok := c.actionOptions.Opts["include_docs"].(bool); ok && v {
				var doc interface{}
				resSet.ScanDoc(&doc)
				item["doc"] = doc
			}
			rows = append(rows, item)
		}
		resObj := make(map[string]interface{})
		resMetadata, _ := resSet.Finish()
		if resSet.Err() != nil {
			res.Rows = append(res.Rows, map[string]interface{}{"error": resSet.Err().Error()})
			return res, nil
		}
		resObj["rows"] = rows
		resObj["total_rows"] = resMetadata.TotalRows
		resObj["offset"] = resMetadata.Offset
		res.Rows = append(res.Rows, resObj)
		res.Success = true
	case RETRIEVE_METHOD:
		docID, ok := c.actionOptions.Opts["_id"].(string)
		if !ok {
			return res, errors.New("doc id is required")
		}
		resSet := db.Get(context.TODO(), docID)
		var doc map[string]interface{}
		resSet.ScanDoc(&doc)
		resSet.Close()
		if resSet.Err() != nil {
			res.Rows = append(res.Rows, map[string]interface{}{"error": resSet.Err().Error()})
			return res, nil
		}
		res.Rows = append(res.Rows, doc)
		res.Success = true
	case CREATE_METHOD:
		docID, rev, err := db.CreateDoc(context.TODO(), c.actionOptions.Opts["record"])
		if err != nil {
			return res, err
		}
		res.Rows = append(res.Rows, map[string]interface{}{"_id": docID, "_rev": rev})
		res.Success = true
	case UPDATE_METHOD:
		type updateOpts struct {
			Record map[string]interface{} `mapstructure:"record"`
			ID     string                 `mapstructure:"_id"`
			Rev    string                 `mapstructure:"_rev"`
		}
		var opts updateOpts
		if err := mapstructure.Decode(c.actionOptions.Opts, &opts); err != nil {
			return res, err
		}
		opts.Record["_rev"] = opts.Rev

		newRev, err := db.Put(context.TODO(), opts.ID, opts.Record)
		if err != nil {
			return res, err
		}
		res.Rows = append(res.Rows, map[string]interface{}{"message": fmt.Sprintf("updated %s, new revision ID: %s", opts.ID, newRev)})
		res.Success = true
	case DELETE_METHOD:
		docID, ok := c.actionOptions.Opts["_id"].(string)
		if !ok {
			return res, errors.New("doc id is required")
		}
		rev, ok := c.actionOptions.Opts["_rev"].(string)
		if !ok {
			return res, errors.New("revision id is required")
		}
		if _, err := db.Delete(context.TODO(), docID, rev); err != nil {
			return res, err
		}
		res.Rows = append(res.Rows, map[string]interface{}{"message": fmt.Sprintf("deleted %s document", docID)})
		res.Success = true
	case FIND_METHOD:
		resSet := db.Find(context.TODO(), c.actionOptions.Opts["mangoQuery"])
		rows := make([]map[string]interface{}, 0)
		for resSet.Next() {
			var doc map[string]interface{}
			resSet.ScanDoc(&doc)
			rows = append(rows, doc)
		}
		resObj := make(map[string]interface{})
		resMetadata, _ := resSet.Finish()
		if resSet.Err() != nil {
			res.Rows = append(res.Rows, map[string]interface{}{"error": resSet.Err().Error()})
			return res, nil
		}
		resObj["docs"] = rows
		resObj["bookmark"] = resMetadata.Bookmark
		resObj["warning"] = resMetadata.Warning
		res.Rows = append(res.Rows, resObj)
		res.Success = true
	case GET_METHOD:
		type getViewOpts struct {
			ViewURL  string      `mapstructure:"viewurl"`
			StartKey string      `mapstructure:"startkey"`
			EndKey   string      `mapstructure:"endkey"`
			Skip     interface{} `mapstructure:"skip"`
			Limit    interface{} `mapstructure:"limit"`
			Include  bool        `mapstructure:"includeDocs"`
		}
		var opts getViewOpts
		if err := mapstructure.Decode(c.actionOptions.Opts, &opts); err != nil {
			return res, err
		}
		viewURLSlice := strings.Split(opts.ViewURL, "/")
		if len(viewURLSlice) != 4 {
			return res, errors.New("invalid view url")
		}
		kOpts := make(map[string]interface{})
		if opts.StartKey != "" {
			kOpts["startkey"] = opts.StartKey
		}
		if opts.EndKey != "" {
			kOpts["endkey"] = opts.EndKey + kivik.EndKeySuffix
		}
		kOpts["include_docs"] = opts.Include
		limitType := reflect.TypeOf(opts.Limit)
		if limitType.Kind() == reflect.Float64 {
			floatTmp := opts.Limit.(float64)
			kOpts["limit"] = int(floatTmp)
		}
		skipType := reflect.TypeOf(opts.Skip)
		if skipType.Kind() == reflect.Float64 {
			floatTmp := opts.Skip.(float64)
			kOpts["skip"] = int(floatTmp)
		}
		resSet := db.Query(context.TODO(), "_design/"+viewURLSlice[1], "_view/"+viewURLSlice[3], kOpts)
		rows := make([]map[string]interface{}, 0)
		for resSet.Next() {
			item := make(map[string]interface{}, 3)
			item["_id"] = resSet.ID()
			var rev map[string]interface{}
			resSet.ScanValue(&rev)
			item["_rev"] = rev["_rev"]
			if opts.Include {
				var doc interface{}
				resSet.ScanDoc(&doc)
				item["doc"] = doc
			}
			rows = append(rows, item)
		}
		resObj := make(map[string]interface{})
		resMetadata, _ := resSet.Finish()
		if resSet.Err() != nil {
			res.Rows = append(res.Rows, map[string]interface{}{"error": resSet.Err().Error()})
			return res, nil
		}
		resObj["rows"] = rows
		resObj["total_rows"] = resMetadata.TotalRows
		resObj["offset"] = resMetadata.Offset
		res.Rows = append(res.Rows, resObj)
		res.Success = true
	default:
		return res, errors.New("invalid method")
	}

	return res, nil
}
