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
	"time"

	"github.com/go-playground/validator/v10"
	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
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
	AttributeGroup  *ac.AttributeGroup
}

func NewResourceRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService, attrg *ac.AttributeGroup) *ResourceRestHandlerImpl {
	return &ResourceRestHandlerImpl{
		logger:          logger,
		resourceService: resourceService,
		AttributeGroup:  attrg,
	}
}

func (impl ResourceRestHandlerImpl) FindAllResources(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canAccess {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	// fetch data
	res, err := impl.resourceService.FindAllResources(teamID)
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
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_CREATE_RESOURCE)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canManage {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	var rsc resource.ResourceDto
	rsc.InitUID()
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

	rsc.SetTeamID(teamID)
	rsc.CreatedAt = time.Now().UTC()
	rsc.CreatedBy = userID
	rsc.UpdatedAt = time.Now().UTC()
	rsc.UpdatedBy = userID
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
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canAccess {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	// fetch data
	res, err := impl.resourceService.GetResource(teamID, resourceID)
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
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canManage {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	// parse request body
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

	// update
	rsc.ID = resourceID
	rsc.UpdatedBy = userID
	rsc.UpdatedAt = time.Now().UTC()
	res, err := impl.resourceService.UpdateResource(rsc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update resource error: " + err.Error(),
		})
		return
	}
	originInfo, _ := impl.resourceService.GetResource(teamID, rsc.ID)
	res.CreatedAt = originInfo.CreatedAt
	res.CreatedBy = originInfo.CreatedBy

	c.JSON(http.StatusOK, res)
}

func (impl ResourceRestHandlerImpl) DeleteResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canDelete, errInCheckAttr := impl.AttributeGroup.CanDelete(ac.ACTION_DELETE)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canDelete {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	if err := impl.resourceService.DeleteResource(teamID, resourceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "delete resource error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resourceId": resourceID,
	})
}

func (impl ResourceRestHandlerImpl) TestConnection(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canManage {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

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
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckAttr.Error(),
		})
		return
	}
	if !canAccess {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "you can not access this attribute due to access control policy.",
		})
		return
	}

	// fetch data
	res, err := impl.resourceService.GetMetaInfo(teamID, resourceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, res)
	return
}
