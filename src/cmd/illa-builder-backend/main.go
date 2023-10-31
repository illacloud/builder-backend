package main

import (
	"os"

	"github.com/illacloud/builder-backend/src/cache"
	"github.com/illacloud/builder-backend/src/controller"
	"github.com/illacloud/builder-backend/src/drive"
	"github.com/illacloud/builder-backend/src/driver/awss3"
	"github.com/illacloud/builder-backend/src/driver/postgres"
	"github.com/illacloud/builder-backend/src/driver/redis"
	"github.com/illacloud/builder-backend/src/router"
	"github.com/illacloud/builder-backend/src/storage"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/cors"
	"github.com/illacloud/builder-backend/src/utils/logger"
	"github.com/illacloud/builder-backend/src/utils/recovery"
	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	engine *gin.Engine
	router *router.Router
	logger *zap.SugaredLogger
	config *config.Config
}

func NewServer(config *config.Config, engine *gin.Engine, router *router.Router, logger *zap.SugaredLogger) *Server {
	return &Server{
		engine: engine,
		config: config,
		router: router,
		logger: logger,
	}
}

func initStorage(globalConfig *config.Config, logger *zap.SugaredLogger) *storage.Storage {
	postgresDriver, err := postgres.NewPostgresConnectionByGlobalConfig(globalConfig, logger)
	if err != nil {
		logger.Errorw("Error in startup, storage init failed.")
	}
	return storage.NewStorage(postgresDriver, logger)
}

func initCache(globalConfig *config.Config, logger *zap.SugaredLogger) *cache.Cache {
	redisDriver, err := redis.NewRedisConnectionByGlobalConfig(globalConfig, logger)
	if err != nil {
		logger.Errorw("Error in startup, cache init failed.")
	}
	return cache.NewCache(redisDriver, logger)
}

func initDrive(globalConfig *config.Config, logger *zap.SugaredLogger) *drive.Drive {
	if globalConfig.IsAWSTypeDrive() {
		teamAWSConfig := awss3.NewTeamAwsConfigByGlobalConfig(globalConfig)
		teamDriveS3Instance := awss3.NewS3Drive(teamAWSConfig)
		return drive.NewDrive(teamDriveS3Instance, logger)
	}
	// failed
	logger.Errorw("Error in startup, drive init failed.")
	return nil
}

func initServer() (*Server, error) {
	globalConfig := config.GetInstance()
	engine := gin.New()
	sugaredLogger := logger.NewSugardLogger()

	// init validator
	validator := tokenvalidator.NewRequestTokenValidator()

	// init driver
	storage := initStorage(globalConfig, sugaredLogger)
	cache := initCache(globalConfig, sugaredLogger)
	drive := initDrive(globalConfig, sugaredLogger)

	// init attribute group
	attrg, errInNewAttributeGroup := accesscontrol.NewRawAttributeGroup()
	if errInNewAttributeGroup != nil {
		return nil, errInNewAttributeGroup
	}

	// init controller
	c := controller.NewControllerForBackend(storage, cache, drive, validator, attrg)
	router := router.NewRouter(c)
	server := NewServer(globalConfig, engine, router, sugaredLogger)
	return server, nil

}

func (server *Server) Start() {
	server.logger.Infow("Starting illa-builder-backend...")

	// init
	gin.SetMode(server.config.ServerMode)

	// init cors
	server.engine.Use(gin.CustomRecovery(recovery.CorsHandleRecovery))
	server.engine.Use(cors.Cors())
	server.router.RegisterRouters(server.engine)

	// run
	err := server.engine.Run(server.config.ServerHost + ":" + server.config.ServerPort)
	if err != nil {
		server.logger.Errorw("Error in startup", "err", err)
		os.Exit(2)
	}
}

func main() {
	server, err := initServer()

	if err != nil {

	}

	server.Start()
}
