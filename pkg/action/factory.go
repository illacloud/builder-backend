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

type AbstractActionFactory interface {
	Build() common.DataConnector
}

type Factory struct {
	Type string
}

func (f *Factory) Build() common.DataConnector {
	switch f.Type {
	case resourcelist.TYPE_RESTAPI:
		restapiAction := &restapi.RESTAPIConnector{}
		return restapiAction
	case resourcelist.TYPE_MYSQL, resourcelist.TYPE_MARIADB, resourcelist.TYPE_TIDB:
		sqlAction := &mysql.MySQLConnector{}
		return sqlAction
	case resourcelist.TYPE_POSTGRESQL, resourcelist.TYPE_SUPABASEDB, resourcelist.TYPE_NEON, resourcelist.TYPE_HYDRA:
		pgsAction := &postgresql.Connector{}
		return pgsAction
	case resourcelist.TYPE_REDIS, resourcelist.TYPE_UPSTASH:
		redisAction := &redis.Connector{}
		return redisAction
	case resourcelist.TYPE_MONGODB:
		mongoAction := &mongodb.Connector{}
		return mongoAction
	case resourcelist.TYPE_ELASTICSEARCH:
		esAction := &elasticsearch.Connector{}
		return esAction
	case resourcelist.TYPE_S3:
		s3Action := &s3.Connector{}
		return s3Action
	case resourcelist.TYPE_SMTP:
		smtpAction := &smtp.Connector{}
		return smtpAction
	case resourcelist.TYPE_FIREBASE:
		firebaseAction := &firebase.Connector{}
		return firebaseAction
	case resourcelist.TYPE_CLICKHOUSE:
		clickhouseAction := &clickhouse.Connector{}
		return clickhouseAction
	case resourcelist.TYPE_GRAPHQL:
		graphqlAction := &graphql.Connector{}
		return graphqlAction
	case resourcelist.TYPE_MSSQL:
		mssqlAction := &mssql.Connector{}
		return mssqlAction
	case resourcelist.TYPE_HUGGINGFACE:
		hfAction := &huggingface.Connector{}
		return hfAction
	case resourcelist.TYPE_DYNAMODB:
		dynamodbAction := &dynamodb.Connector{}
		return dynamodbAction
	case resourcelist.TYPE_SNOWFLAKE:
		snowflakeAction := &snowflake.Connector{}
		return snowflakeAction
	case resourcelist.TYPE_COUCHDB:
		couchdbAction := &couchdb.Connector{}
		return couchdbAction
	case resourcelist.TYPE_HFENDPOINT:
		hfendpointAction := &hfendpoint.Connector{}
		return hfendpointAction
	case resourcelist.TYPE_ORACLE:
		oracleAction := &oracle.Connector{}
		return oracleAction
	case resourcelist.TYPE_APPWRITE:
		appwriteAction := &appwrite.Connector{}
		return appwriteAction
	case resourcelist.TYPE_GOOGLESHEETS:
		googlesheetsAction := &googlesheets.Connector{}
		return googlesheetsAction
	case resourcelist.TYPE_AIRTABLE:
		airtableAction := &airtable.Connector{}
		return airtableAction
	default:
		return nil
	}
}
