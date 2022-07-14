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

package resthandler

import (
	"net/http"

	"github.com/illa-family/builder-backend/pkg/app"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppRestHandler interface {
	CreateApp(c *gin.Context)
	DeleteApp(c *gin.Context)
	RenameApp(c *gin.Context)
	GetAllApp(c *gin.Context)
	GetReleaseVersion(c *gin.Context)
	GetEditingVersion(c *gin.Context)
	DuplicateApp(c *gin.Context)
	ReleaseApp(c *gin.Context)
}

type AppRestHandlerImpl struct {
	logger     *zap.SugaredLogger
	appService app.AppService
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		logger:     logger,
		appService: appService,
	}
}

func (impl AppRestHandlerImpl) CreateApp(c *gin.Context) {
	// Get User from auth middleware
	// Parse request body
	// Validate request body
	// Call `app service` create app
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	// Parse URL param to `app ID`
	// Call `app service` delete app
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) RenameApp(c *gin.Context) {
	// Get User from auth middleware
	// Parse request body
	// Validate request body
	// Call `app service` update app
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) GetAllApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) GetEditingVersion(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) DuplicateApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) ReleaseApp(c *gin.Context) {
	// GetAppByID
	// BumpAppVersion
	// ReleaseKVStateByApp
	// ReleaseTreeStateByApp
	// UpdateApp
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) GetReleaseVersion(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}
