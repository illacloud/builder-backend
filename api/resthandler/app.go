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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AppRequest struct {
	Name       string        `json:"appName" validate:"required"`
	InitScheme []interface{} `json:"initScheme"`
}

type AppRestHandler interface {
	CreateApp(c *gin.Context)
	DeleteApp(c *gin.Context)
	RenameApp(c *gin.Context)
	GetAllApps(c *gin.Context)
	GetMegaData(c *gin.Context)
	DuplicateApp(c *gin.Context)
	ReleaseApp(c *gin.Context)
}

type AppRestHandlerImpl struct {
	logger           *zap.SugaredLogger
	appService       app.AppService
	treeStateService state.TreeStateService
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService, treeStateService state.TreeStateService) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		logger:           logger,
		appService:       appService,
		treeStateService: treeStateService,
	}
}

func (impl AppRestHandlerImpl) CreateApp(c *gin.Context) {
	// Get User from auth middleware
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	appDto := app.AppDto{
		Name:      payload.Name,
		CreatedBy: user,
		UpdatedBy: user,
	}

	// Call `app service` create app
	res, err := impl.appService.CreateApp(appDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "create app error: " + err.Error(),
		})
		return
	}

	if len(payload.InitScheme) > 0 {
		for _, v := range payload.InitScheme {
			componentTree := repository.ConstructComponentNodeByMap(v)
			_ = impl.treeStateService.CreateComponentTree(&res, 0, componentTree)
		}
	}

	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	// Parse URL param to `app ID`
	id, err := strconv.Atoi(c.Param("app"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Call `app service` delete app
	if err := impl.appService.DeleteApp(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "delete app error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"appId": id,
	})
}

func (impl AppRestHandlerImpl) RenameApp(c *gin.Context) {
	// Get User from auth middleware
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// Parse URL param to `app ID`
	id, err := strconv.Atoi(c.Param("app"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Call `app service` update app
	appDTO, err := impl.appService.FetchAppByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "rename app error: " + err.Error(),
		})
		return
	}
	appDTO.Name = payload.Name
	appDTO.UpdatedBy = user
	res, err := impl.appService.UpdateApp(appDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "rename app error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) GetAllApps(c *gin.Context) {
	// Call `app service` get all apps
	res, err := impl.appService.GetAllApps()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get all apps error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) GetMegaData(c *gin.Context) {
	// Parse URL param to `app ID`
	id, err := strconv.Atoi(c.Param("app"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Parse URL param to `version`
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Fetch Mega data via `app` and `version`
	res, err := impl.appService.GetMegaData(id, version)
	if err != nil {
		if err.Error() == "content not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"errorCode":    404,
				"errorMessage": "get app mega data error: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get app mega data error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) DuplicateApp(c *gin.Context) {
	// Get User from auth middleware
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}
	// Parse URL param to `app ID`
	id, err := strconv.Atoi(c.Param("app"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// Call `app service` to duplicate app
	res, err := impl.appService.DuplicateApp(id, user, payload.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "duplicate app error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) ReleaseApp(c *gin.Context) {
	// Parse URL param to `app ID`
	id, err := strconv.Atoi(c.Param("app"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}
	// Call `app service` to release app
	version, err := impl.appService.ReleaseApp(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "release app error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}
