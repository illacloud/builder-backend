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

package mysql

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_sql "github.com/illacloud/builder-backend/src/utils/parser/sql"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
	"github.com/mitchellh/mapstructure"
)

type MySQLConnector struct {
	Resource MySQLOptions
	Action   MySQLQuery
}

func (m *MySQLConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate mysql options
	validate := validator.New()
	if err := validate.Struct(m.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (m *MySQLConnector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format sql options
	if err := mapstructure.Decode(actionOptions, &m.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate mysql options
	validate := validator.New()
	if err := validate.Struct(m.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (m *MySQLConnector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get mysql connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close()

	// test mysql connection
	if err := db.Ping(); err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	return common.ConnectionResult{Success: true}, nil
}

func (m *MySQLConnector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get mysql connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test mysql connection
	if err := db.Ping(); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := fieldsInfo(db, m.Resource.DatabaseName, tablesInfo(db, m.Resource.DatabaseName))

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (m *MySQLConnector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get mysql connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get mysql connection")
	}
	defer db.Close()

	// format query
	if err := mapstructure.Decode(actionOptions, &m.Action); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// set context field
	errInSetRawQuery := m.Action.SetRawQueryAndContext(rawActionOptions)
	if errInSetRawQuery != nil {
		return common.RuntimeResult{Success: false}, errInSetRawQuery
	}

	// run mysql query
	queryResult := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	// check if m.Action.Query is select query
	sqlEscaper := parser_sql.NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, sqlArgs, errInEscapeSQL := sqlEscaper.EscapeSQLActionTemplate(m.Action.RawQuery, m.Action.Context, m.Action.IsSafeMode())
	if errInEscapeSQL != nil {
		return queryResult, errInEscapeSQL
	}
	isSelectQuery := false
	lexer := parser_sql.NewLexer(m.Action.Query)
	isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// start a default context
	ctx, cancel := context.WithTimeout(context.TODO(), common.DEFAULT_QUERY_AND_EXEC_TIMEOUT)
	defer cancel()

	// fetch data
	if isSelectQuery && m.Action.IsSafeMode() {
		log.Printf("[DUMP] db.QueryContext() sql: %s\n", escapedSQL)
		rows, err := db.QueryContext(ctx, escapedSQL, sqlArgs...)
		if err != nil {
			return queryResult, err
		}
		log.Printf("[DUMP] common.RetrieveToMap start.\n")
		mapRes, err := common.RetrieveToMap(rows)
		if err != nil {
			return queryResult, err
		}
		defer rows.Close()
		log.Printf("[DUMP] common.RetrieveToMap done.\n")
		queryResult.Success = true
		queryResult.Rows = mapRes
	} else if isSelectQuery && !m.Action.IsSafeMode() {
		rows, err := db.QueryContext(ctx, escapedSQL)
		if err != nil {
			return queryResult, err
		}
		mapRes, err := common.RetrieveToMap(rows)
		if err != nil {
			return queryResult, err
		}
		defer rows.Close()
		queryResult.Success = true
		queryResult.Rows = mapRes
	} else if !isSelectQuery && m.Action.IsSafeMode() {
		execResult, err := db.ExecContext(ctx, escapedSQL, sqlArgs...)
		if err != nil {
			return queryResult, err
		}
		affectedRows, err := execResult.RowsAffected()
		if err != nil {
			return queryResult, err
		}
		queryResult.Success = true
		queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", affectedRows)
	} else if !isSelectQuery && !m.Action.IsSafeMode() {
		execResult, err := db.ExecContext(ctx, escapedSQL)
		if err != nil {
			return queryResult, err
		}
		affectedRows, err := execResult.RowsAffected()
		if err != nil {
			return queryResult, err
		}
		queryResult.Success = true
		queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", affectedRows)
	}

	return queryResult, nil
}
