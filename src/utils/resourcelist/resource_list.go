package resourcelist

var (
	TYPE_TRANSFORMER             = "transformer"
	TYPE_RESTAPI                 = "restapi"
	TYPE_GRAPHQL                 = "graphql"
	TYPE_REDIS                   = "redis"
	TYPE_MYSQL                   = "mysql"
	TYPE_MARIADB                 = "mariadb"
	TYPE_POSTGRESQL              = "postgresql"
	TYPE_MONGODB                 = "mongodb"
	TYPE_TIDB                    = "tidb"
	TYPE_ELASTICSEARCH           = "elasticsearch"
	TYPE_S3                      = "s3"
	TYPE_SMTP                    = "smtp"
	TYPE_SUPABASEDB              = "supabasedb"
	TYPE_FIREBASE                = "firebase"
	TYPE_CLICKHOUSE              = "clickhouse"
	TYPE_MSSQL                   = "mssql"
	TYPE_HUGGINGFACE             = "huggingface"
	TYPE_DYNAMODB                = "dynamodb"
	TYPE_SNOWFLAKE               = "snowflake"
	TYPE_COUCHDB                 = "couchdb"
	TYPE_HFENDPOINT              = "hfendpoint"
	TYPE_ORACLE                  = "oracle"
	TYPE_APPWRITE                = "appwrite"
	TYPE_GOOGLESHEETS            = "googlesheets"
	TYPE_NEON                    = "neon"
	TYPE_UPSTASH                 = "upstash"
	TYPE_AIRTABLE                = "airtable"
	TYPE_HYDRA                   = "hydra"
	TYPE_AI_AGENT                = "aiagent"
	TYPE_ORACLE_9I               = "oracle9i"
	TYPE_ILLA_DRIVE              = "illadrive"
	TYPE_TRIGGER                 = "trigger"
	TYPE_SERVER_SIDE_TRANSFORMER = "serversidetransformer"
	TYPE_CONDITION               = "condition"
	TYPE_WEBHOOK_RESPONSE        = "webhookresponse"
)

var (
	TYPE_TRANSFORMER_ID             = 0
	TYPE_RESTAPI_ID                 = 1
	TYPE_GRAPHQL_ID                 = 2
	TYPE_REDIS_ID                   = 3
	TYPE_MYSQL_ID                   = 4
	TYPE_MARIADB_ID                 = 5
	TYPE_POSTGRESQL_ID              = 6
	TYPE_MONGODB_ID                 = 7
	TYPE_TIDB_ID                    = 8
	TYPE_ELASTICSEARCH_ID           = 9
	TYPE_S3_ID                      = 10
	TYPE_SMTP_ID                    = 11
	TYPE_SUPABASEDB_ID              = 12
	TYPE_FIREBASE_ID                = 13
	TYPE_CLICKHOUSE_ID              = 14
	TYPE_MSSQL_ID                   = 15
	TYPE_HUGGINGFACE_ID             = 16
	TYPE_DYNAMODB_ID                = 17
	TYPE_SNOWFLAKE_ID               = 18
	TYPE_COUCHDB_ID                 = 19
	TYPE_HFENDPOINT_ID              = 20
	TYPE_ORACLE_ID                  = 21
	TYPE_APPWRITE_ID                = 22
	TYPE_GOOGLESHEETS_ID            = 23
	TYPE_NEON_ID                    = 24
	TYPE_UPSTASH_ID                 = 25
	TYPE_AIRTABLE_ID                = 26
	TYPE_HYDRA_ID                   = 27
	TYPE_AI_AGENT_ID                = 28
	TYPE_ORACLE_9I_ID               = 29
	TYPE_ILLA_DRIVE_ID              = 30
	TYPE_TRIGGER_ID                 = 31
	TYPE_SERVER_SIDE_TRANSFORMER_ID = 32
	TYPE_CONDITION_ID               = 33
	TYPE_WEBHOOK_RESPONSE_ID        = 34
)

var type_array = []string{
	0:  TYPE_TRANSFORMER,
	1:  TYPE_RESTAPI,
	2:  TYPE_GRAPHQL,
	3:  TYPE_REDIS,
	4:  TYPE_MYSQL,
	5:  TYPE_MARIADB,
	6:  TYPE_POSTGRESQL,
	7:  TYPE_MONGODB,
	8:  TYPE_TIDB,
	9:  TYPE_ELASTICSEARCH,
	10: TYPE_S3,
	11: TYPE_SMTP,
	12: TYPE_SUPABASEDB,
	13: TYPE_FIREBASE,
	14: TYPE_CLICKHOUSE,
	15: TYPE_MSSQL,
	16: TYPE_HUGGINGFACE,
	17: TYPE_DYNAMODB,
	18: TYPE_SNOWFLAKE,
	19: TYPE_COUCHDB,
	20: TYPE_HFENDPOINT,
	21: TYPE_ORACLE,
	22: TYPE_APPWRITE,
	23: TYPE_GOOGLESHEETS,
	24: TYPE_NEON,
	25: TYPE_UPSTASH,
	26: TYPE_AIRTABLE,
	27: TYPE_HYDRA,
	28: TYPE_AI_AGENT,
	29: TYPE_ORACLE_9I,
	30: TYPE_ILLA_DRIVE,
	31: TYPE_TRIGGER,
	32: TYPE_SERVER_SIDE_TRANSFORMER,
	33: TYPE_CONDITION,
	34: TYPE_WEBHOOK_RESPONSE,
}

var type_map = map[string]int{
	TYPE_TRANSFORMER:             TYPE_TRANSFORMER_ID,
	TYPE_RESTAPI:                 TYPE_RESTAPI_ID,
	TYPE_GRAPHQL:                 TYPE_GRAPHQL_ID,
	TYPE_REDIS:                   TYPE_REDIS_ID,
	TYPE_MYSQL:                   TYPE_MYSQL_ID,
	TYPE_MARIADB:                 TYPE_MARIADB_ID,
	TYPE_POSTGRESQL:              TYPE_POSTGRESQL_ID,
	TYPE_MONGODB:                 TYPE_MONGODB_ID,
	TYPE_TIDB:                    TYPE_TIDB_ID,
	TYPE_ELASTICSEARCH:           TYPE_ELASTICSEARCH_ID,
	TYPE_S3:                      TYPE_S3_ID,
	TYPE_SMTP:                    TYPE_SMTP_ID,
	TYPE_SUPABASEDB:              TYPE_SUPABASEDB_ID,
	TYPE_FIREBASE:                TYPE_FIREBASE_ID,
	TYPE_CLICKHOUSE:              TYPE_CLICKHOUSE_ID,
	TYPE_MSSQL:                   TYPE_MSSQL_ID,
	TYPE_HUGGINGFACE:             TYPE_HUGGINGFACE_ID,
	TYPE_DYNAMODB:                TYPE_DYNAMODB_ID,
	TYPE_SNOWFLAKE:               TYPE_SNOWFLAKE_ID,
	TYPE_COUCHDB:                 TYPE_COUCHDB_ID,
	TYPE_HFENDPOINT:              TYPE_HFENDPOINT_ID,
	TYPE_ORACLE:                  TYPE_ORACLE_ID,
	TYPE_APPWRITE:                TYPE_APPWRITE_ID,
	TYPE_GOOGLESHEETS:            TYPE_GOOGLESHEETS_ID,
	TYPE_NEON:                    TYPE_NEON_ID,
	TYPE_UPSTASH:                 TYPE_UPSTASH_ID,
	TYPE_AIRTABLE:                TYPE_AIRTABLE_ID,
	TYPE_HYDRA:                   TYPE_HYDRA_ID,
	TYPE_AI_AGENT:                TYPE_AI_AGENT_ID,
	TYPE_ORACLE_9I:               TYPE_ORACLE_9I_ID,
	TYPE_ILLA_DRIVE:              TYPE_ILLA_DRIVE_ID,
	TYPE_TRIGGER:                 TYPE_TRIGGER_ID,
	TYPE_SERVER_SIDE_TRANSFORMER: TYPE_SERVER_SIDE_TRANSFORMER_ID,
	TYPE_CONDITION:               TYPE_CONDITION_ID,
	TYPE_WEBHOOK_RESPONSE:        TYPE_WEBHOOK_RESPONSE_ID,
}

var virtualResourceList = map[string]bool{
	TYPE_TRANSFORMER: true,
	TYPE_AI_AGENT:    true,
	TYPE_ILLA_DRIVE:  true,
}

var localVirtualResourceList = map[string]bool{
	TYPE_TRANSFORMER: true,
}

var remoteVirtualResourceList = map[string]bool{
	TYPE_AI_AGENT:   true,
	TYPE_ILLA_DRIVE: true,
}

var emptyOptionResourceList = map[string]bool{
	TYPE_TRANSFORMER: true,
}

var canCreateOAuthTokenResourceList = map[string]bool{
	TYPE_GOOGLESHEETS: true,
}

var needFetchResourceInfoFromSourceManagerList = map[string]bool{
	TYPE_AI_AGENT: true,
}

func GetResourceIDMappedType(id int) string {
	return type_array[id]
}

func GetResourceNameMappedID(name string) int {
	return type_map[name]
}

// The virtual resource have no resource id
func IsVirtualResource(resourceType string) bool {
	itIs, hit := virtualResourceList[resourceType]
	return itIs && hit
}

func IsLocalVirtualResource(resourceType string) bool {
	itIs, hit := localVirtualResourceList[resourceType]
	return itIs && hit
}

func IsRemoteVirtualResource(resourceType string) bool {
	itIs, hit := remoteVirtualResourceList[resourceType]
	return itIs && hit
}

func IsVirtualResourceByIntType(resourceType int) bool {
	resourceTypeString := GetResourceIDMappedType(resourceType)
	itIs, hit := virtualResourceList[resourceTypeString]
	return itIs && hit
}

func IsLocalVirtualResourceByIntType(resourceType int) bool {
	resourceTypeString := GetResourceIDMappedType(resourceType)
	itIs, hit := localVirtualResourceList[resourceTypeString]
	return itIs && hit
}

func IsRemoteVirtualResourceByIntType(resourceType int) bool {
	resourceTypeString := GetResourceIDMappedType(resourceType)
	itIs, hit := remoteVirtualResourceList[resourceTypeString]
	return itIs && hit
}

func IsVirtualResourceHaveNoOption(resourceType int) bool {
	resourceTypeString := GetResourceIDMappedType(resourceType)
	itIs, hit := emptyOptionResourceList[resourceTypeString]
	return itIs && hit
}

func CanCreateOAuthToken(resourceType int) bool {
	resourceTypeString := GetResourceIDMappedType(resourceType)
	canDo, hit := canCreateOAuthTokenResourceList[resourceTypeString]
	return canDo && hit
}

func NeedFetchResourceInfoFromSourceManager(resourceType string) bool {
	itIs, hit := needFetchResourceInfoFromSourceManagerList[resourceType]
	return itIs && hit
}
