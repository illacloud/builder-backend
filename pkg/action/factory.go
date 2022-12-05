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
	"github.com/illa-family/builder-backend/pkg/plugins/elasticsearch"
	"github.com/illa-family/builder-backend/pkg/plugins/mongodb"
	"github.com/illa-family/builder-backend/pkg/plugins/mysql"
	"github.com/illa-family/builder-backend/pkg/plugins/postgresql"
	"github.com/illa-family/builder-backend/pkg/plugins/redis"
	"github.com/illa-family/builder-backend/pkg/plugins/restapi"
	"github.com/illa-family/builder-backend/pkg/plugins/s3"
	"github.com/illa-family/builder-backend/pkg/plugins/smtp"
)

var (
	REST_ACTION          = "restapi"
	MYSQL_ACTION         = "mysql"
	MARIADB_ACTION       = "mariadb"
	TIDB_ACTION          = "tidb"
	TRANSFORMER_ACTION   = "transformer"
	POSTGRESQL_ACTION    = "postgresql"
	REDIS_ACTION         = "redis"
	MONGODB_ACTION       = "mongodb"
	ELASTICSEARCH_ACTION = "elasticsearch"
	S3_ACTION            = "s3"
	SMTP_ACTION          = "smtp"
	SUPABASEDB_ACTION    = "supabasedb"
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
	case MYSQL_ACTION, MARIADB_ACTION, TIDB_ACTION:
		sqlAction := &mysql.MySQLConnector{}
		return sqlAction
	case POSTGRESQL_ACTION, SUPABASEDB_ACTION:
		pgsAction := &postgresql.Connector{}
		return pgsAction
	case REDIS_ACTION:
		redisAction := &redis.Connector{}
		return redisAction
	case MONGODB_ACTION:
		mongoAction := &mongodb.Connector{}
		return mongoAction
	case ELASTICSEARCH_ACTION:
		esAction := &elasticsearch.Connector{}
		return esAction
	case S3_ACTION:
		s3Action := &s3.Connector{}
		return s3Action
	case SMTP_ACTION:
		smtpAction := &smtp.Connector{}
		return smtpAction
	default:
		return nil
	}
}
