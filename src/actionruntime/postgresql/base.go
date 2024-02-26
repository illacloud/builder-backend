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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"

	"github.com/DmitriyVTitov/size"
	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/jackc/pgx/v5"
	"github.com/mitchellh/mapstructure"
)

const (
	tableSQLStr  = "SELECT TABLE_NAME tableName FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = $1;"
	columnSQLStr = "SELECT COLUMN_NAME columnName, DATA_TYPE columnType FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = $1 AND TABLE_NAME = $2;"
)

func (p *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*pgx.Conn, error) {
	if err := mapstructure.Decode(resourceOptions, &p.Resource); err != nil {
		return nil, err
	}
	var db *pgx.Conn
	var err error
	if p.Resource.SSL.SSL == true {
		db, err = p.connectViaSSL()
	} else {
		db, err = p.connectPure()
	}
	return db, err
}

func (p *Connector) connectPure() (db *pgx.Conn, err error) {
	// @NOTE: following this issue: https://github.com/jackc/pgx/issues/1285
	// the postgres connection string must be escaped in password
	escapedPassword := url.QueryEscape(p.Resource.DatabasePassword)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", p.Resource.DatabaseUsername,
		escapedPassword, p.Resource.Host, p.Resource.Port, p.Resource.DatabaseName)
	pgCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db, err = pgx.ConnectConfig(context.Background(), pgCfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (p *Connector) connectViaSSL() (db *pgx.Conn, err error) {
	// @NOTE: following this issue: https://github.com/jackc/pgx/issues/1285
	// the postgres connection string must be escaped in password
	escapedPassword := url.QueryEscape(p.Resource.DatabasePassword)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", p.Resource.DatabaseUsername,
		escapedPassword, p.Resource.Host, p.Resource.Port, p.Resource.DatabaseName)
	pgCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(p.Resource.SSL.ServerCert)); !ok {
		return nil, errors.New("PostgreSQL SSL/TLS Connection failed")
	}
	tlsConfig := tls.Config{RootCAs: pool, ServerName: p.Resource.Host}
	ccBlock, _ := pem.Decode([]byte(p.Resource.SSL.ClientCert))
	ckBlock, _ := pem.Decode([]byte(p.Resource.SSL.ClientKey))
	if (ccBlock != nil && ccBlock.Type == "CERTIFICATE") && (ckBlock != nil || ckBlock.Type == "PRIVATE KEY") {
		cert, err := tls.X509KeyPair([]byte(p.Resource.SSL.ClientCert), []byte(p.Resource.SSL.ClientKey))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	pgCfg.Config.TLSConfig = &tlsConfig

	db, err = pgx.ConnectConfig(context.Background(), pgCfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func tablesInfo(db *pgx.Conn, tableSchema string) []string {
	tableNames := make([]string, 0, 0)
	tableRows, err := db.Query(context.Background(), tableSQLStr, tableSchema)
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

func fieldsInfo(db *pgx.Conn, tableSchema string, tableNames []string) map[string]interface{} {
	columns := make(map[string]interface{})
	for _, tableName := range tableNames {
		columnRows, err := db.Query(context.Background(), columnSQLStr, tableSchema, tableName)
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

func RetrieveToMap(rows pgx.Rows) ([]map[string]interface{}, error) {
	fieldDescriptions := rows.FieldDescriptions()
	renamedColumns := make([]string, 0)
	columnNameHitMap := make(map[string]int, 0)
	columnNamePosMap := make(map[string]int, 0)

	for pos, col := range fieldDescriptions {
		hitColumnTimes, hitColumn := columnNameHitMap[col.Name]
		cloName := col.Name
		if hitColumn {
			cloName += fmt.Sprintf("_%d", hitColumnTimes)
			// rewrite first column to "_0"
			if columnNameHitMap[col.Name] == 1 {
				firstHitPos := columnNamePosMap[col.Name]
				renamedColumns[firstHitPos] += "_0"
			}
			columnNameHitMap[col.Name]++
		}
		columnNameHitMap[cloName] = 1
		columnNamePosMap[cloName] = pos
		renamedColumns = append(renamedColumns, cloName)
	}

	count := len(renamedColumns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	iteratorNums := 0
	tableDataCapacity := 10000
	for rows.Next() {
		iteratorNums++
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})

		for i, col := range renamedColumns {
			// uuid
			if values[i] != nil && reflect.TypeOf(values[i]).String() == "[16]uint8" {
				byteArray, _ := values[i].([16]uint8)
				tmp, _ := uuid.FromBytes(byteArray[:])
				entry[col] = tmp.String()
				continue
			}

			val := values[i]
			entry[col] = val
		}

		tableData = append(tableData, entry)

		// check tableData size by sample
		if iteratorNums == common.SQL_RESULT_MEMORY_CHECK_SAMPLE {
			tableDataSizeBySample := size.Of(tableData)
			tableDataCapacity = (common.SQL_RESULT_MEMORY_LIMIT / tableDataSizeBySample) * common.SQL_RESULT_MEMORY_CHECK_SAMPLE
		}
		if iteratorNums > tableDataCapacity {
			log.Printf("[ERROR] RetrieveToMap result exceeds 20MiB by iteratorNums: %d, size: %d", iteratorNums, size.Of(tableData))
			return nil, errors.New("returned result exceeds 20MiB, please adjust the query limit to reduce the number of results")
		}
	}

	return tableData, nil
}
