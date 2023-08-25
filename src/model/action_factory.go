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

package model

import (
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/aiagent"
	"github.com/illacloud/builder-backend/src/actionruntime/airtable"
	"github.com/illacloud/builder-backend/src/actionruntime/appwrite"
	"github.com/illacloud/builder-backend/src/actionruntime/clickhouse"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/illacloud/builder-backend/src/actionruntime/couchdb"
	"github.com/illacloud/builder-backend/src/actionruntime/dynamodb"
	"github.com/illacloud/builder-backend/src/actionruntime/elasticsearch"
	"github.com/illacloud/builder-backend/src/actionruntime/firebase"
	"github.com/illacloud/builder-backend/src/actionruntime/googlesheets"
	"github.com/illacloud/builder-backend/src/actionruntime/graphql"
	"github.com/illacloud/builder-backend/src/actionruntime/hfendpoint"
	"github.com/illacloud/builder-backend/src/actionruntime/huggingface"
	"github.com/illacloud/builder-backend/src/actionruntime/mongodb"
	"github.com/illacloud/builder-backend/src/actionruntime/mssql"
	"github.com/illacloud/builder-backend/src/actionruntime/mysql"
	"github.com/illacloud/builder-backend/src/actionruntime/oracle"
	"github.com/illacloud/builder-backend/src/actionruntime/postgresql"
	"github.com/illacloud/builder-backend/src/actionruntime/redis"
	"github.com/illacloud/builder-backend/src/actionruntime/restapi"
	"github.com/illacloud/builder-backend/src/actionruntime/s3"
	"github.com/illacloud/builder-backend/src/actionruntime/smtp"
	"github.com/illacloud/builder-backend/src/actionruntime/snowflake"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type ActionFactory struct {
	Type int
}

func NewActionFactoryByAction(action *Action) *ActionFactory {
	return &ActionFactory{
		Type: action.Type,
	}
}

func (f *ActionFactory) Build() (*common.DataConnector, error) {
	switch f.Type {
	case resourcelist.TYPE_RESTAPI_ID:
		restapiAction := &restapi.RESTAPIConnector{}
		return restapiAction, nil
	case resourcelist.TYPE_AI_AGENT_ID:
		aiAgentAction := &aiagent.AIAgentConnector{}
		return aiAgentAction, nil
	case resourcelist.TYPE_MYSQL_ID, resourcelist.TYPE_MARIADB_ID, resourcelist.TYPE_TIDB_ID:
		sqlAction := &mysql.MySQLConnector{}
		return sqlAction, nil
	case resourcelist.TYPE_POSTGRESQL_ID, resourcelist.TYPE_SUPABASEDB_ID, resourcelist.TYPE_NEON_ID, resourcelist.TYPE_HYDRA_ID:
		pgsAction := &postgresql.Connector{}
		return pgsAction, nil
	case resourcelist.TYPE_REDIS_ID, resourcelist.TYPE_UPSTASH_ID:
		redisAction := &redis.Connector{}
		return redisAction, nil
	case resourcelist.TYPE_MONGODB_ID:
		mongoAction := &mongodb.Connector{}
		return mongoAction, nil
	case resourcelist.TYPE_ELASTICSEARCH_ID:
		esAction := &elasticsearch.Connector{}
		return esAction, nil
	case resourcelist.TYPE_S3_ID:
		s3Action := &s3.Connector{}
		return s3Action, nil
	case resourcelist.TYPE_SMTP_ID:
		smtpAction := &smtp.Connector{}
		return smtpAction, nil
	case resourcelist.TYPE_FIREBASE_ID:
		firebaseAction := &firebase.Connector{}
		return firebaseAction, nil
	case resourcelist.TYPE_CLICKHOUSE_ID:
		clickhouseAction := &clickhouse.Connector{}
		return clickhouseAction, nil
	case resourcelist.TYPE_GRAPHQL_ID:
		graphqlAction := &graphql.Connector{}
		return graphqlAction, nil
	case resourcelist.TYPE_MSSQL_ID:
		mssqlAction := &mssql.Connector{}
		return mssqlAction, nil
	case resourcelist.TYPE_HUGGINGFACE_ID:
		hfAction := &huggingface.Connector{}
		return hfAction, nil
	case resourcelist.TYPE_DYNAMODB_ID:
		dynamodbAction := &dynamodb.Connector{}
		return dynamodbAction, nil
	case resourcelist.TYPE_SNOWFLAKE_ID:
		snowflakeAction := &snowflake.Connector{}
		return snowflakeAction, nil
	case resourcelist.TYPE_COUCHDB_ID:
		couchdbAction := &couchdb.Connector{}
		return couchdbAction, nil
	case resourcelist.TYPE_HFENDPOINT_ID:
		hfendpointAction := &hfendpoint.Connector{}
		return hfendpointAction, nil
	case resourcelist.TYPE_ORACLE_ID:
		oracleAction := &oracle.Connector{}
		return oracleAction, nil
	case resourcelist.TYPE_APPWRITE_ID:
		appwriteAction := &appwrite.Connector{}
		return appwriteAction, nil
	case resourcelist.TYPE_GOOGLESHEETS_ID:
		googlesheetsAction := &googlesheets.Connector{}
		return googlesheetsAction, nil
	case resourcelist.TYPE_AIRTABLE_ID:
		airtableAction := &airtable.Connector{}
		return airtableAction, nil
	default:
		return nil, errors.New("invalid ActionType: unsupported type " + resourcelist.GetResourceIDMappedType(f.Type))
	}
}
