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

const (
	LIST_METHOD   = "list"
	CREATE_METHOD = "create"
	GET_METHOD    = "get"
	UPDATE_METHOD = "update"
	DELETE_METHOD = "delete"
)

type Resource struct {
	Host       string `validate:"required"`
	ProjectID  string `validate:"required"`
	DatabaseID string `validate:"required"`
	APIKey     string `validate:"required"`
}

type Action struct {
	Method string                 `validate:"required,oneof=list create get update delete"`
	Opts   map[string]interface{} `validate:"required"`
}

type ListOpts struct {
	CollectionID string
	Filter       []Filter
	OrderBy      []Order
	Limit        int
}

type Filter struct {
	Attribute string
	Operator  string
	Value     string
}

type Order struct {
	Attribute string
	Value     string
}

type BaseOpts struct {
	CollectionID string
	DocumentID   string
}

type WithDataOpts struct {
	CollectionID string
	DocumentID   string
	Data         map[string]interface{}
}
