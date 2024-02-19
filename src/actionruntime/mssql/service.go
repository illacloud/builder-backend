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

package mssql

import (
	"context"
	"errors"
	"fmt"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_sql "github.com/illacloud/builder-backend/src/utils/parser/sql"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"

	"github.com/go-playground/validator/v10"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (m *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &m.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate mssql options
	validate := validator.New()
	if err := validate.Struct(m.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &m.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate mssql options
	validate := validator.New()
	if err := validate.Struct(m.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (m *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get Microsoft SQL Server connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close()

	// test Microsoft SQL Server connection
	if err := db.Ping(); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (m *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get Microsoft SQL Server connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test Microsoft SQL Server connection
	if err := db.Ping(); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get Microsoft SQL Server tables information
	columns := fieldsInfo(db, tablesInfo(db))

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (m *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get Microsoft SQL Server connection
	db, err := m.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get mssql connection")
	}
	defer db.Close()
	// format query
	if err := mapstructure.Decode(actionOptions, &m.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// set context field
	errInSetRawQuery := m.ActionOpts.SetRawQueryAndContext(rawActionOptions)
	if errInSetRawQuery != nil {
		return common.RuntimeResult{Success: false}, errInSetRawQuery
	}

	queryResult := common.RuntimeResult{Success: false}
	queryResult.Rows = make([]map[string]interface{}, 0, 0)
	queryResult.Extra = make(map[string]interface{})
	err = nil

	// start a default context
	ctx, cancel := context.WithTimeout(context.TODO(), common.DEFAULT_QUERY_AND_EXEC_TIMEOUT)
	defer cancel()

	// action mode switch
	switch m.ActionOpts.Mode {
	case ACTION_SQL_MODE:
		fallthrough
	case ACTION_SQL_SAFE_MODE:
		// check if m.Action.Query is select query
		sqlEscaper := parser_sql.NewSQLEscaper(resourcelist.TYPE_MSSQL_ID)
		escapedSQL, sqlArgs, errInEscapeSQL := sqlEscaper.EscapeSQLActionTemplate(m.ActionOpts.RawQuery, m.ActionOpts.Context, m.ActionOpts.IsSafeMode())
		if errInEscapeSQL != nil {
			return queryResult, errInEscapeSQL
		}
		// check if m.Action.Query["sql"] is select query
		isSelectQuery := false

		query, ok := m.ActionOpts.Query["sql"].(string)
		if !ok {
			return queryResult, errors.New("type error of action content")
		}
		lexer := parser_sql.NewLexer(query)
		isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}

		// fetch data
		if isSelectQuery && m.ActionOpts.IsSafeMode() {
			rows, err := db.QueryContext(ctx, escapedSQL, sqlArgs...)
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
		} else if isSelectQuery && !m.ActionOpts.IsSafeMode() {
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
		} else if !isSelectQuery && m.ActionOpts.IsSafeMode() {
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
		} else if !isSelectQuery && !m.ActionOpts.IsSafeMode() {
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
	case ACTION_GUI_MODE:
		// format data
		var guiQuery GUIQuery
		if err := mapstructure.Decode(m.ActionOpts.Query, &guiQuery); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		tableName := guiQuery.Table
		queryType := guiQuery.Type
		records := guiQuery.Records
		recordsLen := len(records)
		if queryType != ACTION_GUI_TYPE || recordsLen == 0 {
			return queryResult, errors.New("type error of action content")
		}
		record := records[0]
		columnNum := len(record)
		if columnNum == 0 {
			return queryResult, errors.New("type error of action content")
		}
		tableColumns := make([]string, 0, columnNum)
		for k := range record {
			tableColumns = append(tableColumns, k)
		}

		// begin transaction
		txn, err := db.Begin()
		if err != nil {
			return queryResult, err
		}
		// prepare statement
		stmt, err := txn.Prepare(mssql.CopyIn(tableName, mssql.BulkOptions{}, tableColumns...))
		if err != nil {
			return queryResult, err
		}
		// batch data load
		for i := 0; i < recordsLen; i++ {
			tableValues := make([]interface{}, 0, len(tableColumns))
			for _, tableColumn := range tableColumns {
				tableValues = append(tableValues, records[i][tableColumn])
			}
			_, err = stmt.ExecContext(ctx, tableValues...)
			if err != nil {
				stmt.Close()
				txn.Rollback()
				return queryResult, err
			}
		}
		// exec prepared statement with given batch data
		result, err := stmt.ExecContext(ctx)
		if err != nil {
			stmt.Close()
			txn.Rollback()
			return queryResult, err
		}
		// close prepared statement
		if err = stmt.Close(); err != nil {
			txn.Rollback()
			return queryResult, err
		}
		// transaction commit
		if err = txn.Commit(); err != nil {
			txn.Rollback()
			return queryResult, err
		}
		rowCount, err := result.RowsAffected()
		if err != nil {
			return queryResult, err
		}
		queryResult.Success = true
		queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", rowCount)
	default:
		err = errors.New("unsupported action mode")
	}

	return queryResult, err
}
