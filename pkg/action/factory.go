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
	"github.com/illacloud/builder-backend/pkg/plugins/clickhouse"
	"github.com/illacloud/builder-backend/pkg/plugins/common"
	"github.com/illacloud/builder-backend/pkg/plugins/couchdb"
	"github.com/illacloud/builder-backend/pkg/plugins/dynamodb"
	"github.com/illacloud/builder-backend/pkg/plugins/elasticsearch"
	"github.com/illacloud/builder-backend/pkg/plugins/firebase"
	"github.com/illacloud/builder-backend/pkg/plugins/graphql"
	"github.com/illacloud/builder-backend/pkg/plugins/hfendpoint"
	"github.com/illacloud/builder-backend/pkg/plugins/huggingface"
	"github.com/illacloud/builder-backend/pkg/plugins/mongodb"
	"github.com/illacloud/builder-backend/pkg/plugins/mssql"
	"github.com/illacloud/builder-backend/pkg/plugins/mysql"
	"github.com/illacloud/builder-backend/pkg/plugins/oracle"
	"github.com/illacloud/builder-backend/pkg/plugins/postgresql"
	"github.com/illacloud/builder-backend/pkg/plugins/redis"
	"github.com/illacloud/builder-backend/pkg/plugins/restapi"
	"github.com/illacloud/builder-backend/pkg/plugins/s3"
	"github.com/illacloud/builder-backend/pkg/plugins/smtp"
	"github.com/illacloud/builder-backend/pkg/plugins/snowflake"
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
	FIREBASE_ACTION      = "firebase"
	CLICKHOUSE_ACTION    = "clickhouse"
	GRAPHQL_ACTION       = "graphql"
	MSSQL_ACTION         = "mssql"
	HUGGINGFACE_ACTION   = "huggingface"
	DYNAMODB_ACTION      = "dynamodb"
	SNOWFLAKE_ACTION     = "snowflake"
	COUCHDB_ACTION       = "couchdb"
	HFENDPOINT_ACTION    = "hfendpoint"
	ORACLE_ACTION        = "oracle"
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
	case FIREBASE_ACTION:
		firebaseAction := &firebase.Connector{}
		return firebaseAction
	case CLICKHOUSE_ACTION:
		clickhouseAction := &clickhouse.Connector{}
		return clickhouseAction
	case GRAPHQL_ACTION:
		graphqlAction := &graphql.Connector{}
		return graphqlAction
	case MSSQL_ACTION:
		mssqlAction := &mssql.Connector{}
		return mssqlAction
	case HUGGINGFACE_ACTION:
		hfAction := &huggingface.Connector{}
		return hfAction
	case DYNAMODB_ACTION:
		dynamodbAction := &dynamodb.Connector{}
		return dynamodbAction
	case SNOWFLAKE_ACTION:
		snowflakeAction := &snowflake.Connector{}
		return snowflakeAction
	case COUCHDB_ACTION:
		couchdbAction := &couchdb.Connector{}
		return couchdbAction
	case HFENDPOINT_ACTION:
		hfendpointAction := &hfendpoint.Connector{}
		return hfendpointAction
	case ORACLE_ACTION:
		oracleAction := &oracle.Connector{}
		return oracleAction
	default:
		return nil
	}
}
