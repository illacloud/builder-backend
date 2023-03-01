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

package model

import (
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const GENERATE_SQL_ACTION_SELECT = 1
const GENERATE_SQL_ACTION_INSERT = 2
const GENERATE_SQL_ACTION_UPDATE = 3
const GENERATE_SQL_ACTION_DELETE = 4

var ACTION_MAP = map[int]string{
	GENERATE_SQL_ACTION_SELECT: "SELECT",
	GENERATE_SQL_ACTION_INSERT: "INSERT",
	GENERATE_SQL_ACTION_UPDATE: "UPDATE",
	GENERATE_SQL_ACTION_DELETE: "DELETE",
}

type GenerateSQLRequest struct {
	Description string `json:"description" validate:"required"`
	ResourceID  string `json:"resourceID" validate:"required"`
	Action      int    `json:"action" validate:"required"`
}

func NewGenerateSQLRequest() *GenerateSQLRequest {
	return &GenerateSQLRequest{}
}

func (req *GenerateSQLRequest) GetActionInString() string {
	return ACTION_MAP[req.Action]
}

func (req *GenerateSQLRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}
