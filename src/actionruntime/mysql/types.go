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
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
)

const (
	FIELD_CONTEXT = "context"
	FIELD_QUERY   = "query"
)

type MySQLOptions struct {
	Host             string `validate:"required"`
	Port             string `validate:"required"`
	DatabaseName     string `validate:"required"`
	DatabaseUsername string `validate:"required"`
	DatabasePassword string `validate:"required"`
	SSL              SSLOptions
}

type SSLOptions struct {
	SSL        bool
	ServerCert string `validate:"required_unless=SSL false"`
	ClientKey  string
	ClientCert string
}

type MySQLQuery struct {
	Mode     string `validate:"required,oneof=gui sql sql-safe"`
	Query    string
	RawQuery string
	Context  map[string]interface{}
}

func (q *MySQLQuery) IsSafeMode() bool {
	return q.Mode == common.MODE_SQL_SAFE
}

func (q *MySQLQuery) SetRawQueryAndContext(rawTemplate map[string]interface{}) error {
	queryRaw, hit := rawTemplate[FIELD_QUERY]
	if !hit {
		return errors.New("missing query field for SetRawQueryAndContext() in query")
	}
	queryAsserted, assertPass := queryRaw.(string)
	if !assertPass {
		return errors.New("query field assert failed in SetRawQueryAndContext() method")

	}
	q.RawQuery = queryAsserted
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
