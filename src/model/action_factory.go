package model

import (
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/aiagent"
	"github.com/illacloud/builder-backend/src/actionruntime/airtable"
	"github.com/illacloud/builder-backend/src/actionruntime/appwrite"
	"github.com/illacloud/builder-backend/src/actionruntime/clickhouse"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/illacloud/builder-backend/src/actionruntime/condition"
	"github.com/illacloud/builder-backend/src/actionruntime/couchdb"
	"github.com/illacloud/builder-backend/src/actionruntime/dynamodb"
	"github.com/illacloud/builder-backend/src/actionruntime/elasticsearch"
	"github.com/illacloud/builder-backend/src/actionruntime/firebase"
	"github.com/illacloud/builder-backend/src/actionruntime/googlesheets"
	"github.com/illacloud/builder-backend/src/actionruntime/graphql"
	"github.com/illacloud/builder-backend/src/actionruntime/hfendpoint"
	"github.com/illacloud/builder-backend/src/actionruntime/huggingface"
	"github.com/illacloud/builder-backend/src/actionruntime/illadrive"
	"github.com/illacloud/builder-backend/src/actionruntime/mongodb"
	"github.com/illacloud/builder-backend/src/actionruntime/mssql"
	"github.com/illacloud/builder-backend/src/actionruntime/mysql"
	"github.com/illacloud/builder-backend/src/actionruntime/oracle"
	"github.com/illacloud/builder-backend/src/actionruntime/oracle9i"
	"github.com/illacloud/builder-backend/src/actionruntime/postgresql"
	"github.com/illacloud/builder-backend/src/actionruntime/redis"
	"github.com/illacloud/builder-backend/src/actionruntime/restapi"
	"github.com/illacloud/builder-backend/src/actionruntime/s3"
	"github.com/illacloud/builder-backend/src/actionruntime/serversidetransformer"
	"github.com/illacloud/builder-backend/src/actionruntime/smtp"
	"github.com/illacloud/builder-backend/src/actionruntime/snowflake"
	"github.com/illacloud/builder-backend/src/actionruntime/trigger"
	"github.com/illacloud/builder-backend/src/actionruntime/webhookresponse"
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

func NewFlowActionFactoryByFlowAction(flowAction *FlowAction) *ActionFactory {
	return &ActionFactory{
		Type: flowAction.Type,
	}
}

func NewActionFactoryByResource(resource *Resource) *ActionFactory {
	return &ActionFactory{
		Type: resource.Type,
	}
}

func (f *ActionFactory) Build() (common.DataConnector, error) {
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
	case resourcelist.TYPE_ORACLE_9I_ID:
		oracle9iAction := &oracle9i.Connector{}
		return oracle9iAction, nil
	case resourcelist.TYPE_ILLA_DRIVE_ID:
		illadriveAction := &illadrive.IllaDriveConnector{}
		return illadriveAction, nil
	case resourcelist.TYPE_TRIGGER_ID:
		triggerAction := &trigger.TriggerConnector{}
		return triggerAction, nil
	case resourcelist.TYPE_SERVER_SIDE_TRANSFORMER_ID:
		transformerAction := &serversidetransformer.ServerSideTransformerConnector{}
		return transformerAction, nil
	case resourcelist.TYPE_CONDITION_ID:
		conditionAction := &condition.ConditionConnector{}
		return conditionAction, nil
	case resourcelist.TYPE_WEBHOOK_RESPONSE_ID:
		webhookResponseAction := &webhookresponse.WebhookResponseConnector{}
		return webhookResponseAction, nil
	default:
		return nil, errors.New("invalid ActionType: unsupported type " + resourcelist.GetResourceIDMappedType(f.Type))
	}
}
