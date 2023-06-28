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

	"github.com/go-playground/validator/v10"
	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/internal/tokenvalidator"
	"github.com/illacloud/builder-backend/pkg/resource"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type InternalActionRestHandler interface {
	GenerateSQL(c *gin.Context)
}

type InternalActionRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	ResourceService resource.ResourceService
	AttributeGroup  *ac.AttributeGroup
}

func NewInternalActionRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService, attrg *ac.AttributeGroup) *InternalActionRestHandlerImpl {
	return &InternalActionRestHandlerImpl{
		logger:          logger,
		ResourceService: resourceService,
		AttributeGroup:  attrg,
	}
}

func (impl InternalActionRestHandlerImpl) GenerateSQL(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// fetch payload
	req := repository.NewGenerateSQLRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}
	resourceID := req.ExportResourceIDInInt()

	// validate internal action
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_PERIPHERAL_SERVICE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_SPECIAL_GENERATE_SQL)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// validate resource access
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canAccessResource, errInCheckResourceAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckResourceAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckResourceAttr.Error(),
		})
		return
	}
	if !canAccessResource {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch resource
	resource, errInGetResource := impl.ResourceService.GetResource(teamID, resourceID)
	if errInGetResource != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "error in fetch resource: "+errInGetResource.Error())
		return
	}

	// fetch resource meta info
	resourceMetaInfo, errInGetMetaInfo := impl.ResourceService.GetMetaInfo(teamID, resourceID)
	if errInGetMetaInfo != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO, "error in fetch resource meta info: "+errInGetMetaInfo.Error())
		return
	}

	tokenValidator := tokenvalidator.NewRequestTokenValidator()

	// form request payload
	generateSQLPeriReq, errInNewReq := repository.NewGenerateSQLPeripheralRequest(resource.Type, resourceMetaInfo, req)
	if errInNewReq != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate request failed: "+errInNewReq.Error())
		return
	}
	token := tokenValidator.GenerateValidateToken(generateSQLPeriReq.Description)
	generateSQLPeriReq.SetValidateToken(token)

	// call remote generate sql API
	generateSQLResp, errInGGenerateSQL := repository.GenerateSQL(generateSQLPeriReq, req)
	if errInGGenerateSQL != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate sql failed: "+errInGGenerateSQL.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, generateSQLResp)
}
