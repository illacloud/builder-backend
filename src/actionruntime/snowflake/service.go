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

package snowflake

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_sql "github.com/illacloud/builder-backend/src/utils/parser/sql"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	resourceOptions Resource
	actionOptions   Action
}

func (s *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &s.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate snowflake options
	validate := validator.New()
	if err := validate.Struct(s.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &s.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate snowflake options
	validate := validator.New()
	if err := validate.Struct(s.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close()

	// test snowflake connection
	if err := db.Ping(); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (s *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test snowflake connection
	if err := db.Ping(); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := fieldsInfo(db, tablesInfo(db, fmt.Sprintf("%s.%s", s.resourceOptions.Database, s.resourceOptions.Schema)))

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (s *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get snowflake connection")
	}
	defer db.Close()

	// format query
	if err := mapstructure.Decode(actionOptions, &s.actionOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// set context field
	errInSetRawQuery := s.actionOptions.SetRawQueryAndContext(rawActionOptions)
	if errInSetRawQuery != nil {
		return common.RuntimeResult{Success: false}, errInSetRawQuery
	}

	// run clickhouse query
	queryResult := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// check if m.Action.Query is select query
	sqlEscaper := parser_sql.NewSQLEscaper(resourcelist.TYPE_SNOWFLAKE_ID)
	escapedSQL, sqlArgs, errInEscapeSQL := sqlEscaper.EscapeSQLActionTemplate(s.actionOptions.RawQuery, s.actionOptions.Context, s.actionOptions.IsSafeMode())
	if errInEscapeSQL != nil {
		return queryResult, errInEscapeSQL
	}

	// check if m.Action.Query is select query
	isSelectQuery := false

	lexer := parser_sql.NewLexer(s.actionOptions.Query)
	isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// start a default context
	ctx, cancel := context.WithTimeout(context.TODO(), common.DEFAULT_QUERY_AND_EXEC_TIMEOUT)
	defer cancel()

	// fetch data
	if isSelectQuery && s.actionOptions.IsSafeMode() {
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
	} else if isSelectQuery && !s.actionOptions.IsSafeMode() {
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
	} else if !isSelectQuery && s.actionOptions.IsSafeMode() {
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
	} else if !isSelectQuery && !s.actionOptions.IsSafeMode() {
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
