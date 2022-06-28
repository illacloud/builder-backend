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

	"github.com/illa-family/builder-backend/pkg/resource"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ResourceRestHandler interface {
	FindAllResources(c *gin.Context)
	CreateResource(c *gin.Context)
	GetResource(c *gin.Context)
	UpdateResource(c *gin.Context)
	DeleteResource(c *gin.Context)
	TestConnection(c *gin.Context)
}

type ResourceRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService resource.ResourceService
}

func NewResourceRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService) *ResourceRestHandlerImpl {
	return &ResourceRestHandlerImpl{
		logger:          logger,
		resourceService: resourceService,
	}
}

func (impl ResourceRestHandlerImpl) FindAllResources(c *gin.Context) {
	res, err := impl.resourceService.FindAllResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) CreateResource(c *gin.Context) {
	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	rsc.ResourceId = uuid.New()
	res, err := impl.resourceService.CreateResource(rsc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) GetResource(c *gin.Context) {
	resourceId := c.Param("id")
	id, err := uuid.Parse(resourceId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	res, err := impl.resourceService.GetResource(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) UpdateResource(c *gin.Context) {
	resourceId := c.Param("id")
	id, err := uuid.Parse(resourceId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	rsc.ResourceId = id
	res, err := impl.resourceService.UpdateResource(rsc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) DeleteResource(c *gin.Context) {
	resourceId := c.Param("id")
	id, err := uuid.Parse(resourceId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	if err := impl.resourceService.DeleteResource(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resourceId": resourceId,
	})
}

func (impl ResourceRestHandlerImpl) TestConnection(c *gin.Context) {
	// format data to DTO struct
	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	dbConn, err := impl.resourceService.OpenConnection(rsc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	if err := dbConn.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	defer dbConn.Close()
	c.JSON(http.StatusOK, gin.H{
		"message": "test connection successfully",
	})
}
