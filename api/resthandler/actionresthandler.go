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

	"github.com/illa-family/builder-backend/pkg/action"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	versionId := c.Param("versionId")
	vId, err := uuid.Parse(versionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	act.ActionId = uuid.New()
	res, err := impl.actionService.CreateAction(vId, act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) UpdateAction(c *gin.Context) {
	actionId := c.Param("id")
	aId, err := uuid.Parse(actionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	versionId := c.Param("versionId")
	vId, err := uuid.Parse(versionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	act.ActionId = aId
	res, err := impl.actionService.UpdateAction(vId, act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) DeleteAction(c *gin.Context) {
	actionId := c.Param("id")
	aId, err := uuid.Parse(actionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	if err := impl.actionService.DeleteAction(aId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"actionId": actionId,
	})
}

func (impl ActionRestHandlerImpl) GetAction(c *gin.Context) {
	actionId := c.Param("id")
	aId, err := uuid.Parse(actionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	res, err := impl.actionService.GetAction(aId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) FindActions(c *gin.Context) {
	versionId := c.Param("versionId")
	vId, err := uuid.Parse(versionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	res, err := impl.actionService.FindActionsByVersion(vId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) PreviewAction(c *gin.Context) {
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	res, err := impl.actionService.RunAction(act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) RunAction(c *gin.Context) {
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	res, err := impl.actionService.RunAction(act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}
