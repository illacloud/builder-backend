package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/caarlos0/env"
)

const DEPLOY_MODE_SELF_HOST = "self-host"
const DEPLOY_MODE_CLOUD = "cloud"
const DEPLOY_MODE_CLOUD_TEST = "cloud-test"
const DEPLOY_MODE_CLOUD_BETA = "cloud-beta"
const DEPLOY_MODE_CLOUD_PRODUCTION = "cloud-production"
const DRIVE_TYPE_AWS = "aws"
const DRIVE_TYPE_DO = "do"
const DRIVE_TYPE_MINIO = "minio"
const PROTOCOL_WEBSOCKET = "ws"
const PROTOCOL_WEBSOCKET_OVER_TLS = "wss"

var instance *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		var err error
		if instance == nil {
			instance, err = getConfig() // not thread safe
			if err != nil {
				panic(err)
			}
		}
	})
	return instance
}

type Config struct {
	// server config
	ServerHost          string `env:"ILLA_SERVER_HOST"              envDefault:"0.0.0.0"`
	ServerPort          string `env:"ILLA_SERVER_PORT"              envDefault:"8003"`
	InternalServerPort  string `env:"ILLA_SERVER_INTERNAL_PORT"     envDefault:"9001"`
	WebsocketServerPort string `env:"ILLA_SERVER_WEBSOCKET_PORT"    envDefault:"8002"`
	ServerMode          string `env:"ILLA_SERVER_MODE"              envDefault:"debug"`
	DeployMode          string `env:"ILLA_DEPLOY_MODE"              envDefault:"cloud-test"`
	SecretKey           string `env:"ILLA_SECRET_KEY" 			    envDefault:"8xEMrWkBARcDDYQ"`
	WSSEnabled          string `env:"ILLA_WSS_ENABLED" 			    envDefault:"false"`
	// key for idconvertor
	RandomKey string `env:"ILLA_RANDOM_KEY"  envDefault:"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`
	// storage config
	PostgresAddr     string `env:"ILLA_PG_ADDR" envDefault:"localhost"`
	PostgresPort     string `env:"ILLA_PG_PORT" envDefault:"5433"`
	PostgresUser     string `env:"ILLA_PG_USER" envDefault:"illa_cloud"`
	PostgresPassword string `env:"ILLA_PG_PASSWORD" envDefault:"illa2022"`
	PostgresDatabase string `env:"ILLA_PG_DATABASE" envDefault:"illa_cloud"`
	// cache config
	RedisAddr     string `env:"ILLA_REDIS_ADDR" envDefault:"localhost"`
	RedisPort     string `env:"ILLA_REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"ILLA_REDIS_PASSWORD" envDefault:"illa2022"`
	RedisDatabase int    `env:"ILLA_REDIS_DATABASE" envDefault:"0"`
	// drive config
	DriveType             string `env:"ILLA_DRIVE_TYPE"               envDefault:""`
	DriveAccessKeyID      string `env:"ILLA_DRIVE_ACCESS_KEY_ID"      envDefault:""`
	DriveAccessKeySecret  string `env:"ILLA_DRIVE_ACCESS_KEY_SECRET"  envDefault:""`
	DriveRegion           string `env:"ILLA_DRIVE_REGION"             envDefault:""`
	DriveEndpoint         string `env:"ILLA_DRIVE_ENDPOINT"           envDefault:""`
	DriveSystemBucketName string `env:"ILLA_DRIVE_SYSTEM_BUCKET_NAME" envDefault:"illa-cloud"`
	DriveTeamBucketName   string `env:"ILLA_DRIVE_TEAM_BUCKET_NAME"   envDefault:"illa-cloud-team"`
	DriveUploadTimeoutRaw string `env:"ILLA_DRIVE_UPLOAD_TIMEOUT"     envDefault:"30s"`
	DriveUploadTimeout    time.Duration
	// peripheral API
	IllaPeripheralAPI string `env:"ILLA_PERIPHERAL_API" envDefault:"https://peripheral-api.illasoft.com/v1/"`
	// resource manager API
	IllaResourceManagerRestAPI         string `env:"ILLA_RESOURCE_MANAGER_API"     envDefault:"http://illa-resource-manager-backend:8006"`
	IllaResourceManagerInternalRestAPI string `env:"ILLA_RESOURCE_MANAGER_INTERNAL_API"     envDefault:"http://illa-resource-manager-backend-internal:9004"`
	// illa marketplace config
	IllaMarketplaceInternalRestAPI string `env:"ILLA_MARKETPLACE_INTERNAL_API"     envDefault:"http://illa-marketplace-backend-internal:9003/api/v1"`
	// token for internal api
	ControlToken string `env:"ILLA_CONTROL_TOKEN"     envDefault:""`
}

func getConfig() (*Config, error) {
	// fetch
	cfg := &Config{}
	err := env.Parse(cfg)
	// process data
	var errInParseDuration error
	cfg.DriveUploadTimeout, errInParseDuration = time.ParseDuration(cfg.DriveUploadTimeoutRaw)
	if errInParseDuration != nil {
		return nil, errInParseDuration
	}
	// ok
	fmt.Printf("----------------\n")
	fmt.Printf("%+v\n", cfg)
	fmt.Printf("%+v\n", err)

	return cfg, err
}

func (c *Config) IsSelfHostMode() bool {
	return c.DeployMode == DEPLOY_MODE_SELF_HOST
}

func (c *Config) IsCloudMode() bool {
	if c.DeployMode == DEPLOY_MODE_CLOUD || c.DeployMode == DEPLOY_MODE_CLOUD_TEST || c.DeployMode == DEPLOY_MODE_CLOUD_BETA || c.DeployMode == DEPLOY_MODE_CLOUD_PRODUCTION {
		return true
	}
	return false
}

func (c *Config) IsCloudTestMode() bool {
	return c.DeployMode == DEPLOY_MODE_CLOUD_TEST
}

func (c *Config) IsCloudBetaMode() bool {
	return c.DeployMode == DEPLOY_MODE_CLOUD_BETA
}

func (c *Config) IsCloudProductionMode() bool {
	return c.DeployMode == DEPLOY_MODE_CLOUD_PRODUCTION
}

func (c *Config) GetWebScoketServerListenAddress() string {
	return c.ServerHost + ":" + c.WebsocketServerPort
}

func (c *Config) GetWebsocketProtocol() string {
	if c.WSSEnabled == "true" {
		return PROTOCOL_WEBSOCKET_OVER_TLS
	}
	return PROTOCOL_WEBSOCKET
}

func (c *Config) GetRuntimeEnv() string {
	if c.IsCloudBetaMode() {
		return DEPLOY_MODE_CLOUD_BETA
	} else if c.IsCloudProductionMode() {
		return DEPLOY_MODE_CLOUD_PRODUCTION
	} else {
		return DEPLOY_MODE_CLOUD_TEST
	}
}

func (c *Config) GetSecretKey() string {
	return c.SecretKey
}

func (c *Config) GetRandomKey() string {
	return c.RandomKey
}

func (c *Config) GetPostgresAddr() string {
	return c.PostgresAddr
}

func (c *Config) GetPostgresPort() string {
	return c.PostgresPort
}

func (c *Config) GetPostgresUser() string {
	return c.PostgresUser
}

func (c *Config) GetPostgresPassword() string {
	return c.PostgresPassword
}

func (c *Config) GetPostgresDatabase() string {
	return c.PostgresDatabase
}

func (c *Config) GetRedisAddr() string {
	return c.RedisAddr
}

func (c *Config) GetRedisPort() string {
	return c.RedisPort
}

func (c *Config) GetRedisPassword() string {
	return c.RedisPassword
}

func (c *Config) GetRedisDatabase() int {
	return c.RedisDatabase
}

func (c *Config) GetDriveType() string {
	return c.DriveType
}

func (c *Config) IsAWSTypeDrive() bool {
	if c.DriveType == DRIVE_TYPE_AWS || c.DriveType == DRIVE_TYPE_DO {
		return true
	}
	return false
}

func (c *Config) IsMINIODrive() bool {
	return c.DriveType == DRIVE_TYPE_MINIO
}

func (c *Config) GetAWSS3Endpoint() string {
	return c.DriveEndpoint
}

func (c *Config) GetAWSS3AccessKeyID() string {
	return c.DriveAccessKeyID
}

func (c *Config) GetAWSS3AccessKeySecret() string {
	return c.DriveAccessKeySecret
}

func (c *Config) GetAWSS3Region() string {
	return c.DriveRegion
}

func (c *Config) GetAWSS3SystemBucketName() string {
	return c.DriveSystemBucketName
}

func (c *Config) GetAWSS3TeamBucketName() string {
	return c.DriveTeamBucketName
}

func (c *Config) GetAWSS3Timeout() time.Duration {
	return c.DriveUploadTimeout
}

func (c *Config) GetMINIOAccessKeyID() string {
	return c.DriveAccessKeyID
}

func (c *Config) GetMINIOAccessKeySecret() string {
	return c.DriveAccessKeySecret
}

func (c *Config) GetMINIOEndpoint() string {
	return c.DriveEndpoint
}

func (c *Config) GetMINIOSystemBucketName() string {
	return c.DriveSystemBucketName
}

func (c *Config) GetMINIOTeamBucketName() string {
	return c.DriveTeamBucketName
}

func (c *Config) GetMINIOTimeout() time.Duration {
	return c.DriveUploadTimeout
}

func (c *Config) GetControlToken() string {
	return c.ControlToken
}

func (c *Config) GetIllaPeripheralAPI() string {
	return c.IllaPeripheralAPI
}

func (c *Config) GetIllaResourceManagerRestAPI() string {
	return c.IllaResourceManagerRestAPI
}

func (c *Config) GetIllaResourceManagerInternalRestAPI() string {
	return c.IllaResourceManagerInternalRestAPI
}

func (c *Config) GetIllaMarketplaceInternalRestAPI() string {
	return c.IllaMarketplaceInternalRestAPI
}
