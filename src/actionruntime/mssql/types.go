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
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
)

const (
	FIELD_CONTEXT = "context"
	FIELD_QUERY   = "query"
	FIELD_SQL     = "sql"
)

type Resource struct {
	Host           string `validate:"required"`
	Port           string `validate:"required"`
	DatabaseName   string `validate:"required"`
	Username       string
	Password       string
	ConnectionOpts []map[string]string `validate:"required"`
	SSL            SSLOptions
}

type SSLOptions struct {
	SSL              bool
	CACert           string
	PrivateKey       string
	ClientCert       string
	VerificationMode string `validate:"required,oneof=full skip"`
}

type Action struct {
	Query    map[string]interface{} `validate:"required"`
	Mode     string                 `validate:"required,oneof=gui sql sql-safe"`
	RawQuery string
	Context  map[string]interface{}
}

func (q *Action) IsSafeMode() bool {
	return q.Mode == common.MODE_SQL_SAFE
}

type GUIQuery struct {
	Table   string
	Type    string
	Records []map[string]interface{}
}

func (q *Action) SetRawQueryAndContext(rawTemplate map[string]interface{}) error {
	queryRaw, hit := rawTemplate[FIELD_QUERY]
	if !hit {
		return errors.New("missing query field for SetRawQueryAndContext() in query")
	}
	queryAsserted, assertPass := queryRaw.(map[string]interface{})
	if !assertPass {
		return errors.New("query field assert failed in SetRawQueryAndContext() method")

	}
	sqlRaw, hitSQL := queryAsserted[FIELD_SQL]
	if !hitSQL {
		return errors.New("missing query.sql field for SetRawQueryAndContext() in query")
	}
	sqlAsserted, assertPassSQL := sqlRaw.(string)
	if !assertPassSQL {
		return errors.New("query.sql field assert failed in SetRawQueryAndContext() method")

	}

	q.RawQuery = sqlAsserted
	contextRaw, hit := rawTemplate[FIELD_CONTEXT]
	if !hit {
		return errors.New("missing context field SetRawQueryAndContext() in query")
	}
	contextAsserted, assertPass := contextRaw.(map[string]interface{})
	if !assertPass {
		return errors.New("context field assert failed in SetRawQueryAndContext() method")

	}
	q.Context = contextAsserted
	return nil
}
