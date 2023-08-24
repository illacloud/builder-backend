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

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/tokenvalidator"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type InternalActionRestHandler interface {
	GenerateSQL(c *gin.Context)
}

type InternalActionRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	ResourceService resource.ResourceService
	AttributeGroup  *accesscontrol.AttributeGroup
}

func NewInternalActionRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService, attrg *accesscontrol.AttributeGroup) *InternalActionRestHandlerImpl {
	return &InternalActionRestHandlerImpl{
		logger:          logger,
		ResourceService: resourceService,
		AttributeGroup:  attrg,
	}
}

func (impl InternalActionRestHandlerImpl) GenerateSQL(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// fetch payload
	req := model.NewGenerateSQLRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}
	resourceID := req.ExportResourceIDInInt()

	// validate sql generate special management
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_PERIPHERAL_SERVICE)
	controller.AttributeGroup.SetUnitID(accesscontrol.DEFAULT_UNIT_ID)
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(accesscontrol.ACTION_SPECIAL_GENERATE_SQL)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// validate resource access
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_RESOURCE)
	controller.AttributeGroup.SetUnitID(resourceID)
	canAccessResource, errInCheckResourceAttr := controller.AttributeGroup.CanAccess(accesscontrol.ACTION_ACCESS_VIEW)
	if errInCheckResourceAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckResourceAttr.Error(),
		})
		return
	}
	if !canAccessResource {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_TRIGGER_TASK,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		TaskName:  auditlogger.TASK_GENERATE_SQL,
		TaskInput: map[string]interface{}{"content": req.Description},
	})

	// fetch resource
	resource, errInGetResource := controller.ResourceService.GetResource(teamID, resourceID)
	if errInGetResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "error in fetch resource: "+errInGetResource.Error())
		return
	}

	// fetch resource meta info
	resourceMetaInfo, errInGetMetaInfo := controller.ResourceService.GetMetaInfo(teamID, resourceID)
	if errInGetMetaInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO, "error in fetch resource meta info: "+errInGetMetaInfo.Error())
		return
	}

	tokenValidator := tokenvalidator.NewRequestTokenValidator()

	// form request payload
	generateSQLPeriReq, errInNewReq := model.NewGenerateSQLPeripheralRequest(resource.Type, resourceMetaInfo, req)
	if errInNewReq != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate request failed: "+errInNewReq.Error())
		return
	}
	token := tokenValidator.GenerateValidateToken(generateSQLPeriReq.Description)
	generateSQLPeriReq.SetValidateToken(token)

	// call remote generate sql API
	generateSQLResp, errInGGenerateSQL := model.GenerateSQL(generateSQLPeriReq, req)
	if errInGGenerateSQL != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate sql failed: "+errInGGenerateSQL.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, generateSQLResp)
}
