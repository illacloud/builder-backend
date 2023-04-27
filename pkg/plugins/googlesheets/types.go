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

package googlesheets

type Resource struct {
	Authentication string                 `validate:"required,oneof=serviceAccount oauth2"`
	Opts           map[string]interface{} `validate:"required"`
}

type SAOpts struct {
	PrivateKey string `validate:"required"`
}

type OAuth2Opts struct {
	AccessType   string `validate:"required,oneof=rw r"`
	AccessToken  string
	TokenType    string
	RefreshToken string
	Status       int
}

type Action struct {
	Method string                 `validate:"required,oneof=read append update bulkUpdate delete create copy list get"`
	Opts   map[string]interface{} `validate:"required_unless=Method list"`
}

type ReadOpts struct {
	Spreadsheet string `validate:"required"`
	SheetName   string
	Limit       int
	Offset      int
	RangeType   string `validate:"required,oneof=a1 limit"`
	A1Notation  string
}

type AppendOpts struct {
	Spreadsheet string `validate:"required"`
	SheetName   string
	Values      []map[string]interface{}
}

type UpdateOpts struct {
	Spreadsheet string `validate:"required"`
	FilterType  string `validate:"required,oneof=a1 filter"`
	A1Notation  string
	Values      []map[string]interface{}
	SheetName   string
	Filters     []Filter
}

type Filter struct {
	Key      string
	Operator string
	Value    string
}

type BulkUpdateOpts struct {
	Spreadsheet string `validate:"required"`
	SheetName   string
	PrimaryKey  string `validate:"required"`
	RowsArray   []map[string]interface{}
}

type DeleteOpts struct {
	Spreadsheet string `validate:"required"`
	SheetName   string
	RowIndex    int
}

type CreateOpts struct {
	Title string `validate:"required"`
}

type CopyOpts struct {
	Spreadsheet   string `validate:"required"`
	SheetName     string
	ToSpreadsheet string `validate:"required"`
	ToSheet       string
}

type GetOpts struct {
	Spreadsheet string `validate:"required"`
}
