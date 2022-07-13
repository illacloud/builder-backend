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
	GetCurrentReleaseVersion(c *gin.Context)
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
		{
			"appId":            "1f221b62-568b-448c-989e-d3a376273134",
			"appName":          "illa example app",
			"currentVersionId": "450ca3c2-38ff-4f27-a1f7-3e71452f49cd",
			"updatedBy":        "Zhanjiao Deng",
			"updatedAt":        "2022-06-06T14:00:30.780+00:00",
		},
	})
}

func (impl AppRestHandlerImpl) GetCurrentVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"versionId": "450ca3c2-38ff-4f27-a1f7-3e71452f49cd",
		"appInfo": map[string]string{
			"appId":            "1f221b62-568b-448c-989e-d3a376273134",
			"appName":          "illa example app",
			"currentVersionId": "450ca3c2-38ff-4f27-a1f7-3e71452f49cd",
			"updatedBy":        "Zhanjiao Deng",
			"updatedAt":        "2022-06-06T14:00:30.780+00:00",
		},
		"versionName": "v1",
		"components": map[string]interface{}{
			"rootDsl": map[string]interface{}{
				"displayName":    "root",
				"parentNode":     nil,
				"showName":       "root",
				"childrenNode":   []map[string]interface{}{},
				"type":           "DOT_PANEL",
				"containerType":  "EDITOR_DOT_PANEL",
				"verticalResize": true,
				"h":              0,
				"w":              0,
				"x":              -1,
				"y":              -1,
			},
		},
		"actions": []map[string]interface{}{
			{
				"actionId":    "7a68c10e-16b4-4459-9be0-f55a03321a17",
				"resourceId":  "6448c819-2e6b-4f19-976d-19d290e42c3a",
				"displayName": "mysql1",
				"actionType":  "mysql",
				"actionTemplate": map[string]interface{}{
					"mode":  "sql",
					"query": "SELECT * FROM `order` WHERE charge_total > 100 LIMIT 100;",
				},
				"createdBy": "00000000-0000-0000-0000-000000000000",
				"createdAt": "2022-06-27T07:46:08.384931Z",
				"updatedBy": "00000000-0000-0000-0000-000000000000",
				"updatedAt": "2022-06-27T16:03:26.658313Z",
			},
		},
		"dependenciesState": map[string]interface{}{},
		"executionState": map[string]interface{}{
			"result": map[string]interface{}{},
			"error":  map[string]interface{}{},
		},
		"dragShadowState":       map[string]interface{}{},
		"dottedLineSquareState": map[string]interface{}{},
		"displayNameState":      []string{},
		"createdBy":             "1f221b62-568b-448c-989easdqwe2",
		"updatedBy":             "1f221b62-568b-448c-989easdqwe2",
		"createdAt":             "2022-06-06T12:00:30.780+00:00",
		"updatedAt":             "2022-06-06T14:00:30.780+00:00",
	})
}

func (impl AppRestHandlerImpl) DuplicateApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl AppRestHandlerImpl) ReleaseApp(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}
