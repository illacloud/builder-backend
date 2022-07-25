//go:build wireinject
// +build wireinject

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
	"github.com/illa-family/builder-backend/api/router"
	"github.com/illa-family/builder-backend/cmd/wireset"
	"github.com/illa-family/builder-backend/internal/util"
	"github.com/illa-family/builder-backend/pkg/db"
	"github.com/illa-family/builder-backend/pkg/smtp"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func Initialize() (*Server, error) {
	wire.Build(
		db.DbWireSet,
		smtp.SMTPWireSet,
		util.NewSugardLogger,
		wireset.ResourceWireSet,
		wireset.AppWireSet,
		wireset.ActionWireSet,
		wireset.UserWireSet,
		router.NewRESTRouter,
		GetAppConfig,
		gin.New,
		NewServer,
	)
	return &Server{}, nil
}
