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
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) RenameApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) GetAllApp(c *gin.Context) {
	c.JSON(http.StatusOK, []map[string]interface{}{
		{"appId": "1f221b62-568b-448c-989e-d3a376273134",
			"appName":          "illa example app",
			"currentVersionId": "450ca3c2-38ff-4f27-a1f7-3e71452f49cd",
			"lastModifiedBy":   "Zhanjiao Deng",
			"lastModifiedAt":   "2022-06-06T14:00:30.780+00:00",
		},
	})
}
