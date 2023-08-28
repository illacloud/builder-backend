package postgres

import (
	"fmt"
	"time"

	"github.com/illacloud/builder-backend/src/utils/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const RETRY_TIMES = 6

type PostgresConfig struct {
	Addr     string `env:"ILLA_PG_ADDR" envDefault:"localhost"`
	Port     string `env:"ILLA_PG_PORT" envDefault:"5433"`
	User     string `env:"ILLA_PG_USER" envDefault:"illa_cloud"`
	Password string `env:"ILLA_PG_PASSWORD" envDefault:"illa2022"`
	Database string `env:"ILLA_PG_DATABASE" envDefault:"illa_cloud"`
}

func NewPostgresConnectionByGlobalConfig(config *config.Config, logger *zap.SugaredLogger) (*gorm.DB, error) {
	postgresConfig := &PostgresConfig{
		Addr:     config.GetPostgresAddr(),
		Port:     config.GetPostgresPort(),
		User:     config.GetPostgresUser(),
		Password: config.GetPostgresPassword(),
		Database: config.GetPostgresDatabase(),
	}
	return NewPostgresConnection(postgresConfig, logger)
}

func NewPostgresConnection(config *PostgresConfig, logger *zap.SugaredLogger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	retries := RETRY_TIMES

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host='%s' user='%s' password='%s' dbname='%s' port='%s'",
			config.Addr, config.User, config.Password, config.Database, config.Port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	for err != nil {
		if logger != nil {
			logger.Errorw("Failed to connect to database, %d", retries)
		}
		if retries > 1 {
			retries--
			time.Sleep(10 * time.Second)
			db, err = gorm.Open(postgres.New(postgres.Config{
				DSN: fmt.Sprintf("host='%s' user='%s' password='%s' dbname='%s' port='%s'",
					config.Addr, config.User, config.Password, config.Database, config.Port),
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{})
			continue
		}
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorw("error in connecting db ", "db", config, "err", err)
		return nil, err
	}

	// check db connection
	err = sqlDB.Ping()
	if err != nil {
		logger.Errorw("error in connecting db ", "db", config, "err", err)
		return nil, err
	}

	logger.Infow("connected with db", "db", config)

	return db, err
}
