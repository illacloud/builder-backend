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
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"

	mssqldb "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
	"github.com/mitchellh/mapstructure"
)

const (
	VERIFY_FULL_MODE     = "full"
	SKIP_CA_MODE         = "skip"
	ACTION_SQL_MODE      = "sql"
	ACTION_SQL_SAFE_MODE = "sql-safe"
	ACTION_GUI_MODE      = "gui"
	ACTION_GUI_TYPE      = "bulk_insert"
	tableSQLStr          = "SELECT TABLE_NAME tableName, TABLE_SCHEMA tableSchema FROM INFORMATION_SCHEMA.TABLES;"
	columnSQLStr         = "SELECT COLUMN_NAME columnName, DATA_TYPE columnType FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2"
)

func (m *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*sql.DB, error) {
	if err := mapstructure.Decode(resourceOptions, &m.ResourceOpts); err != nil {
		return nil, err
	}
	escapedPassword := url.QueryEscape(m.ResourceOpts.Password)
	// build base Microsoft SQL Server connection string
	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&connection+timeout=120", m.ResourceOpts.Username, escapedPassword, m.ResourceOpts.Host, m.ResourceOpts.Port, m.ResourceOpts.DatabaseName)

	// append connection options
	for _, opt := range m.ResourceOpts.ConnectionOpts {
		if opt["key"] != "" {
			connString += fmt.Sprintf("&%s=%s", opt["key"], opt["value"])
		}
	}

	// add SSL/TLS verification parameters
	if m.ResourceOpts.SSL.SSL {
		if m.ResourceOpts.SSL.VerificationMode == VERIFY_FULL_MODE {
			if m.ResourceOpts.SSL.CACert == "" {
				return nil, errors.New("CA Cert required")
			} else {
				connString += "&encrypt=true&trustServerCertificate=false"
			}
		} else if m.ResourceOpts.SSL.VerificationMode == SKIP_CA_MODE {
			connString += "&encrypt=true&trustServerCertificate=true"
		} else {
			return nil, errors.New("unsupported verification mode")
		}
	}
	// parse connection string to msdsn.Config
	cfg, err := msdsn.Parse(connString)
	if err != nil {
		return nil, err
	}
	// add CA cert for tls.config when Verification mode is `full`
	if m.ResourceOpts.SSL.SSL && m.ResourceOpts.SSL.VerificationMode == VERIFY_FULL_MODE && m.ResourceOpts.SSL.CACert != "" {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(m.ResourceOpts.SSL.CACert)); !ok {
			return nil, errors.New("error parsing CA Cert")
		}
		cfg.TLSConfig.RootCAs = pool
		ccBlock, _ := pem.Decode([]byte(m.ResourceOpts.SSL.ClientCert))
		ckBlock, _ := pem.Decode([]byte(m.ResourceOpts.SSL.PrivateKey))
		if (ccBlock != nil && ccBlock.Type == "CERTIFICATE") && (ckBlock != nil || ckBlock.Type == "PRIVATE KEY") {
			cert, err := tls.X509KeyPair([]byte(m.ResourceOpts.SSL.ClientCert), []byte(m.ResourceOpts.SSL.PrivateKey))
			if err != nil {
				return nil, err
			}
			cfg.TLSConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// convert msdsn.Config to driver.Connector interface implemented by go-mssqldb
	conn := mssqldb.NewConnectorConfig(cfg)
	// connect to db
	db := sql.OpenDB(conn)

	return db, nil
}

func tablesInfo(db *sql.DB) []map[string]string {
	tableNames := make([]map[string]string, 0, 0)
	tableRows, err := db.Query(tableSQLStr)
	if err != nil {
		return nil
	}
	for tableRows.Next() {
		var tableName, tableSchema string
		err = tableRows.Scan(&tableName, &tableSchema)
		if err != nil {
			return nil
		}

		tableNames = append(tableNames, map[string]string{"table": tableName, "schema": tableSchema})
	}

	return tableNames
}

func fieldsInfo(db *sql.DB, tableNames []map[string]string) map[string]interface{} {
	columns := make(map[string]interface{})
	for _, tableName := range tableNames {
		columnRows, err := db.Query(columnSQLStr, tableName["schema"], tableName["table"])
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
		tableStr := fmt.Sprintf("%s.%s", tableName["schema"], tableName["table"])
		columns[tableStr] = tables
	}
	return columns
}
