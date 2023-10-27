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

	"github.com/illacloud/builder-backend/src/actionruntime/common"
)

const (
	FIELD_CONTEXT = "context"
	FIELD_QUERY   = "query"
	FIELD_OPTS    = "opts"
	FIELD_RAW     = "raw"
)

type Resource struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Type     string `mapstructure:"connectionType" validate:"oneof=SID Service"`
	Name     string `mapstructure:"name"`
	SSL      bool   `mapstructure:"ssl"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Action struct {
	Mode     string                 `mapstructure:"mode" validate:"oneof=gui sql sql-safe"`
	Opts     map[string]interface{} `mapstructure:"opts"`
	RawQuery string
	Context  map[string]interface{}
}

func (q *Action) IsSafeMode() bool {
	return q.Mode == common.MODE_SQL_SAFE
}

func (q *Action) SetRawQueryAndContext(rawTemplate map[string]interface{}) error {
	optsRaw, hitOpts := rawTemplate[FIELD_OPTS]
	if !hitOpts {
		return errors.New("missing opts field for SetRawQueryAndContext() in query")
	}
	optsAsserted, assertOptsPass := optsRaw.(map[string]interface{})
	if !assertOptsPass {
		return errors.New("opts field assert failed in SetRawQueryAndContext() method")

	}
	rawRaw, hitRaw := optsAsserted[FIELD_RAW]
	if !hitRaw {
		return errors.New("missing opts.raw field for SetRawQueryAndContext() in query")
	}
	rawAsserted, assertRawPass := rawRaw.(string)
	if !assertRawPass {
		return errors.New("opts.araw field assert failed in SetRawQueryAndContext() method")

	}
	q.RawQuery = rawAsserted
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

type SQL struct {
	Raw string `mapstructure:"raw"`
}

type GUIBulkOpts struct {
	Table   string                   `mapstructure:"table"`
	Type    string                   `mapstructure:"actionType"`
	Records []map[string]interface{} `mapstructure:"records"`
	Key     string                   `mapstructure:"primaryKey"`
}
