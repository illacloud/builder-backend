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

func (s *Connector) ValidateResourceOptions(resourceOpts map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOpts, &s.resourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate snowflake options
	validate := validator.New()
	if err := validate.Struct(s.resourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) ValidateActionOptions(actionOpts map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOpts, &s.actionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate snowflake options
	validate := validator.New()
	if err := validate.Struct(s.actionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (s *Connector) TestConnection(resourceOpts map[string]interface{}) (common.ConnectionResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOpts)
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

func (s *Connector) GetMetaInfo(resourceOpts map[string]interface{}) (common.MetaInfoResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOpts)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}
	defer db.Close()

	// test snowflake connection
	if err := db.Ping(); err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	columns := fieldsInfo(db, tablesInfo(db, fmt.Sprintf("%s.%s", s.resourceOpts.Database, s.resourceOpts.Schema)))

	return common.MetaInfoResult{
		Success: true,
		Schema:  columns,
	}, nil
}

func (s *Connector) Run(resourceOpts map[string]interface{}, actionOpts map[string]interface{}) (common.RuntimeResult, error) {
	// get snowflake connection
	db, err := s.getConnectionWithOptions(resourceOpts)
	if err != nil {
		return common.RuntimeResult{Success: false}, errors.New("failed to get snowflake connection")
	}
	defer db.Close()

	// format query
	if err := mapstructure.Decode(actionOpts, &s.actionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// run clickhouse query
	queryResult := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	// check if m.Action.Query is select query
	isSelectQuery := false

	lexer := parser_sql.NewLexer(s.actionOpts.Query)
	isSelectQuery, err = parser_sql.IsSelectSQL(lexer)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// fetch data
	if isSelectQuery {
		rows, err := db.Query(s.actionOpts.Query)
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
		execResult, err := db.Exec(s.actionOpts.Query)
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
