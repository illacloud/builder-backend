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

package db

import (
	"fmt"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Addr     string `env:"ILLA_PG_ADDR" envDefault:"127.0.0.1"`
	Port     string `env:"ILLA_PG_PORT" envDefault:"5432"`
	User     string `env:"ILLA_PG_USER" envDefault:"illa"`
	Password string `env:"ILLA_PG_PASSWORD" envDefault:"illa2022"`
	Database string `env:"ILLA_PG_DATABASE" envDefault:"illa"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewDbConnection(cfg *Config, logger *zap.SugaredLogger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host='%s' user='%s' password='%s' dbname='%s' port='%s'",
			cfg.Addr, cfg.User, cfg.Password, cfg.Database, cfg.Port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorw("error in connecting db ", "db", cfg, "err", err)
		return nil, err
	}

	// check db connection
	err = sqlDB.Ping()
	if err != nil {
		logger.Errorw("error in connecting db ", "db", cfg, "err", err)
		return nil, err
	}

	logger.Infow("connected with db", "db", cfg)

	return db, err
}
