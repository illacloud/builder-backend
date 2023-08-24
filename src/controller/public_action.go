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
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/illacloud/builder-backend/internal/auditlogger"
	dc "github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/pkg/action"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PublicActionRestHandler interface {
	RunAction(c *gin.Context)
}

type PublicActionRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	appService      app.AppService
	resourceService resource.ResourceService
	actionService   action.ActionService
	AttributeGroup  *accesscontrol.AttributeGroup
}

func NewPublicActionRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService, resourceService resource.ResourceService,
	actionService action.ActionService, attrg *accesscontrol.AttributeGroup) *PublicActionRestHandlerImpl {
	return &PublicActionRestHandlerImpl{
		logger:          logger,
		appService:      appService,
		resourceService: resourceService,
		actionService:   actionService,
		AttributeGroup:  attrg,
	}
}

func (impl PublicActionRestHandlerImpl) RunAction(c *gin.Context) {
	// fetch needed param
	teamIdentifier, errInGetTeamIdentifier := GetStringParamFromRequest(c, PARAM_TEAM_IDENTIFIER)
	publicActionID, errInGetPublicActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	appID, errInGetAppID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	if errInGetTeamIdentifier != nil || errInGetPublicActionID != nil || errInGetAppID != nil {
		return
	}

	// get team id by team teamIdentifier
	team, errInGetTeamInfo := dc.GetTeamInfoByIdentifier(teamIdentifier)
	if errInGetTeamInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_TEAM, "get target team by identifier error: "+errInGetTeamInfo.Error())
		return
	}
	teamID := team.GetID()

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(accesscontrol.ANONYMOUS_AUTH_TOKEN)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_ACTION)
	controller.AttributeGroup.SetUnitID(accesscontrol.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_RUN_ACTION)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// check if action is public action
	if !controller.actionService.IsPublicAction(teamID, publicActionID) {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this action.")
		return
	}

	// execute
	c.Header("Timing-Allow-Origin", "*")
	var actForExport action.ActionDtoForExport
	if err := json.NewDecoder(c.Request.Body).Decode(&actForExport); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error"+err.Error())
		return
	}
	act := actForExport.ExportActionDto()

	// fetch app
	appDTO, _ := controller.appService.FetchAppByID(teamID, appID)

	// fetch resource data
	rsc, errInGetRSC := controller.resourceService.GetResource(teamID, act.Resource)
	if errInGetRSC != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resource error: "+errInGetRSC.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:       auditlogger.AUDIT_LOG_RUN_ACTION,
		TeamID:          teamID,
		UserID:          -1,
		IP:              c.ClientIP(),
		AppID:           appID,
		AppName:         appDTO.Name,
		ResourceID:      act.Resource,
		ResourceName:    rsc.Name,
		ResourceType:    rsc.Type,
		ActionID:        act.ID,
		ActionName:      act.DisplayName,
		ActionParameter: act.Template,
	})

	// run
	actionRuntimeInfo := model.NewActionRuntimeInfo(team.ExportIDInString(), actForExport.ExportResourceID(), actForExport.ExportID(), "")
	res, err := controller.actionService.RunAction(teamID, act, actionRuntimeInfo)
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
				"errorFlag":    ERROR_FLAG_EXECUTE_ACTION_FAILED,
				"errorMessage": errors.New("SQL syntax error").Error(),
				"errorData": map[string]interface{}{
					"lineNumber": lineNumber,
					"message":    "SQL syntax error" + message,
				},
			})
			return
		}
		controller.FeedbackBadRequest(c, ERROR_FLAG_EXECUTE_ACTION_FAILED, "run action error: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}
