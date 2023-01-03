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
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/pkg/resource"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ResourceRestHandler interface {
	FindAllResources(c *gin.Context)
	CreateResource(c *gin.Context)
	GetResource(c *gin.Context)
	UpdateResource(c *gin.Context)
	DeleteResource(c *gin.Context)
	TestConnection(c *gin.Context)
	GetMetaInfo(c *gin.Context)
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
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get resources error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) CreateResource(c *gin.Context) {
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

	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}
	if err := impl.resourceService.ValidateResourceOptions(rsc.Type, rsc.Options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	rsc.CreatedAt = time.Now().UTC()
	rsc.CreatedBy = user
	rsc.UpdatedAt = time.Now().UTC()
	rsc.UpdatedBy = user
	res, err := impl.resourceService.CreateResource(rsc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "create resource error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) GetResource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("resource"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}

	res, err := impl.resourceService.GetResource(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get resource error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) UpdateResource(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("resource"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}

	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error",
		})
		return
	}

	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}
	if err := impl.resourceService.ValidateResourceOptions(rsc.Type, rsc.Options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	rsc.ID = id
	rsc.UpdatedBy = user
	rsc.UpdatedAt = time.Now().UTC()
	res, err := impl.resourceService.UpdateResource(rsc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update resource error: " + err.Error(),
		})
		return
	}
	originInfo, _ := impl.resourceService.GetResource(rsc.ID)
	res.CreatedAt = originInfo.CreatedAt
	res.CreatedBy = originInfo.CreatedBy

	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) DeleteResource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("resource"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse url param error: " + err.Error(),
		})
		return
	}

	if err := impl.resourceService.DeleteResource(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "delete resource error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resourceId": id,
	})
}

func (impl ResourceRestHandlerImpl) TestConnection(c *gin.Context) {
	// format data to DTO struct
	var rsc resource.ResourceDto
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	connRes, err := impl.resourceService.TestConnection(rsc)
	if err != nil || !connRes {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "test connection failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "test connection successfully",
	})
}

func (impl ResourceRestHandlerImpl) GetMetaInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("resource"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	res, err := impl.resourceService.GetMetaInfo(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, res)
	return
}
