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
	"github.com/illacloud/builder-backend/internal/util/resourcelist"
	"github.com/illacloud/builder-backend/pkg/plugins/airtable"
	"github.com/illacloud/builder-backend/pkg/plugins/appwrite"
	"github.com/illacloud/builder-backend/pkg/plugins/clickhouse"
	"github.com/illacloud/builder-backend/pkg/plugins/common"
	"github.com/illacloud/builder-backend/pkg/plugins/couchdb"
	"github.com/illacloud/builder-backend/pkg/plugins/dynamodb"
	"github.com/illacloud/builder-backend/pkg/plugins/elasticsearch"
	"github.com/illacloud/builder-backend/pkg/plugins/firebase"
	"github.com/illacloud/builder-backend/pkg/plugins/googlesheets"
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

type AbstractResourceFactory interface {
	Build() common.DataConnector
}

type Factory struct {
	Type string
}

func (f *Factory) Generate() common.DataConnector {
	switch f.Type {
	case resourcelist.TYPE_RESTAPI:
		restapiRsc := &restapi.RESTAPIConnector{}
		return restapiRsc
	case resourcelist.TYPE_MYSQL, resourcelist.TYPE_MARIADB, resourcelist.TYPE_TIDB:
		sqlRsc := &mysql.MySQLConnector{}
		return sqlRsc
	case resourcelist.TYPE_POSTGRESQL, resourcelist.TYPE_SUPABASEDB, resourcelist.TYPE_NEON, resourcelist.TYPE_HYDRA:
		pgsRsc := &postgresql.Connector{}
		return pgsRsc
	case resourcelist.TYPE_REDIS, resourcelist.TYPE_UPSTASH:
		redisRsc := &redis.Connector{}
		return redisRsc
	case resourcelist.TYPE_MONGODB:
		mongoRsc := &mongodb.Connector{}
		return mongoRsc
	case resourcelist.TYPE_ELASTICSEARCH:
		esRsc := &elasticsearch.Connector{}
		return esRsc
	case resourcelist.TYPE_S3:
		s3Rsc := &s3.Connector{}
		return s3Rsc
	case resourcelist.TYPE_SMTP:
		smtpRsc := &smtp.Connector{}
		return smtpRsc
	case resourcelist.TYPE_FIREBASE:
		firebaseRsc := &firebase.Connector{}
		return firebaseRsc
	case resourcelist.TYPE_CLICKHOUSE:
		clickhouseRsc := &clickhouse.Connector{}
		return clickhouseRsc
	case resourcelist.TYPE_GRAPHQL:
		graphqlRsc := &graphql.Connector{}
		return graphqlRsc
	case resourcelist.TYPE_MSSQL:
		mssqlRsc := &mssql.Connector{}
		return mssqlRsc
	case resourcelist.TYPE_HUGGINGFACE:
		hfRsc := &huggingface.Connector{}
		return hfRsc
	case resourcelist.TYPE_DYNAMODB:
		dynamodbRsc := &dynamodb.Connector{}
		return dynamodbRsc
	case resourcelist.TYPE_SNOWFLAKE:
		snowflakeRsc := &snowflake.Connector{}
		return snowflakeRsc
	case resourcelist.TYPE_COUCHDB:
		couchdbRsc := &couchdb.Connector{}
		return couchdbRsc
	case resourcelist.TYPE_HFENDPOINT:
		hfendpointRsc := &hfendpoint.Connector{}
		return hfendpointRsc
	case resourcelist.TYPE_ORACLE:
		oracleRsc := &oracle.Connector{}
		return oracleRsc
	case resourcelist.TYPE_APPWRITE:
		appwriteRsc := &appwrite.Connector{}
		return appwriteRsc
	case resourcelist.TYPE_GOOGLESHEETS:
		googlesheetsRsc := &googlesheets.Connector{}
		return googlesheetsRsc
	case resourcelist.TYPE_AIRTABLE:
		airtableRsc := &airtable.Connector{}
		return airtableRsc
	default:
		return nil
	}
}
