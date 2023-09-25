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

package airtable

const (
	AIRTABLE_API = "https://api.airtable.com/v0/{baseId}/{tableName}"

	PERSONAL_TOKEN_AUTHENTICATION = "personalToken"
	TOKEN_AUTHENTICATION          = "token"
	API_KEY_AUTHENTICATION        = "apiKey"

	LIST_METHOD       = "list"
	GET_METHOD        = "get"
	CREATE_METHOD     = "create"
	UPDATE_METHOD     = "update"
	BULKUPDATE_METHOD = "bulkUpdate"
	DELETE_METHOD     = "delete"
	BULKDELETE_METHOD = "bulkDelete"

	JSON_CELL_FORMAT   = "json"
	STRING_CELL_FORMAT = "string"
)

type Resource struct {
	AuthenticationType   string            `mapstructure:"authenticationType" validate:"oneof=personalToken apiKey"`
	AuthenticationConfig map[string]string `mapstructure:"authenticationConfig" validate:"required"`
}

type Action struct {
	Method     string                 `mapstructure:"method" validate:"oneof=list get create update bulkUpdate delete bulkDelete"`
	BaseConfig BaseConfig             `mapstructure:"baseConfig"`
	Config     map[string]interface{} `mapstructure:"config" validate:"required"`
}

type BaseConfig struct {
	BaseID    string `mapstructure:"baseId"`
	TableName string `mapstructure:"tableName"`
}

type ListConfig struct {
	Fields          []string     `mapstructure:"fields"`
	FilterByFormula string       `mapstructure:"filterByFormula"`
	MaxRecords      int          `mapstructure:"maxRecords"`
	PageSize        int          `mapstructure:"pageSize"`
	Sort            []SortObject `mapstructure:"sort"`
	View            string       `mapstructure:"view"`
	CellFormat      string       `mapstructure:"cellFormat"`
	TimeZone        string       `mapstructure:"timeZone"`
	UserLocale      string       `mapstructure:"userLocale"`
	Offset          string       `mapstructure:"offset"`
}

type SortObject struct {
	Field     string `mapstructure:"field"`
	Direction string `mapstructure:"direction" validate:"oneof=asc desc"`
}

type GetConfig struct {
	RecordID string `mapstructure:"recordId" validate:"required"`
}

type CreateConfig struct {
	Records []map[string]interface{} `mapstructure:"records" validate:"required,gt=0,lt=11"`
}

type BulkUpdateConfig struct {
	Records []map[string]interface{} `mapstructure:"records" validate:"required,gt=0,lt=11"`
}

type UpdateConfig struct {
	RecordID string                 `mapstructure:"recordId" validate:"required"`
	Record   map[string]interface{} `mapstructure:"record" validate:"required"`
}

type DeleteConfig struct {
	RecordID string `mapstructure:"recordId" validate:"required"`
}

type BulkDeleteConfig struct {
	RecordIDs []string `mapstructure:"recordIds" validate:"required,gt=0,lt=11"`
}
