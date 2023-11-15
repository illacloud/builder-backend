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

package oracle9i

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_sql "github.com/illacloud/builder-backend/src/utils/parser/sql"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
	go_ora_v1 "github.com/illacloud/go-ora-v1"
	"github.com/mitchellh/mapstructure"
)

const (
	DEFAULT_CONNECTION_TIMEOUT = time.Second * 60
)

type Connector struct {
	resourceOptions Resource
	actionOptions   Action
}

func (o *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &o.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate oracle options
	validate := validator.New()
	if err := validate.Struct(o.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (o *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &o.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate oracle options
	validate := validator.New()
	if err := validate.Struct(o.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (o *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get oracle connection
	db, err := o.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close()

	// test oracle connection
	connectCtx, connectCancel := context.WithTimeout(context.Background(), DEFAULT_CONNECTION_TIMEOUT)
	defer connectCancel()
	if err := db.Ping(connectCtx); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (o *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get oracle connection
	db, err := o.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test oracle connection
	connectCtx, connectCancel := context.WithTimeout(context.Background(), DEFAULT_CONNECTION_TIMEOUT)
	defer connectCancel()
	if err := db.Ping(connectCtx); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := mapColumns(db)

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (o *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get Oracle connection
	db, err := o.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get oracle connection")
	}
	defer db.Close()
	// format query
	if err := mapstructure.Decode(actionOptions, &o.actionOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// set context field
	errInSetRawQuery := o.actionOptions.SetRawQueryAndContext(rawActionOptions)
	if errInSetRawQuery != nil {
		return common.RuntimeResult{Success: false}, errInSetRawQuery
	}

	queryResult := common.RuntimeResult{Success: false}
	queryResult.Rows = make([]map[string]interface{}, 0, 0)
	queryResult.Extra = make(map[string]interface{})
	err = nil

	// action mode switch
	switch o.actionOptions.Mode {
	case ACTION_SQL_MODE:
		fallthrough
	case ACTION_SQL_SAFE_MODE:
		sqlEscaper := parser_sql.NewSQLEscaper(resourcelist.TYPE_ORACLE_9I_ID)
		escapedSQL, sqlArgs, errInEscapeSQL := sqlEscaper.EscapeSQLActionTemplate(o.actionOptions.RawQuery, o.actionOptions.Context, o.actionOptions.IsSafeMode())
		if errInEscapeSQL != nil {
			return queryResult, errInEscapeSQL
		}
		// check if o.actionOptions.Opts.Raw is select query
		isSelectQuery := false

		var query SQL
		if err := mapstructure.Decode(o.actionOptions.Opts, &query); err != nil {
			return queryResult, errors.New("type error of action content")
		}
		lexer := parser_sql.NewLexer(query.Raw)
		isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}

		// fetch data
		if isSelectQuery && o.actionOptions.IsSafeMode() {
			stmt, errInPrepare := db.Prepare(escapedSQL)
			defer stmt.Close()
			if errInPrepare != nil {
				return queryResult, errInPrepare
			}
			driverValues := ConvertSQlArgsToDriverValues(sqlArgs)
			rows, err := stmt.Query(driverValues)
			if err != nil {
				return queryResult, err
			}
			defer rows.Close()
			mapRes, err := common.RetrieveToMapByDriverRows(rows)
			if err != nil {
				return queryResult, err
			}
			queryResult.Success = true
			queryResult.Rows = mapRes
		} else if isSelectQuery && !o.actionOptions.IsSafeMode() {
			stmt := go_ora_v1.NewStmt(escapedSQL, db)
			defer stmt.Close()
			rows, err := stmt.Query(nil)
			if err != nil {
				return queryResult, err
			}
			defer rows.Close()
			mapRes, err := common.RetrieveToMapByDriverRows(rows)
			if err != nil {
				return queryResult, err
			}
			queryResult.Success = true
			queryResult.Rows = mapRes
		} else if !isSelectQuery && o.actionOptions.IsSafeMode() {
			stmt, errInPrepare := db.Prepare(escapedSQL)
			defer stmt.Close()
			if errInPrepare != nil {
				return queryResult, errInPrepare
			}
			driverValues := ConvertSQlArgsToDriverValues(sqlArgs)
			execResult, err := stmt.Exec(driverValues)
			if err != nil {
				return queryResult, err
			}
			affectedRows, err := execResult.RowsAffected()
			if err != nil {
				return queryResult, err
			}
			queryResult.Success = true
			queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", affectedRows)
		} else if !isSelectQuery && !o.actionOptions.IsSafeMode() {
			stmt := go_ora_v1.NewStmt(escapedSQL, db)
			defer stmt.Close()
			execResult, err := stmt.Exec(nil)
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
		err = errors.New("unsupported action mode")
	default:
		err = errors.New("unsupported action mode")
	}
	return queryResult, err
}

func ConvertSQlArgsToDriverValues(sqlArgs []interface{}) []driver.Value {
	ret := make([]driver.Value, 0)
	for _, value := range sqlArgs {
		ret = append(ret, value.(driver.Value))
	}
	return ret
}
