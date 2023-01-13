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

	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/pkg/action"

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
	logger         *zap.SugaredLogger
	actionService  action.ActionService
	AttributeGroup *ac.AttributeGroup
}

func NewActionRestHandlerImpl(logger *zap.SugaredLogger, actionService action.ActionService, attrg *ac.AttributeGroup) *ActionRestHandlerImpl {
	return &ActionRestHandlerImpl{
		logger:         logger,
		actionService:  actionService,
		AttributeGroup: attrg,
	}
}

func (impl ActionRestHandlerImpl) CreateAction(c *gin.Context) {
	// fetch payload
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}
	if err := impl.actionService.ValidateActionOptions(act.Type, act.Template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	appID, errInGetAPPID := GetIntParamFromRequest(PARAM_APP_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_CREATE_ACTION)
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

	// create
	act.App = appID
	act.Version = 0
	act.CreatedAt = time.Now().UTC()
	act.CreatedBy = userID
	act.UpdatedAt = time.Now().UTC()
	act.UpdatedBy = userID
	res, err := impl.actionService.CreateAction(act)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "create action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) UpdateAction(c *gin.Context) {
	// fetch payload
	var act action.ActionDto
	if err := json.NewDecoder(c.Request.Body).Decode(&act); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error" + err.Error(),
		})
		return
	}
	if err := impl.actionService.ValidateActionOptions(act.Type, act.Template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	appID, errInGetAPPID := GetIntParamFromRequest(PARAM_APP_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	actionID, errInGetActionID := GetIntParamFromRequest(PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(actionID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_ACTION)
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

	// update
	act.ID = actionID
	act.UpdatedBy = userID
	act.App = appID
	act.Version = 0
	act.UpdatedAt = time.Now().UTC()
	act.UpdatedBy = userID
	res, err := impl.actionService.UpdateAction(act)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update action error: " + err.Error(),
		})
		return
	}
	originInfo, _ := impl.actionService.GetAction(act.ID)
	res.CreatedBy = originInfo.CreatedBy
	res.CreatedAt = originInfo.CreatedAt

	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) DeleteAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	actionID, errInGetActionID := GetIntParamFromRequest(PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(actionID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanDelete(ac.ACTION_DELETE)
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

	// delete
	if err := impl.actionService.DeleteAction(actionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "delete action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"actionId": id,
	})
}

func (impl ActionRestHandlerImpl) GetAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	actionID, errInGetActionID := GetIntParamFromRequest(PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(actionID)
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
	res, err := impl.actionService.GetAction(actionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) FindActions(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	appID, errInGetAPPID := GetIntParamFromRequest(PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(appID)
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
	res, err := impl.actionService.FindActionsByAppVersion(appID, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get actions error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) PreviewAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_PREVIEW_ACTION)
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

	// execute
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
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "run action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (impl ActionRestHandlerImpl) RunAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetIntParamFromRequest(PARAM_TEAM_ID)
	actionID, errInGetActionID := GetIntParamFromRequest(PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_ACTION)
	impl.AttributeGroup.SetUnitID(actionID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_RUN_ACTION)
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

	// execute
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
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "run action error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, res)
}
