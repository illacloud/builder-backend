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

	ac "github.com/illacloud/illa-builder-backend/internal/accesscontrol"
	"github.com/illacloud/illa-builder-backend/internal/repository"
	"github.com/illacloud/illa-builder-backend/pkg/action"
	"github.com/illacloud/illa-builder-backend/pkg/app"
	"github.com/illacloud/illa-builder-backend/pkg/state"

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
	ConfigApp(c *gin.Context)
	GetAllApps(c *gin.Context)
	GetMegaData(c *gin.Context)
	DuplicateApp(c *gin.Context)
	ReleaseApp(c *gin.Context)
}

type AppRestHandlerImpl struct {
	logger           *zap.SugaredLogger
	appService       app.AppService
	actionService    action.ActionService
	AttributeGroup   *ac.AttributeGroup
	treeStateService state.TreeStateService
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService, actionService action.ActionService, attrg *ac.AttributeGroup, treeStateService state.TreeStateService) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		logger:           logger,
		appService:       appService,
		actionService:    actionService,
		AttributeGroup:   attrg,
		treeStateService: treeStateService,
	}
}

func (impl AppRestHandlerImpl) CreateApp(c *gin.Context) {
	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_CREATE_APP)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	appDto := app.AppDto{
		TeamID:    teamID,
		Name:      payload.Name,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	appDto.InitUID()

	// Call `app service` create app
	res, err := impl.appService.CreateApp(appDto)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "create app error: "+err.Error())
		return
	}

	if len(payload.InitScheme) > 0 {
		for _, v := range payload.InitScheme {
			componentTree := repository.ConstructComponentNodeByMap(v)
			_ = impl.treeStateService.CreateComponentTree(&res, 0, componentTree)
		}
	}

	// feedback
	FeedbackOK(c, app.NewAppDtoForExport(&res))
	return
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
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
	canDelete, errInCheckAttr := impl.AttributeGroup.CanDelete(ac.ACTION_DELETE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canDelete {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Call `app service` delete app
	if err := impl.appService.DeleteApp(teamID, appID); err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_DELETE_APP, "delete app error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewDeleteAppResponse(appID))
	return
}

func (impl AppRestHandlerImpl) RenameApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(appID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_APP)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// Call `app service` update app
	appDTO, err := impl.appService.FetchAppByID(teamID, appID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
		return
	}
	appDTO.Name = payload.Name
	appDTO.UpdatedBy = userID
	res, err := impl.appService.UpdateApp(appDTO)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "rename app error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, res)
	return
}

func (impl AppRestHandlerImpl) ConfigApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// get request body
	var rawRequest map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&rawRequest); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(appID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_APP)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// update app config
	appConfig, errInNewAppConfig := repository.NewAppConfigByConfigAppRawRequest(rawRequest)
	if errInNewAppConfig != nil {
		FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new app config failed: "+errInNewAppConfig.Error())
		return
	}

	// Call `app service` update app
	appDTO, err := impl.appService.FetchAppByID(teamID, appID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
		return
	}
	appDTO.UpdateAppDTOConfig(appConfig, userID)
	res, err := impl.appService.UpdateApp(appDTO)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "config app error: "+err.Error())
		return
	}

	// Call `action service` update action public config (the action follows the app config)
	actionConfig, errInNewActionConfig := repository.NewActionConfigByConfigAppRawRequest(rawRequest)
	if errInNewActionConfig != nil {
		FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new action config failed: "+errInNewActionConfig.Error())
		return
	}
	errInUpdatePublic := impl.actionService.UpdatePublic(teamID, appID, userID, actionConfig)
	if errInUpdatePublic != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "config action error: "+errInUpdatePublic.Error())
		return
	}
	// feedback
	FeedbackOK(c, res)
	return
}

func (impl AppRestHandlerImpl) GetAllApps(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Call `app service` get all apps
	res, err := impl.appService.GetAllApps(teamID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "get all apps error: "+err.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, res)
}

func (impl AppRestHandlerImpl) GetMegaData(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	version, errInGetVersion := GetIntParamFromRequest(c, PARAM_VERSION)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetVersion != nil || errInGetAuthToken != nil {
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
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Fetch Mega data via `app` and `version`
	res, err := impl.appService.GetMegaData(teamID, appID, version)
	if err != nil {
		if err.Error() == "content not found" {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error: "+err.Error())
			return
		}
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, res)
	return
}

func (impl AppRestHandlerImpl) DuplicateApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_APP)
	impl.AttributeGroup.SetUnitID(appID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_APP)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Parse request body
	var payload AppRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// Call `app service` to duplicate app
	res, err := impl.appService.DuplicateApp(teamID, appID, userID, payload.Name)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_DUPLICATE_APP, "duplicate app error: "+err.Error())
		return
	}
	// feedback
	FeedbackOK(c, res)
	return
}

func (impl AppRestHandlerImpl) ReleaseApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
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
	canManageSpecial, errInCheckAttr := impl.AttributeGroup.CanManageSpecial(ac.ACTION_SPECIAL_RELEASE_APP)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Call `app service` to release app
	version, err := impl.appService.ReleaseApp(teamID, appID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_RELEASE_APP, "release app error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewReleaseAppResponse(version))
	return
}
