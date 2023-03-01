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

package illacloudservice

import (
	"errors"
	"fmt"
)

// field is:
// - database type, like "Postgres SQL"
// - table fields as: #{table_nbame}({column_a}, {column_b})
// - user description, like "A query to list the names of the departments which employed more than 10 employees in the last 3 months"
// - action: like "INSERT, UPDATE ... "
const GENERATE_SQL_DESCRIPTION_TEMPALTE = "### %s tables, with their properties: #%s #### %s %s"

type GenerateSQLPeripheralRequest struct {
	Description   string `json:"description" validate:"required"`
	ValidateToken string `json:"validateToken" validate:"required"`
}

func (m *GenerateSQLPeripheralRequest) Export() map[string]string {
	payload := map[string]string{
		"description":   m.Description,
		"validateToken": m.ValidateToken,
	}
	return payload
}

func NewGenerateSQLPeripheralRequest(resourceType string, metaInfo map[string]interface{}, req *GenerateSQLRequest) (*GenerateSQLPeripheralRequest, error) {
	// generate meta info, the meta info like:
	// {
	//     "resourceName": "mssqlExample",
	//     "schema": {
	//         "dbo.MSreplication_options": { // table name
	//             "install_failures": { // field name
	//                 "data_type": "int"
	//             },
	//             "major_version": {
	//                 "data_type": "int"
	//             },
	//             "minor_version": {
	//                 "data_type": "int"
	//             },
	//             "optname": {
	//                 "data_type": "nvarchar"
	//             },
	//             "revision": {
	//                 "data_type": "int"
	//             },
	//             "value": {
	//                 "data_type": "bit"
	//             }
	//         },
	//         "dbo.spt_fallback_db": {
	//             "dbid": {
	//                 "data_type": "smallint"
	//             },
	//             "name": {
	//                 "data_type": "varchar"
	//             },
	//             "status": {
	//                 "data_type": "smallint"
	//             },
	//             "version": {
	//                 "data_type": "smallint"
	//             },
	//             "xdttm_ins": {
	//                 "data_type": "datetime"
	//             },
	//             "xdttm_last_ins_upd": {
	//                 "data_type": "datetime"
	//             },
	//             "xfallback_dbid": {
	//                 "data_type": "smallint"
	//             },
	//             "xserver_name": {
	//                 "data_type": "varchar"
	//             }
	//         },
	// 		...
	allTableDesc := ""
	for k1, v1 := range metaInfo {
		if k1 == "schema" {
			tableInfo, ok := v1.(map[string]interface{})
			if !ok {
				return nil, errors.New("resource meta info do not include table name and table field info. please check your resource type.")
			}
			for tableName, tableFieldsRaw := range tableInfo {
				tableFields, ok := tableFieldsRaw.(map[string]interface{})
				if !ok {
					return nil, errors.New("resource meta info do not include table name and table field info. please check your resource type.")
				}
				tableDesc := generateTableDescription(tableName, tableFields)
				allTableDesc += tableDesc
			}
		}
	}
	description := fmt.Sprintf(GENERATE_SQL_DESCRIPTION_TEMPALTE, resourceType, allTableDesc, req.Description, req.GetActionInString())
	return &GenerateSQLPeripheralRequest{
		Description: description,
	}, nil
}

func (m *GenerateSQLPeripheralRequest) SetValidateToken(token string) {
	m.ValidateToken = token
}

func generateTableDescription(tableName string, tableFields map[string]interface{}) string {
	tableDesc := "# " + tableName + "("
	tflen := len(tableFields)
	for field, _ := range tableFields {
		tflen--
		if tflen != 0 {
			tableDesc += field + ", "
		} else {
			tableDesc += field + " "
		}
	}
	tableDesc += ")"
	return tableDesc
}
