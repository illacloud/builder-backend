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

package common

const (
	MODE_GUI      = "gui"
	MODE_SQL      = "sql"
	MODE_SQL_SAFE = "sql-safe"
)

type ValidateResult struct {
	Valid bool
	Extra map[string]interface{}
}

type ConnectionResult struct {
	Success bool
}

type RuntimeResult struct {
	Success bool
	Rows    []map[string]interface{}
	Extra   map[string]interface{}
}

func (i *RuntimeResult) SetSuccess() {
	i.Success = true
}

type MetaInfoResult struct {
	Success bool
	Schema  map[string]interface{}
}

func (metaInfoResult *MetaInfoResult) ExportSchema() map[string]interface{} {
	return metaInfoResult.Schema
}
