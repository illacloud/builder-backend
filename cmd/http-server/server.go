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

package main

import (
	"os"

	"github.com/illacloud/builder-backend/api/router"
	"github.com/illacloud/builder-backend/pkg/cors"
	"github.com/illacloud/builder-backend/pkg/recovery"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	ILLA_SERVER_HOST string `env:"ILLA_SERVER_HOST" envDefault:"0.0.0.0"`
	ILLA_SERVER_PORT string `env:"ILLA_SERVER_PORT" envDefault:"9999"`
	ILLA_SERVER_MODE string `env:"ILLA_SERVER_MODE" envDefault:"debug"`
}

type Server struct {
	engine     *gin.Engine
	restRouter *router.RESTRouter
	logger     *zap.SugaredLogger
	cfg        *Config
}

func GetAppConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewServer(cfg *Config, engine *gin.Engine, restRouter *router.RESTRouter, logger *zap.SugaredLogger) *Server {
	return &Server{
		engine:     engine,
		cfg:        cfg,
		restRouter: restRouter,
		logger:     logger,
	}
}

func (server *Server) Start() {
	server.logger.Infow("Starting server")

	gin.SetMode(server.cfg.ILLA_SERVER_MODE)

	corsHandleRecovery := recovery.CorsHandleRecovery()
	server.engine.Use(gin.CustomRecovery(corsHandleRecovery))
	server.engine.Use(cors.Cors())
	server.restRouter.InitRouter(server.engine.Group("/api"))

	err := server.engine.Run(server.cfg.ILLA_SERVER_HOST + ":" + server.cfg.ILLA_SERVER_PORT)
	if err != nil {
		server.logger.Errorw("Error in startup", "err", err)
		os.Exit(2)
	}
}
