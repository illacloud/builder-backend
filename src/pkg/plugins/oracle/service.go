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

package oracle

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	parser_sql "github.com/illacloud/builder-backend/internal/parser/sql"
	"github.com/illacloud/builder-backend/pkg/plugins/common"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	resourceOpts Resource
	actionOpts   Action
}

func (o *Connector) ValidateResourceOptions(resourceOpts map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOpts, &o.resourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate oracle options
	validate := validator.New()
	if err := validate.Struct(o.resourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (o *Connector) ValidateActionOptions(actionOpts map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOpts, &o.actionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate oracle options
	validate := validator.New()
	if err := validate.Struct(o.actionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (o *Connector) TestConnection(resourceOpts map[string]interface{}) (common.ConnectionResult, error) {
	// get oracle connection
	db, err := o.getConnectionWithOptions(resourceOpts)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}
	defer db.Close()

	// test oracle connection
	if err := db.Ping(); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	return common.ConnectionResult{Success: true}, nil
}

func (o *Connector) GetMetaInfo(resourceOpts map[string]interface{}) (common.MetaInfoResult, error) {
	// get oracle connection
	db, err := o.getConnectionWithOptions(resourceOpts)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test oracle connection
	if err := db.Ping(); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := mapColumns(db)

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (o *Connector) Run(resourceOpts map[string]interface{}, actionOpts map[string]interface{}) (common.RuntimeResult, error) {
	// get Oracle connection
	db, err := o.getConnectionWithOptions(resourceOpts)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get clickhouse connection")
	}
	defer db.Close()
	// format query
	if err := mapstructure.Decode(actionOpts, &o.actionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	queryResult := common.RuntimeResult{Success: false}
	queryResult.Rows = make([]map[string]interface{}, 0, 0)
	queryResult.Extra = make(map[string]interface{})
	err = nil

	// action mode switch
	switch o.actionOpts.Mode {
	case ACTION_SQL_MODE:
		// check if o.actionOpts.Opts.Raw is select query
		isSelectQuery := false

		var query SQL
		if err := mapstructure.Decode(o.actionOpts.Opts, &query); err != nil {
			return queryResult, errors.New("type error of action content")
		}
		lexer := parser_sql.NewLexer(query.Raw)
		isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}

		// fetch data
		if isSelectQuery {
			rows, err := db.Query(query.Raw)
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
		} else { // update, insert, delete data
			execResult, err := db.Exec(query.Raw)
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
