// Copyright 2023 Illa Soft, Inc.
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

package auditlogger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const RETRY_TIMES = 6

var once sync.Once
var instance *AuditLogger

type AuditLogger struct {
	db *gorm.DB
}

const ILLA_DEPLOY_MODE_CLOUD = "cloud"

func GetInstance() *AuditLogger {
	once.Do(func() {
		var err error
		if instance == nil {
			instance, err = getLogger() // not thread safe
			if err != nil {
				panic(err)
			}
		}
	})
	return instance
}

type Config struct {
	Addr     string `env:"ILLA_AUDIT_PG_ADDR" envDefault:"localhost"`
	Port     string `env:"ILLA_AUDIT_PG_PORT" envDefault:"5432"`
	User     string `env:"ILLA_AUDIT_PG_USER" envDefault:"illa_supervisor"`
	Password string `env:"ILLA_AUDIT_PG_PASSWORD" envDefault:"71De5JllWSetLYU"`
	Database string `env:"ILLA_AUDIT_PG_DATABASE" envDefault:"illa_supervisor"`
}

func getLogger() (*AuditLogger, error) {
	if os.Getenv("ILLA_DEPLOY_MODE") != ILLA_DEPLOY_MODE_CLOUD {
		return &AuditLogger{db: nil}, nil
	}
	config := &Config{}
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	var db *gorm.DB
	retries := RETRY_TIMES

	// new PostgreSQL connection and retry logic
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host='%s' user='%s' password='%s' dbname='%s' port='%s'",
			config.Addr, config.User, config.Password, config.Database, config.Port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	for err != nil {
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
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// check PostgreSQL connection
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	logger := &AuditLogger{db: db}

	return logger, nil
}
