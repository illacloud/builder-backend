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

package clickhouse

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/mitchellh/mapstructure"
)

const (
	tableSQLStr  = "SHOW TABLES"
	columnSQLStr = "DESCRIBE TABLE "
)

func (c *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*sql.DB, error) {
	if err := mapstructure.Decode(resourceOptions, &c.ResourceOpts); err != nil {
		return nil, err
	}

	opts := clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", c.ResourceOpts.Host, c.ResourceOpts.Port)},
		Auth: clickhouse.Auth{
			Database: c.ResourceOpts.DatabaseName,
			Username: c.ResourceOpts.Username,
			Password: c.ResourceOpts.Password,
		},
	}

	if c.ResourceOpts.SSL.SSL {
		t := &tls.Config{InsecureSkipVerify: false}
		opts.TLS = t
	}
	if c.ResourceOpts.SSL.SSL && c.ResourceOpts.SSL.SelfSigned {
		t := &tls.Config{}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(c.ResourceOpts.SSL.CACert)); !ok {
			return nil, errors.New("clickhouse SSL/TLS Connection failed")
		}
		t.RootCAs = pool
		ccBlock, _ := pem.Decode([]byte(c.ResourceOpts.SSL.ClientCert))
		ckBlock, _ := pem.Decode([]byte(c.ResourceOpts.SSL.PrivateKey))
		if (ccBlock != nil && ccBlock.Type == "CERTIFICATE") && (ckBlock != nil || ckBlock.Type == "PRIVATE KEY") {
			cert, err := tls.X509KeyPair([]byte(c.ResourceOpts.SSL.ClientCert), []byte(c.ResourceOpts.SSL.PrivateKey))
			if err != nil {
				return nil, err
			}
			t.Certificates = []tls.Certificate{cert}
		}
		opts.TLS = t
	}

	db := clickhouse.OpenDB(&opts)

	return db, nil
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
		tmpSQLStr := columnSQLStr + tableName
		columnRows, err := db.Query(tmpSQLStr)
		if err != nil {
			return nil
		}
		tables := make(map[string]interface{})
		for columnRows.Next() {
			var Name, Type, defaultType, defaultExpression, ttlExpression, comment, codecExpression string
			err = columnRows.Scan(&Name, &Type, &defaultType, &defaultExpression, &ttlExpression, &comment, &codecExpression)
			if err != nil {
				return nil
			}
			tables[Name] = map[string]string{"data_type": Type}

		}
		columns[tableName] = tables
	}
	return columns
}
