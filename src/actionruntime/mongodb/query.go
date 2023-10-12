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

package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueryRunner struct {
	client *mongo.Client
	query  Query
	db     string
}

func (q *QueryRunner) aggregate() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var aggregateOptions AggregateContent
	if err := mapstructure.Decode(q.query.TypeContent, &aggregateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var aggregateStage []bson.D
	if aggregateOptions.Aggregation != "" && aggregateOptions.Aggregation != "[]" {
		if err := bson.UnmarshalExtJSON([]byte(aggregateOptions.Aggregation), true, &aggregateStage); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	// build `AggregateOptions`
	if aggregateOptions.Options == "" {
		aggregateOptions.Options = "{}"
	}
	var rawAggregateOptions map[string]interface{}
	if err := json.Unmarshal([]byte(aggregateOptions.Options), &rawAggregateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var parsedAggregateOptions AggregateOptions
	if err := mapstructure.Decode(rawAggregateOptions, &parsedAggregateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	opts := options.Aggregate()
	if _, ok := rawAggregateOptions["collation"]; ok {
		opts = opts.SetCollation(parsedAggregateOptions.Collation)
	}
	if _, ok := rawAggregateOptions["hint"]; ok {
		opts = opts.SetHint(parsedAggregateOptions.Hint)
	}
	if _, ok := rawAggregateOptions["batchSize"]; ok {
		opts = opts.SetBatchSize(parsedAggregateOptions.BatchSize)
	}

	cursor, err := coll.Aggregate(context.TODO(), aggregateStage, opts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) bulkWrite() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var bulkWriteOptions BulkWriteContent
	if err := mapstructure.Decode(q.query.TypeContent, &bulkWriteOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var bulkOperations []bson.D
	if bulkWriteOptions.Operations != "" && bulkWriteOptions.Operations != "[]" {
		if err := bson.UnmarshalExtJSON([]byte(bulkWriteOptions.Operations), true, &bulkOperations); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	models := make([]mongo.WriteModel, 0, len(bulkOperations))
	for _, operationOptions := range bulkOperations {
		var writeModel mongo.WriteModel
		tmpMap := operationOptions.Map()
		switch operationOptions[0].Key {
		case "insertOne":
			document := tmpMap["insertOne"].(bson.D)
			documentMap := document.Map()
			documentContent := documentMap["document"].(bson.D)
			writeModel = mongo.NewInsertOneModel().SetDocument(documentContent)
			models = append(models, writeModel)
		case "updateOne":
			uM := tmpMap["updateOne"].(bson.D)
			uMMap := uM.Map()
			filterCondition := uMMap["filter"].(bson.D)
			updateDoc := uMMap["update"].(bson.D)
			writeModel = mongo.NewUpdateOneModel().SetFilter(filterCondition).SetUpdate(updateDoc)
			models = append(models, writeModel)
		case "updateMany":
			uM := tmpMap["updateMany"].(bson.D)
			uMMap := uM.Map()
			filterCondition := uMMap["filter"].(bson.D)
			updateDoc := uMMap["update"].(bson.D)
			writeModel = mongo.NewUpdateManyModel().SetFilter(filterCondition).SetUpdate(updateDoc)
			models = append(models, writeModel)
		case "deleteOne":
			dO := tmpMap["deleteOne"].(bson.D)
			dOMap := dO.Map()
			filterCondition := dOMap["filter"].(bson.D)
			writeModel = mongo.NewDeleteOneModel().SetFilter(filterCondition)
			models = append(models, writeModel)
		case "deleteMany":
			dO := tmpMap["deleteMany"].(bson.D)
			dOMap := dO.Map()
			filterCondition := dOMap["filter"].(bson.D)
			writeModel = mongo.NewDeleteManyModel().SetFilter(filterCondition)
			models = append(models, writeModel)
		case "replaceOne":
			rO := tmpMap["replaceOne"].(bson.D)
			rOMap := rO.Map()
			filterCondition := rOMap["filter"].(bson.D)
			replacement := rOMap["replacement"].(bson.D)
			writeModel = mongo.NewReplaceOneModel().SetFilter(filterCondition).SetReplacement(replacement)
			models = append(models, writeModel)
		default:
			break
		}
	}
	results, err := coll.BulkWrite(context.TODO(), models)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) count() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var countOptions CountContent
	if err := mapstructure.Decode(q.query.TypeContent, &countOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if countOptions.Query != "" && countOptions.Query != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(countOptions.Query), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": count}}}, nil
}

func (q *QueryRunner) deleteMany() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var dMOptions DeleteManyContent
	if err := mapstructure.Decode(q.query.TypeContent, &dMOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if dMOptions.Filter != "" && dMOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(dMOptions.Filter), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	results, err := coll.DeleteMany(context.TODO(), filter)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) deleteOne() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var dOOptions DeleteOneContent
	if err := mapstructure.Decode(q.query.TypeContent, &dOOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if dOOptions.Filter != "" && dOOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(dOOptions.Filter), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	results, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) distinct() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var distinctOptions DistinctContent
	if err := mapstructure.Decode(q.query.TypeContent, &distinctOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if distinctOptions.Query != "" && distinctOptions.Query != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(distinctOptions.Query), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	// build `DistinctOptions`
	if distinctOptions.Options == "" {
		distinctOptions.Options = "{}"
	}
	var rawDistinctOptions map[string]interface{}
	if err := json.Unmarshal([]byte(distinctOptions.Options), &rawDistinctOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var parsedAggregateOptions DistinctOptions
	if err := mapstructure.Decode(rawDistinctOptions, &parsedAggregateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	opts := options.Distinct()
	if _, ok := rawDistinctOptions["collation"]; ok {
		opts = opts.SetCollation(parsedAggregateOptions.Collation)
	}

	results, err := coll.Distinct(context.TODO(), distinctOptions.Field, filter, opts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) find() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var findOptions FindContent
	if err := mapstructure.Decode(q.query.TypeContent, &findOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if findOptions.Query != "" && findOptions.Query != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(findOptions.Query), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	opts := options.Find()
	if findOptions.Projection != "" && findOptions.Projection != "{}" {
		var projection bson.D
		if err := bson.UnmarshalExtJSON([]byte(findOptions.Projection), true, &projection); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetProjection(projection)
	}
	if findOptions.SortBy != "" && findOptions.SortBy != "{}" {
		var sortBy bson.D
		if err := bson.UnmarshalExtJSON([]byte(findOptions.SortBy), true, &sortBy); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetSort(sortBy)
	}
	if findOptions.Limit != "" {
		limit, err := strconv.ParseInt(findOptions.Limit, 10, 64)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetLimit(limit)
	}
	if findOptions.Skip != "" {
		skip, err := strconv.ParseInt(findOptions.Skip, 10, 64)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetSkip(skip)
	}

	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) findOne() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var fOOptions FindOneContent
	if err := mapstructure.Decode(q.query.TypeContent, &fOOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if fOOptions.Query != "" && fOOptions.Query != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(fOOptions.Query), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}
	fmt.Printf("[DUMP] findOne().filter: %+v\n", filter)

	opts := options.FindOne()
	if fOOptions.Projection != "" && fOOptions.Projection != "{}" {
		var projection bson.D
		if err := bson.UnmarshalExtJSON([]byte(fOOptions.Projection), true, &projection); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetProjection(projection)
	}
	if fOOptions.Skip != "" {
		skip, err := strconv.ParseInt(fOOptions.Skip, 10, 64)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		opts = opts.SetSkip(skip)
	}

	var results bson.M
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&results)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) findOneAndUpdate() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var fAUOptions FindOneAndUpdateContent
	if err := mapstructure.Decode(q.query.TypeContent, &fAUOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if fAUOptions.Filter != "" && fAUOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(fAUOptions.Filter), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	update := bson.D{}
	if fAUOptions.Update != "" && fAUOptions.Update != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(fAUOptions.Update), true, &update); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	// build `FindOneAndUpdateOptions`
	if fAUOptions.Options == "" {
		fAUOptions.Options = "{}"
	}
	var rawFindOneAndUpdateOptions map[string]interface{}
	if err := json.Unmarshal([]byte(fAUOptions.Options), &rawFindOneAndUpdateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var parsedFindOneAndUpdateOptions FindOneAndUpdateOptions
	if err := mapstructure.Decode(rawFindOneAndUpdateOptions, &parsedFindOneAndUpdateOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	opts := options.FindOneAndUpdate()
	if _, ok := rawFindOneAndUpdateOptions["collation"]; ok {
		opts = opts.SetCollation(parsedFindOneAndUpdateOptions.Collation)
	}
	if _, ok := rawFindOneAndUpdateOptions["hint"]; ok {
		opts = opts.SetHint(parsedFindOneAndUpdateOptions.Hint)
	}
	if _, ok := rawFindOneAndUpdateOptions["arrayFilters"]; ok {
		opts = opts.SetArrayFilters(options.ArrayFilters{Filters: parsedFindOneAndUpdateOptions.ArrayFilters})
	}
	if _, ok := rawFindOneAndUpdateOptions["upsert"]; ok {
		opts = opts.SetUpsert(parsedFindOneAndUpdateOptions.Upsert)
	}
	if _, ok := rawFindOneAndUpdateOptions["projection"]; ok {
		opts = opts.SetProjection(parsedFindOneAndUpdateOptions.Projection)
	}
	if _, ok := rawFindOneAndUpdateOptions["sort"]; ok {
		opts = opts.SetSort(parsedFindOneAndUpdateOptions.Sort)
	}
	if _, ok := rawFindOneAndUpdateOptions["returnDocument"]; ok {
		if parsedFindOneAndUpdateOptions.ReturnDocument == "after" {
			opts = opts.SetReturnDocument(options.After)
		} else if parsedFindOneAndUpdateOptions.ReturnDocument == "before" {
			opts = opts.SetReturnDocument(options.Before)
		}
	}

	var results bson.M
	if err := coll.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&results); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) insertOne() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)

	var iOOptions InsertOneContent
	if err := mapstructure.Decode(q.query.TypeContent, &iOOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	doc := bson.D{}
	if iOOptions.Document != "" && iOOptions.Document != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(iOOptions.Document), true, &doc); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	results, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) insertMany() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)

	var iMOptions InsertManyContent
	if err := mapstructure.Decode(q.query.TypeContent, &iMOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var documents []bson.D
	if iMOptions.Document != "" && iMOptions.Document != "[]" {
		if err := bson.UnmarshalExtJSON([]byte(iMOptions.Document), true, &documents); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	docs := make([]interface{}, 0, len(documents))
	for _, v := range documents {
		docs = append(docs, v)
	}

	results, err := coll.InsertMany(context.TODO(), docs)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) listCollections() (common.RuntimeResult, error) {
	db := q.client.Database(q.db)

	var lCOptions ListCollectionsContent
	if err := mapstructure.Decode(q.query.TypeContent, &lCOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if lCOptions.Query != "" && lCOptions.Query != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(lCOptions.Query), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	cursor, err := db.ListCollections(context.TODO(), filter)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) updateMany() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var uMOptions UpdateManyContent
	if err := mapstructure.Decode(q.query.TypeContent, &uMOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if uMOptions.Filter != "" && uMOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(uMOptions.Filter), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	update := bson.D{}
	if uMOptions.Update != "" && uMOptions.Update != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(uMOptions.Update), true, &update); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	// build `UpdateManyOptions`
	if uMOptions.Options == "" {
		uMOptions.Options = "{}"
	}
	var rawUpdateManyOptions map[string]interface{}
	if err := json.Unmarshal([]byte(uMOptions.Options), &rawUpdateManyOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var parsedUpdateManyOptions UpdateManyOptions
	if err := mapstructure.Decode(rawUpdateManyOptions, &parsedUpdateManyOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	opts := &options.UpdateOptions{}
	if _, ok := rawUpdateManyOptions["collation"]; ok {
		opts = opts.SetCollation(parsedUpdateManyOptions.Collation)
	}
	if _, ok := rawUpdateManyOptions["hint"]; ok {
		opts = opts.SetHint(parsedUpdateManyOptions.Hint)
	}
	if _, ok := rawUpdateManyOptions["arrayFilters"]; ok {
		opts = opts.SetArrayFilters(options.ArrayFilters{Filters: parsedUpdateManyOptions.ArrayFilters})
	}
	if _, ok := rawUpdateManyOptions["upsert"]; ok {
		opts = opts.SetUpsert(parsedUpdateManyOptions.Upsert)
	}

	results, err := coll.UpdateMany(context.TODO(), filter, update, opts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) updateOne() (common.RuntimeResult, error) {
	coll := q.client.Database(q.db).Collection(q.query.Collection)
	var uOOptions UpdateOneContent
	if err := mapstructure.Decode(q.query.TypeContent, &uOOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	filter := bson.D{}
	if uOOptions.Filter != "" && uOOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(uOOptions.Filter), true, &filter); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	update := bson.D{}
	if uOOptions.Filter != "" && uOOptions.Filter != "{}" {
		if err := bson.UnmarshalExtJSON([]byte(uOOptions.Update), true, &update); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
	}

	// build `UpdateOneOptions`
	if uOOptions.Options == "" {
		uOOptions.Options = "{}"
	}
	var rawUpdateOneOptions map[string]interface{}
	if err := json.Unmarshal([]byte(uOOptions.Options), &rawUpdateOneOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var parsedUpdateOneOptions UpdateOneOptions
	if err := mapstructure.Decode(rawUpdateOneOptions, &parsedUpdateOneOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	opts := &options.UpdateOptions{}
	if _, ok := rawUpdateOneOptions["collation"]; ok {
		opts = opts.SetCollation(parsedUpdateOneOptions.Collation)
	}
	if _, ok := rawUpdateOneOptions["hint"]; ok {
		opts = opts.SetHint(parsedUpdateOneOptions.Hint)
	}
	if _, ok := rawUpdateOneOptions["arrayFilters"]; ok {
		opts = opts.SetArrayFilters(options.ArrayFilters{Filters: parsedUpdateOneOptions.ArrayFilters})
	}
	if _, ok := rawUpdateOneOptions["upsert"]; ok {
		opts = opts.SetUpsert(parsedUpdateOneOptions.Upsert)
	}

	results, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}

func (q *QueryRunner) command() (common.RuntimeResult, error) {
	db := q.client.Database(q.db)

	var cmdOptions CommandContent
	if err := mapstructure.Decode(q.query.TypeContent, &cmdOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	var doc bson.D
	if err := bson.UnmarshalExtJSON([]byte(cmdOptions.Document), true, &doc); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	var results bson.M
	if err := db.RunCommand(context.TODO(), doc).Decode(&results); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{{"result": results}}}, nil
}
