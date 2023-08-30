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

import "go.mongodb.org/mongo-driver/mongo/options"

const (
	STANDARD_FORMAT    = "standard"
	DNSSEEDLIST_FORMAT = "mongodb+srv"
	GUI_OPTIONS        = "gui"
	URI_OPTIONS        = "uri"
)

var (
	CONNECTION_FORMAT = map[string]string{STANDARD_FORMAT: "mongodb", DNSSEEDLIST_FORMAT: "mongodb+srv"}
)

type Options struct {
	ConfigType    string                 `validate:"required,oneof=gui uri"`
	ConfigContent map[string]interface{} `validate:"required"`
	SSL           SSLOptions
}

type GUIOptions struct {
	Host             string `validate:"required"`
	ConnectionFormat string `validate:"required,oneof=standard mongodb+srv"`
	Port             string `validate:"required_unless=ConnectionFormat mongodb+srv"`
	DatabaseName     string
	DatabaseUsername string
	DatabasePassword string
}

type URIOptions struct {
	URI string `validate:"required"`
}

type SSLOptions struct {
	Open   bool
	Client string
	CA     string
}

type Query struct {
	ActionType  string `validate:"required"`
	Collection  string
	TypeContent map[string]interface{} `validate:"required"`
}

type AggregateContent struct {
	Aggregation string
	Options     string
}

type BulkWriteContent struct {
	Operations string
	Options    string
}

type CountContent struct {
	Query string
}

type DeleteManyContent struct {
	Filter string
}

type DeleteOneContent struct {
	Filter string
}

type DistinctContent struct {
	Query   string
	Field   string
	Options string
}

type FindContent struct {
	Query      string
	Projection string
	SortBy     string
	Limit      string
	Skip       string
}

type FindOneContent struct {
	Query      string
	Projection string
	Skip       string
}

type FindOneAndUpdateContent struct {
	Filter  string
	Update  string
	Options string
}

type InsertOneContent struct {
	Document string
}

type InsertManyContent struct {
	Document string
}

type ListCollectionsContent struct {
	Query string
}

type UpdateManyContent struct {
	Filter  string
	Update  string
	Options string
}

type UpdateOneContent struct {
	Filter  string
	Update  string
	Options string
}

type CommandContent struct {
	Document string
}

type AggregateOptions struct {
	Collation *options.Collation
	Hint      interface{}
	BatchSize int32
}

type DistinctOptions struct {
	Collation *options.Collation
}

type FindOneAndUpdateOptions struct {
	Collation      *options.Collation
	Hint           interface{}
	ArrayFilters   []interface{}
	Upsert         bool
	Projection     interface{}
	Sort           interface{}
	ReturnDocument string
}

type UpdateManyOptions struct {
	Collation    *options.Collation
	Hint         interface{}
	ArrayFilters []interface{}
	Upsert       bool
}

type UpdateOneOptions struct {
	Collation    *options.Collation
	Hint         interface{}
	ArrayFilters []interface{}
	Upsert       bool
}
