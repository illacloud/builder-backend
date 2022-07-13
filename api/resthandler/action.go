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
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/illa-family/builder-backend/pkg/action"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ActionRestHandler interface {
	CreateAction(c *gin.Context)
	DeleteAction(c *gin.Context)
	UpdateAction(c *gin.Context)
	GetAction(c *gin.Context)
	FindActions(c *gin.Context)
	PreviewAction(c *gin.Context)
	RunAction(c *gin.Context)
}

type ActionRestHandlerImpl struct {
	logger        *zap.SugaredLogger
	actionService action.ActionService
}

func NewActionRestHandlerImpl(logger *zap.SugaredLogger, actionService action.ActionService) *ActionRestHandlerImpl {
	return &ActionRestHandlerImpl{
		logger:        logger,
		actionService: actionService,
	}
}

func (impl ActionRestHandlerImpl) CreateAction(c *gin.Context) {
	// get user as creator
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	app, errA := strconv.Atoi(c.Param("app"))
	version, errV := strconv.Atoi(c.Param("version"))
	if errA != nil || errV != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url error",
		})
		return
	}
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	act.App = app
	act.Version = version
	act.CreatedAt = time.Now().UTC()
	act.CreatedBy = user
	act.UpdatedAt = time.Now().UTC()
	act.UpdatedBy = user
	res, err := impl.actionService.CreateAction(act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "create action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) UpdateAction(c *gin.Context) {
	// get user as creator
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	app, errA := strconv.Atoi(c.Param("app"))
	version, errV := strconv.Atoi(c.Param("version"))
	id, errAc := strconv.Atoi(c.Param("action"))
	if errA != nil || errV != nil || errAc != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url error",
		})
		return
	}
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error" + err.Error(),
		})
		return
	}

	act.ID = id
	act.UpdatedBy = user
	act.App = app
	act.Version = version
	act.UpdatedAt = time.Now().UTC()
	act.UpdatedBy = user
	res, err := impl.actionService.UpdateAction(act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "update action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) DeleteAction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("action"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url error" + err.Error(),
		})
		return
	}
	if err := impl.actionService.DeleteAction(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "delete action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"actionId": id,
	})
}

func (impl ActionRestHandlerImpl) GetAction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("action"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url error" + err.Error(),
		})
		return
	}
	res, err := impl.actionService.GetAction(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "get action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) FindActions(c *gin.Context) {
	app, errA := strconv.Atoi(c.Param("app"))
	version, errV := strconv.Atoi(c.Param("version"))
	if errA != nil || errV != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url error",
		})
		return
	}
	res, err := impl.actionService.FindActionsByAppVersion(app, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "get actions error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) PreviewAction(c *gin.Context) {
	c.Header("Timing-Allow-Origin", "*")
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error" + err.Error(),
		})
		return
	}
	res, err := impl.actionService.RunAction(act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "preview action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) RunAction(c *gin.Context) {
	c.Header("Timing-Allow-Origin", "*")
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error" + err.Error(),
		})
		return
	}
	res, err := impl.actionService.RunAction(act)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1064:") {
			lineNumber, _ := strconv.Atoi(err.Error()[len(err.Error())-1:])
			message := ""
			regexp, _ := regexp.Compile(`to use`)
			match := regexp.FindStringIndex(err.Error())
			if len(match) == 2 {
				message = err.Error()[match[1]:]
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"errorCode":    400,
				"errorMessage": errors.New("SQL syntax error").Error(),
				"errorData": map[string]interface{}{
					"lineNumber": lineNumber,
					"message":    "SQL syntax error" + message,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": "run action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}
