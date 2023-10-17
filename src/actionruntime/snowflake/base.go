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
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	sf "github.com/snowflakedb/gosnowflake"
)

const (
	BASIC_AUTH    = "basic"
	KEY_PAIR_AUTH = "key"

	tableSQLStr  = "SHOW TERSE TABLES IN SCHEMA "
	columnSQLStr = "DESCRIBE TABLE "
)

func (s *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*sql.DB, error) {
	if err := mapstructure.Decode(resourceOptions, &s.resourceOptions); err != nil {
		return nil, err
	}

	config := sf.Config{
		Account:   s.resourceOptions.AccountName,
		Database:  s.resourceOptions.Database,
		Schema:    s.resourceOptions.Schema,
		Warehouse: s.resourceOptions.Warehouse,
		Role:      s.resourceOptions.Role,
	}

	switch s.resourceOptions.Authentication {
	case BASIC_AUTH:
		config.User = s.resourceOptions.AuthContent["username"]
		config.Password = s.resourceOptions.AuthContent["password"]
	case KEY_PAIR_AUTH:
		block, _ := pem.Decode([]byte(s.resourceOptions.AuthContent["privateKey"]))
		if block == nil {
			return nil, errors.New("failed to parse PEM block containing the private key")
		}
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("failed to parse the private key")
		}
		config.User = s.resourceOptions.AuthContent["username"]
		config.Authenticator = sf.AuthTypeJwt
		config.PrivateKey = rsaPrivateKey
	default:
		return nil, errors.New("unsupported authentication method")
	}

	dsn, err := sf.DSN(&config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func tablesInfo(db *sql.DB, dbName string) []map[string]string {
	tableNames := make([]map[string]string, 0, 0)
	queryStr := tableSQLStr + dbName
	tableRows, err := db.Query(queryStr)
	defer tableRows.Close()
	if err != nil {
		return nil
	}
	for tableRows.Next() {
		var createdOn interface{}
		var name, kind, databaseName, schemaName string
		err = tableRows.Scan(&createdOn, &name, &kind, &databaseName, &schemaName)
		if err != nil {
			return nil
		}

		tableNames = append(tableNames, map[string]string{"table": name, "schema": schemaName})
	}

	return tableNames
}

func fieldsInfo(db *sql.DB, tableNames []map[string]string) map[string]interface{} {
	columns := make(map[string]interface{})
	for _, tableName := range tableNames {
		queryStr := columnSQLStr + fmt.Sprintf("%s.%s", tableName["schema"], tableName["table"])
		columnRows, err := db.Query(queryStr)
		if err != nil {
			return nil
		}
		tables := make(map[string]interface{})

		for columnRows.Next() {
			var columnName, columnType string
			var c3, c4, c5, c6, c7, c8, c9, c10, c11 interface{}
			err = columnRows.Scan(&columnName, &columnType, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10, &c11)
			if err != nil {
				return nil
			}
			tables[columnName] = map[string]string{"data_type": columnType}

		}
		tableStr := fmt.Sprintf("%s.%s", tableName["schema"], tableName["table"])
		columns[tableStr] = tables
	}

	return columns
}
