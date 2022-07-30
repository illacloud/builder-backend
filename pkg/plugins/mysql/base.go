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
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
)

const (
	tableSQLStr  = "SELECT TABLE_NAME tableName FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?"
	columnSQLStr = "SELECT COLUMN_NAME columnName, DATA_TYPE columnType FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
)

func (m *MySQLConnector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*sql.DB, error) {
	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
		return nil, err
	}
	var db *sql.DB
	var err error
	// TODO: connect via ssh or ssl
	db, err = m.connectPure()
	return db, err
}

func (m *MySQLConnector) connectPure() (db *sql.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.Resource.DatabaseUsername,
		m.Resource.DatabasePassword, m.Resource.Host, m.Resource.Port, m.Resource.DatabaseName)
	db, err = sql.Open("mysql", dsn+"?timeout=5s")
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (m *MySQLConnector) connectViaSSH() (db *sql.DB, err error) {
	// TODO: implement
	return nil, errors.New("inaccessible")
}

func (m *MySQLConnector) connectViaSSL() (db *sql.DB, err error) {
	// TODO: implement
	return nil, errors.New("inaccessible")
}

func tablesInfo(db *sql.DB, dbName string) []string {
	tableNames := make([]string, 0, 0)
	tableRows, err := db.Query(tableSQLStr, dbName)
	if err != nil {
		return nil
	}
	for tableRows.Next() {
		var tableName string
		err = tableRows.Scan(&tableName)
		if err != nil {
			return nil
		}

		tableNames = append(tableNames, tableName)
	}

	return tableNames
}

func fieldsInfo(db *sql.DB, dbName string, tableNames []string) map[string]interface{} {
	columns := make(map[string]interface{})
	for _, tableName := range tableNames {
		columnRows, err := db.Query(columnSQLStr, dbName, tableName)
		if err != nil {
			return nil
		}
		tables := make(map[string]interface{})

		for columnRows.Next() {
			var columnName, columnType string
			err = columnRows.Scan(&columnName, &columnType)
			if err != nil {
				return nil
			}
			tables[columnName] = map[string]string{"data_type": columnType}

		}
		columns[tableName] = tables
	}
	return columns
}
