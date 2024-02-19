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

package postgresql

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_sql "github.com/illacloud/builder-backend/src/utils/parser/sql"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	Resource Options
	Action   Query
}

func (p *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &p.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate postgresql options
	validate := validator.New()
	if err := validate.Struct(p.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (p *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format sql options
	if err := mapstructure.Decode(actionOptions, &p.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate postgresql options
	validate := validator.New()
	if err := validate.Struct(p.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (p *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// get postgresql connection
	db, err := p.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close(context.Background())

	// test postgresql connection
	if err := db.Ping(context.Background()); err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	return common.ConnectionResult{Success: true}, nil
}

func (p *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get postgresql connection
	db, err := p.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close(context.Background())

	// test postgresql connection
	if err := db.Ping(context.Background()); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := fieldsInfo(db, "public", tablesInfo(db, "public"))

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (p *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get postgresql connection
	db, err := p.getConnectionWithOptions(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get postgresql connection")
	}
	defer db.Close(context.Background())

	fmt.Printf("[DUMP] Run.actionOptions: %+v\n", actionOptions)
	fmt.Printf("[DUMP] Run.rawActionOptions: %+v\n", rawActionOptions)

	// format query
	if err := mapstructure.Decode(actionOptions, &p.Action); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// set context field
	errInSetRawQuery := p.Action.SetRawQueryAndContext(rawActionOptions)
	if errInSetRawQuery != nil {
		return common.RuntimeResult{Success: false}, errInSetRawQuery
	}
	// run postgresql query
	queryResult := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	// check if m.Action.Query is select query
	sqlEscaper := parser_sql.NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, sqlArgs, errInEscapeSQL := sqlEscaper.EscapeSQLActionTemplate(p.Action.RawQuery, p.Action.Context, p.Action.IsSafeMode())
	if errInEscapeSQL != nil {
		return queryResult, errInEscapeSQL
	}
	isSelectQuery := false

	lexer := parser_sql.NewLexer(escapedSQL)
	isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	fmt.Printf("[DUMP] escapedSQL: %s\n", escapedSQL)

	// start a default context
	ctx, cancel := context.WithTimeout(context.TODO(), common.DEFAULT_QUERY_AND_EXEC_TIMEOUT)
	defer cancel()

	// fetch data
	if isSelectQuery && p.Action.IsSafeMode() {
		log.Printf("[DUMP] db.Query() sql: %s\n", escapedSQL)
		rows, err := db.Query(ctx, escapedSQL, sqlArgs...)
		if err != nil {
			return queryResult, err
		}
		log.Printf("[DUMP] common.RetrieveToMap start.\n")
		mapRes, err := RetrieveToMap(rows)
		if err != nil {
			return queryResult, err
		}
		defer rows.Close()
		log.Printf("[DUMP] common.RetrieveToMap end.\n")

		queryResult.Success = true
		queryResult.Rows = mapRes
	} else if isSelectQuery && !p.Action.IsSafeMode() {
		rows, err := db.Query(ctx, escapedSQL)
		if err != nil {
			return queryResult, err
		}
		mapRes, err := RetrieveToMap(rows)
		if err != nil {
			return queryResult, err
		}
		defer rows.Close()
		queryResult.Success = true
		queryResult.Rows = mapRes
	} else if !isSelectQuery && p.Action.IsSafeMode() { // update, insert, delete data
		execResult, err := db.Exec(ctx, escapedSQL, sqlArgs...)
		if err != nil {
			return queryResult, err
		}
		affectedRows := execResult.RowsAffected()
		queryResult.Success = true
		queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", affectedRows)
	} else if !isSelectQuery && !p.Action.IsSafeMode() {
		execResult, err := db.Exec(ctx, escapedSQL)
		if err != nil {
			return queryResult, err
		}
		affectedRows := execResult.RowsAffected()
		queryResult.Success = true
		queryResult.Extra["message"] = fmt.Sprintf("Affeted %d rows.", affectedRows)
	}

	return queryResult, nil
}
