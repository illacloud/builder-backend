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

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/api/router"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	ILLA_SERVER_HOST string
	ILLA_SERVER_PORT string
	ILLA_SERVER_MODE string
}

type Server struct {
	Engine     *gin.Engine
	restRouter *router.RESTRouter
	logger     *zap.SugaredLogger
	cfg        *Config
	// TODO: websocket gateway
}

func GetAppConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, nil
}

func NewServer(cfg *Config, engine *gin.Engine, restRouter *router.RESTRouter, logger *zap.SugaredLogger) *Server {
	return &Server{
		Engine:     engine,
		cfg:        cfg,
		restRouter: restRouter,
		logger:     logger,
	}
}

func Initialize() (*Server, error) {
	return nil, nil
}

func (server *Server) Start() {
	server.logger.Infow("Starting server")

	server.restRouter.Init()

	err := server.Engine.Run(server.cfg.ILLA_SERVER_HOST + ":" + server.cfg.ILLA_SERVER_PORT)
	if err != nil {
		server.logger.Errorw("Error in startup", "err", err)
		os.Exit(2)
	}
}
