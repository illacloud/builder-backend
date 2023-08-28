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

package elasticsearch

const (
	SEARCH_OPERATION = "search"
	INSERT_OPERATION = "insert"
	GET_OPERATION    = "get"
	UPDATE_OPERATION = "update"
	DELETE_OPERATION = "delete"
)

type Resource struct {
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	Username string `validate:"required"`
	Password string `validate:"required"`
}

type Action struct {
	Operation string `validate:"required,oneof=search insert get update delete"`
	Index     string
	ID        string
	Body      string
	Query     string
}
