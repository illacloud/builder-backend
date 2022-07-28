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

package action

import (
	"github.com/illa-family/builder-backend/pkg/plugins/common"
	"github.com/illa-family/builder-backend/pkg/plugins/mysql"
	"github.com/illa-family/builder-backend/pkg/plugins/restapi"
)

var (
	REST_ACTION        = "restapi"
	MYSQL_ACTION       = "mysql"
	TRANSFORMER_ACTION = "transformer"
)

type AbstractActionFactory interface {
	Build() common.DataConnector
}

type Factory struct {
	Type string
}

func (f *Factory) Build() common.DataConnector {
	switch f.Type {
	case REST_ACTION:
		restapiAction := &restapi.RESTAPIConnector{}
		return restapiAction
	case MYSQL_ACTION:
		sqlAction := &mysql.MySQLConnector{}
		return sqlAction
	default:
		return nil
	}
}
