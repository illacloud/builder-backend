package illacloudperipheralapisdk

import (
	"errors"
	"fmt"

	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"
)

// field is:
// - database type, like "Postgres SQL"
// - table fields as: #{table_nbame}({column_a}, {column_b})
// - user description, like "A query to list the names of the departments which employed more than 10 employees in the last 3 months"
// - action: like "INSERT, UPDATE ... "
const GENERATE_SQL_DESCRIPTION_TEMPALTE = "### %s tables, with their properties: #%s #### %s %s"
const TABLE_DESC_CHARACTER_LIMIT = 1000

type GenerateSQLPeripheralRequest struct {
	Description   string `json:"description" validate:"required"`
	ValidateToken string `json:"validateToken" validate:"required"`
	SQLAction     string `json:"-"` // which is one of "SELECT", "INSERT", "UPDATE", "DELETE"
}

func (m *GenerateSQLPeripheralRequest) Export() map[string]string {
	payload := map[string]string{
		"description":   m.Description,
		"validateToken": m.ValidateToken,
	}
	return payload
}

func NewGenerateSQLPeripheralRequest(resourceType string, metaInfo interface{}, description string, sqlAction string) (*GenerateSQLPeripheralRequest, error) {
	// decode meta info
	metaInfoAsserted, assertMetaInfoPass := metaInfo.(map[string]interface{})
	if !assertMetaInfoPass {
		return nil, errors.New("invalied meta info")
	}

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
	for tableName, tableFieldsRaw := range metaInfoAsserted {
		tableFields, ok := tableFieldsRaw.(map[string]interface{})
		if !ok {
			return nil, errors.New("resource meta info do not include table name and table field info. please check your resource type.")
		}
		tableDesc := generateTableDescription(tableName, tableFields)
		allTableDesc += tableDesc
		if len(allTableDesc) > TABLE_DESC_CHARACTER_LIMIT {
			break
		}
	}
	prompt := fmt.Sprintf(GENERATE_SQL_DESCRIPTION_TEMPALTE, resourceType, allTableDesc, description, sqlAction)

	// generate validate token
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	token := tokenValidator.GenerateValidateToken(prompt)

	return &GenerateSQLPeripheralRequest{
		Description:   prompt,
		SQLAction:     sqlAction,
		ValidateToken: token,
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
