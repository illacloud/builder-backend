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
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
)

const (
	CONNECTION_SID       = "SID"
	CONNECTION_SERVICE   = "Service"
	ACTION_SQL_MODE      = "sql"
	ACTION_SQL_SAFE_MODE = "sql-safe"
	ACTION_GUI_MODE      = "gui"
	ACTION_GUI_TYPE      = "bulk_insert"

	columnsSQL = "SELECT tabs.table_name, tabs.tablespace_name, cols.column_name, cols.data_type FROM user_tables tabs JOIN user_tab_columns cols ON tabs.table_name = cols.table_name LEFT JOIN user_cons_columns col_cons ON cols.column_name = col_cons.column_name AND cols.table_name = col_cons.table_name WHERE tabs.tablespace_name IS NOT NULL"
)

func (o *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*sql.DB, error) {
	if err := mapstructure.Decode(resourceOptions, &o.resourceOptions); err != nil {
		return nil, err
	}

	// build connection string
	serviceName := ""
	urlopts := map[string]string{
		"SSL": "true",
	}
	if !o.resourceOptions.SSL {
		urlopts["SSL"] = "false"
	}
	if o.resourceOptions.Type == CONNECTION_SID {
		urlopts["SID"] = o.resourceOptions.Name
	} else if o.resourceOptions.Type == CONNECTION_SERVICE {
		serviceName = o.resourceOptions.Name
	}
	port, err := strconv.Atoi(o.resourceOptions.Port)
	if err != nil {
		return nil, err
	}
	databaseURL := go_ora.BuildUrl(o.resourceOptions.Host, port, serviceName, o.resourceOptions.Username, o.resourceOptions.Password, urlopts)

	db, err := sql.Open("oracle", databaseURL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func mapColumns(db *sql.DB) map[string]interface{} {
	columnRows, err := db.Query(columnsSQL)
	if err != nil {
		return nil
	}

	tables := make(map[string]map[string]map[string]string)

	for columnRows.Next() {
		var table, tablespace, column, columnType string
		err = columnRows.Scan(&table, &tablespace, &column, &columnType)
		if err != nil {
			return nil
		}
		tableStr := fmt.Sprintf("%s.%s", tablespace, table)
		columnMap := map[string]string{"data_type": columnType}
		if _, ok := tables[tableStr]; ok {
			tables[tableStr][column] = columnMap
			continue
		} else {
			tables[tableStr] = map[string]map[string]string{column: columnMap}
			continue
		}
	}

	res := make(map[string]interface{})
	b, _ := json.Marshal(&tables)
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}

	return res
}
