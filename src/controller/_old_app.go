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
	"fmt"
	"log"
	"net/http"

	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/pkg/action"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"
	"github.com/illacloud/illa-cloud-backend/src/accesscontrol"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AppRestHandler interface {
	CreateApp(c *gin.Context)
	DeleteApp(c *gin.Context)
	ConfigApp(c *gin.Context)
	GetAllApps(c *gin.Context)
	GetMegaData(c *gin.Context)
	DuplicateApp(c *gin.Context)
	ReleaseApp(c *gin.Context)
	TakeSnapshot(c *gin.Context)
	GetSnapshotList(c *gin.Context)
	GetSnapshot(c *gin.Context)
	RecoverSnapshot(c *gin.Context)
}

type AppRestHandlerImpl struct {
	Logger                *zap.SugaredLogger
	AttributeGroup        *accesscontrol.AttributeGroup
	AppService            app.AppService
	ActionService         action.ActionService
	TreeStateService      state.TreeStateService
	AppRepository         model.AppRepository
	ActionRepository      model.ActionRepository
	TreeStateRepository   model.TreeStateRepository
	KVStateRepository     model.KVStateRepository
	SetStateRepository    model.SetStateRepository
	AppSnapshotRepository model.AppSnapshotRepository
}

func NewAppRestHandlerImpl(Logger *zap.SugaredLogger,
	attrg *accesscontrol.AttributeGroup,
	AppService app.AppService,
	ActionService action.ActionService,
	TreeStateService state.TreeStateService,
	AppRepository model.AppRepository,
	ActionRepository model.ActionRepository,
	TreeStateRepository model.TreeStateRepository,
	KVStateRepository model.KVStateRepository,
	SetStateRepository model.SetStateRepository,
	AppSnapshotRepository model.AppSnapshotRepository,
) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		Logger:                Logger,
		AttributeGroup:        attrg,
		AppService:            AppService,
		ActionService:         ActionService,
		TreeStateService:      TreeStateService,
		AppRepository:         AppRepository,
		ActionRepository:      ActionRepository,
		TreeStateRepository:   TreeStateRepository,
		KVStateRepository:     KVStateRepository,
		SetStateRepository:    SetStateRepository,
		AppSnapshotRepository: AppSnapshotRepository,
	}
}

func (controller *Controller) CreateApp(c *gin.Context) {
	// Parse request body
	req := model.NewCreateAppRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(accesscontrol.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_CREATE_APP)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// construct app object
	newApp := model.NewApp(req.ExportAppName(), teamID, userID)
	fmt.Printf("[DUMP] CreateApp.newApp: %+v\n", newApp)

	// storage app
	_, errInCreateApp := controller.Storage.AppStorage.Create(newApp)
	if errInCreateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// fill component node by given init schema
	// @NOTE: that the root node will created by InitScheme in request
	componentTree := model.ConstructComponentNodeByMap(req.ExportInitScheme())
	_ = controller.TreeStateService.CreateComponentTree(newApp, 0, componentTree)

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_CREATE_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     newApp.ExportID(),
		AppName:   newApp.ExportAppName(),
	})

	// init app snapshot
	_, errInInitAppSnapSHot := controller.InitAppSnapshot(c, teamID, newApp.ExportID())
	if errInInitAppSnapSHot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "error in create app snapshot: "+errInInitAppSnapSHot.Error())

	}

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps([]*model.App{newApp})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewAppForExport(newApp, usersLT))
	return
}

func (controller *Controller) DeleteApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canDelete, errInCheckAttr := controller.AttributeGroup.CanDelete(accesscontrol.ACTION_DELETE)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canDelete {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, err := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_DELETE_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     appID,
		AppName:   app.ExportAppName(),
	})

	// delete app related states and action
	_ = controller.Storage.TreeStateStorage.DeleteAllTypeTreeStatesByApp(teamID, appID)
	_ = controller.Storage.KVStateStorage.DeleteAllTypeKVStatesByApp(teamID, appID)
	_ = controller.Storage.ActionStorage.DeleteActionsByApp(teamID, appID)
	_ = controller.Storage.SetStateStorage.DeleteAllTypeSetStatesByApp(teamID, appID)
	errInDeleteApp := controller.Storage.AppStorage.Delete(teamID, appID)
	if errInDeleteApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_DELETE_APP, "delete app error: "+errInDeleteApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewDeleteAppResponse(appID))
	return
}

func (controller *Controller) ConfigApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// get request body
	var rawRequest map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&rawRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_EDIT_APP)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+errInRetrieveApp.Error())
		return
	}

	// update app config
	appConfig := app.ExportConfig()
	errInNewAppConfig := appConfig.UpdateAppConfigByConfigAppRawRequest(rawRequest)
	if errInNewAppConfig != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new app config failed: "+errInNewAppConfig.Error())
		return
	}

	// update app all field
	app.UpdateAppConfig(appConfig, userID)
	app.UpdateAppByConfigAppRawRequest(rawRequest) // for app name the field which not in config struct

	// execute update
	errInUpdateApp := controller.Storage.AppStorage.Update(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "config app error: "+errInUpdateApp.Error())
		return
	}

	// Call `action service` update action public config (the action follows the app config)
	actionConfig, errInNewActionConfig := model.NewActionConfigByConfigAppRawRequest(rawRequest)
	if errInNewActionConfig != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new action config failed: "+errInNewActionConfig.Error())
		return
	}
	errInUpdatePublic := controller.ActionService.UpdatePublic(teamID, appID, userID, actionConfig)
	if errInUpdatePublic != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "config action error: "+errInUpdatePublic.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_EDIT_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     appID,
		AppName:   app.ExportAppName(),
	})

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps([]*model.App{app})

	fmt.Printf("[DUMP] allUserIDs: %+v\n", allUserIDs)

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewAppForExport(app, usersLT))
	return
}

func (controller *Controller) GetAllApps(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(accesscontrol.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(accesscontrol.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get all apps
	allApps, errInRetrieveAllApps := controller.Storage.AppStorage.RetrieveAllByUpdatedTime(teamID)
	if errInRetrieveAllApps != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get apps by team id failed: "+errInRetrieveAllApps.Error())
		return
	}

	// build user look up table
	usersLT := make(map[int]*model.User)
	if len(allApps) > 0 {
		// get all modifier user ids from all apps
		allUserIDs := model.ExtractAllEditorIDFromApps(allApps)
		fmt.Printf("[DUMP] GetAllApps.allUserIDs: %+v\n", allUserIDs)

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	c.JSON(http.StatusOK, model.GenerateGetAllAppsResponse(allApps, usersLT))
}

func (controller *Controller) GetMegaData(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	version, errInGetVersion := controller.GetIntParamFromRequest(c, PARAM_VERSION)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetVersion != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(accesscontrol.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// do get app for editor method
	app, _ := controller.GetTargetVersionApp(c, teamID, appID, version)

	// audit log
	eventType := auditlogger.AUDIT_LOG_VIEW_APP
	if version == 0 {
		eventType = auditlogger.AUDIT_LOG_EDIT_APP
	}
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: eventType,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     appID,
		AppName:   app.ExportAppName(),
	})

}

func (controller *Controller) DuplicateApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_EDIT_APP)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// Parse request body
	req := model.NewDuplicateAppRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// Call `app service` to duplicate app
	duplicatedAppID, errInDuplicateApp := controller.AppService.DuplicateApp(teamID, appID, userID, req.ExportAppName())
	if errInDuplicateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_DUPLICATE_APP, "duplicate app error: "+errInDuplicateApp.Error())
		return
	}

	// get duplicated app
	duplicatedApp, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(duplicatedAppID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get user info failed: "+errInRetrieveApp.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_CREATE_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     duplicatedApp.ExportID(),
		AppName:   duplicatedApp.ExportAppName(),
	})

	// init app snapshot
	_, errInInitAppSnapSHot := controller.InitAppSnapshot(c, teamID, duplicatedApp.ExportID())
	if errInInitAppSnapSHot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "error in create app snapshot: "+errInInitAppSnapSHot.Error())

	}

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps([]*model.App{duplicatedApp})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewAppForExport(duplicatedApp, usersLT))
	return
}

func (controller *Controller) ReleaseApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// get request body
	var rawRequest map[string]interface{}
	publicApp := false
	json.NewDecoder(c.Request.Body).Decode(&rawRequest)
	fmt.Printf("[DUMP] ReleaseApp.rawRequest: %+v\n", rawRequest)
	isPublicRaw, hitIsPublic := rawRequest["public"]
	if hitIsPublic {
		publicApp = isPublicRaw.(bool)
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(accesscontrol.ACTION_SPECIAL_RELEASE_APP)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// check team can release public app
	if publicApp {
		canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(accesscontrol.ACTION_SPECIAL_RELEASE_PUBLIC_APP)
		if errInCheckAttr != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
			return
		}
		if !canManageSpecial {
			controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
			return
		}
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// config app
	app.MainlineVersion += 1
	app.ReleaseVersion = app.MainlineVersion

	// release app following components & actions
	_ = controller.AppService.ReleaseTreeStateByApp(teamID, appID, app.MainlineVersion)
	_ = controller.AppService.ReleaseKVStateByApp(teamID, appID, app.MainlineVersion)
	_ = controller.AppService.ReleaseSetStateByApp(teamID, appID, app.MainlineVersion)
	_ = controller.AppService.ReleaseActionsByApp(teamID, appID, app.MainlineVersion)

	// config app & action public status
	if publicApp {
		app.SetPublic(userID)
		controller.Storage.ActionStorage.MakeActionPublicByTeamIDAndAppID(teamID, appID, userID)
	} else {
		app.SetPrivate(userID)
		controller.Storage.ActionStorage.MakeActionPrivateByTeamIDAndAppID(teamID, appID, userID)
	}

	// release app
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_DEPLOY_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     appID,
		AppName:   app.ExportAppName(),
	})

	// feedback
	controller.FeedbackOK(c, model.NewReleaseAppResponse(app.ReleaseVersion))
	return
}

// For take snapshot for app, we should:
// - get target app
// - bump target app mainline version
// - snapshot all following components & actions by app mainline version
//   - snapshot tree state
//   - snapshot k-v state
//   - snapshot set state
//   - snapshot actions
// - save app snapshot
// - update app for version bump

func (controller *Controller) TakeSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(accesscontrol.ACTOIN_SPECIAL_TAKE_SNAPSHOT)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// config app version
	treeStateLatestVersion, _ := controller.Storage.TreeStateStorage.RetrieveTreeStatesLatestVersion(teamID, appID)
	app.SyncMainlineVersoinWithTreeStateLatestVersion(treeStateLatestVersion)
	app.BumpMainlineVersionOverReleaseVersoin()
	log.Printf("[DUMP] app.MainlineVersion: %d, app.ReleaseVersion: %d, treeStateLatestVersion: %d\n", app.MainlineVersion, app.ReleaseVersion, treeStateLatestVersion)

	// do snapshot for app following components and actions
	if controller.SnapshotTreeState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotKVState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotSetState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotAction(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}

	// save snapshot
	_, errInTakeSnapshot := controller.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), model.SNAPSHOT_TRIGGER_MODE_MANUAL)
	if errInTakeSnapshot != nil {
		return
	}

	// update app for version bump
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return

}

func (controller *Controller) GetSnapshotList(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	pageLimit, errInGetPageLimit := controller.GetIntParamFromRequest(c, PARAM_PAGE_LIMIT)
	page, errInGetPage := controller.GetIntParamFromRequest(c, PARAM_PAGE)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetPageLimit != nil || errInGetPage != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(accesscontrol.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// retrieve by page
	pagination := model.NewPagiNation(pageLimit, page)
	snapshotTotalRows, errInRetrieveSnapshotCount := controller.AppSnapshotmodel.RetrieveCountByTeamIDAndAppID(teamID, appID)
	if errInRetrieveSnapshotCount != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshotCount.Error())
		return
	}
	pagination.CalculateTotalPagesByTotalRows(snapshotTotalRows)
	snapshots, errInRetrieveSnapshot := controller.AppSnapshotmodel.RetrieveByTeamIDAppIDAndPage(teamID, appID, pagination)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllModifierIDFromAppSnapshot(snapshots)

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewGetSnapshotListResponse(snapshots, pagination.GetTotalPages(), usersLT))
	return

}

func (controller *Controller) GetSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	snapshotID, errInGetSnapshotID := controller.GetMagicIntParamFromRequest(c, PARAM_SNAPSHOT_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetSnapshotID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(accesscontrol.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get target snapshot
	snapshot, errInRetrieveSnapshot := controller.AppSnapshotmodel.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}

	// get app
	controller.GetTargetVersionApp(c, teamID, appID, snapshot.ExportTargetVersion())
	return
}

// For recover snapshot for app, we should:
// - phrase 1: take snapshot for current edit version
//   - get target app
//   - bump app mainline versoin
//   - snapshot all following components & actions by app mainline version
//     - snapshot tree state
//     - snapshot k-v state
//     - snapshot set state
//     - snapshot actions
//   - save app snapshot
//   - update app for version bump
// - phrase 2: clean edit version app following components & actions
//   - clean edit version tree state
//   - clean edit version k-v state
//   - clean edit version set state
//   - clean edit version actions
// - phrase 3: duplicate target version app data to edit version
//   - get target snapshot for export target app version
//   - copy target version app following components & actions to edit version
//     - copy target version tree state to edit version
//     - copy target version k-v state to edit version
//     - copy target version set state to edit version
//     - copy target version actions to edit version
//   - create a snapshot.ModifyHistory for recover snapshot

func (controller *Controller) RecoverSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	snapshotID, errInGetSnapshotID := controller.GetMagicIntParamFromRequest(c, PARAM_SNAPSHOT_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetSnapshotID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_APP)
	controller.AttributeGroup.SetUnitID(appID)
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(accesscontrol.ACTOIN_SPECIAL_RECOVER_SNAPSHOT)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// phrase 1: take snapshot for current edit version

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// bump app mainline versoin
	app.BumpMainlineVersion()

	// do snapshot for app following components and actions
	if controller.SnapshotTreeState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotKVState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotSetState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if controller.SnapshotAction(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}

	// save app snapshot
	newAppSnapshot, errInTakeSnapshot := controller.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), model.SNAPSHOT_TRIGGER_MODE_AUTO)
	if errInTakeSnapshot != nil {
		return
	}

	// update app version
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// phrase 2: clean edit version app following components & actions
	controller.Storage.TreeStateStorage.DeleteAllTypeTreeStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.KVStateStorage.DeleteAllTypeKVStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.SetStateStorage.DeleteAllTypeSetStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.ActionStorage.DeleteAllActionsByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)

	// phrase 3: duplicate target version app data to edit version

	// get target snapshot
	targetSnapshot, errInRetrieveSnapshot := controller.AppSnapshotmodel.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}
	targetVersion := targetSnapshot.ExportTargetVersion()

	// copy target version app following components & actions to edit version
	controller.DuplicateTreeStateByVersion(c, teamID, appID, targetVersion, model.APP_EDIT_VERSION)
	controller.DuplicateKVStateByVersion(c, teamID, appID, targetVersion, model.APP_EDIT_VERSION)
	controller.DuplicateSetStateByVersion(c, teamID, appID, targetVersion, model.APP_EDIT_VERSION)
	controller.DuplicateActionByVersion(c, teamID, appID, targetVersion, model.APP_EDIT_VERSION)

	// create a snapshot.ModifyHistory for recover snapshot
	modifyHistoryLog := model.NewRecoverAppSnapshotModifyHistory(userID, targetSnapshot)
	newAppSnapshot.PushModifyHistory(modifyHistoryLog)

	// update app snapshot
	errInUpdateSnapshot := controller.AppSnapshotmodel.UpdateWholeSnapshot(newAppSnapshot)
	if errInUpdateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update app snapshot failed: "+errInUpdateSnapshot.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return

}
