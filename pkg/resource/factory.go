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

package resource

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
	REST_RESOURCE          = "restapi"
	MYSQL_RESOURCE         = "mysql"
	MARIADB_RESOURCE       = "mariadb"
	TIDB_RESOURCE          = "tidb"
	POSTGRES_RESOURCE      = "postgresql"
	REDIS_RESOURCE         = "redis"
	MONGODB_RESOURCE       = "mongodb"
	ELASTICSEARCH_RESOURCE = "elasticsearch"
	S3_RESOURCE            = "s3"
	SMTP_RESOURCE          = "smtp"
	SUPABASEDB_RESOURCE    = "supabasedb"
	FIREBASE_RESOURCE      = "firebase"
	CLICKHOUSE_RESOURCE    = "clickhouse"
	GRAPHQL_RESOURCE       = "graphql"
	MSSQL_RESOURCE         = "mssql"
	HUGGINGFACE_RESOURCE   = "huggingface"
	DYNAMODB_RESOURCE      = "dynamodb"
	SNOWFLAKE_RESOURCE     = "snowflake"
	COUCHDB_RESOURCE       = "couchdb"
	HFENDPOINT_RESOURCE    = "hfendpoint"
	ORACLE_RESOURCE        = "oracle"
)

type AbstractResourceFactory interface {
	Build() common.DataConnector
}

type Factory struct {
	Type string
}

func (f *Factory) Generate() common.DataConnector {
	switch f.Type {
	case REST_RESOURCE:
		restapiRsc := &restapi.RESTAPIConnector{}
		return restapiRsc
	case MYSQL_RESOURCE, MARIADB_RESOURCE, TIDB_RESOURCE:
		sqlRsc := &mysql.MySQLConnector{}
		return sqlRsc
	case POSTGRES_RESOURCE, SUPABASEDB_RESOURCE:
		pgsRsc := &postgresql.Connector{}
		return pgsRsc
	case REDIS_RESOURCE:
		redisRsc := &redis.Connector{}
		return redisRsc
	case MONGODB_RESOURCE:
		mongoRsc := &mongodb.Connector{}
		return mongoRsc
	case ELASTICSEARCH_RESOURCE:
		esRsc := &elasticsearch.Connector{}
		return esRsc
	case S3_RESOURCE:
		s3Rsc := &s3.Connector{}
		return s3Rsc
	case SMTP_RESOURCE:
		smtpRsc := &smtp.Connector{}
		return smtpRsc
	case FIREBASE_RESOURCE:
		firebaseRsc := &firebase.Connector{}
		return firebaseRsc
	case CLICKHOUSE_RESOURCE:
		clickhouseRsc := &clickhouse.Connector{}
		return clickhouseRsc
	case GRAPHQL_RESOURCE:
		graphqlRsc := &graphql.Connector{}
		return graphqlRsc
	case MSSQL_RESOURCE:
		mssqlRsc := &mssql.Connector{}
		return mssqlRsc
	case HUGGINGFACE_RESOURCE:
		hfRsc := &huggingface.Connector{}
		return hfRsc
	case DYNAMODB_RESOURCE:
		dynamodbRsc := &dynamodb.Connector{}
		return dynamodbRsc
	case SNOWFLAKE_RESOURCE:
		snowflakeRsc := &snowflake.Connector{}
		return snowflakeRsc
	case COUCHDB_RESOURCE:
		couchdbRsc := &couchdb.Connector{}
		return couchdbRsc
	case HFENDPOINT_RESOURCE:
		hfendpointRsc := &hfendpoint.Connector{}
		return hfendpointRsc
	case ORACLE_RESOURCE:
		oracleRsc := &oracle.Connector{}
		return oracleRsc
	default:
		return nil
	}
}
