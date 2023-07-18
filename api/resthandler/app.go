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
	"fmt"
	"net/http"

	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/action"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

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
}

type AppRestHandlerImpl struct {
	Logger                *zap.SugaredLogger
	AttributeGroup        *ac.AttributeGroup
	AppService            app.AppService
	ActionService         action.ActionService
	TreeStateService      state.TreeStateService
	AppRepository         repository.AppRepository
	ActionRepository      repository.ActionRepository
	TreeStateRepository   repository.TreeStateRepository
	KVStateRepository     repository.KVStateRepository
	SetStateRepository    repository.SetStateRepository
	AppSnapshotRepository repository.AppSnapshotRepository
}

func NewAppRestHandlerImpl(Logger *zap.SugaredLogger,
	attrg *ac.AttributeGroup,
	AppService app.AppService,
	ActionService action.ActionService,
	TreeStateService state.TreeStateService,
	AppRepository repository.AppRepository,
	ActionRepository repository.ActionRepository,
	TreeStateRepository repository.TreeStateRepository,
	KVStateRepository repository.KVStateRepository,
	SetStateRepository repository.SetStateRepository,
	AppSnapshotRepository repository.AppSnapshotRepository,
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

func (impl AppRestHandlerImpl) CreateApp(c *gin.Context) {
	// Parse request body
	req := repository.NewCreateAppRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
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

	// construct app object
	newApp := repository.NewApp(req.ExportAppName(), teamID, userID)
	fmt.Printf("[DUMP] CreateApp.newApp: %+v\n", newApp)

	// storage app
	_, errInCreateApp := impl.AppRepository.Create(newApp)
	if errInCreateApp != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// fill component node by given init schema
	// @NOTE: that the root node will created by InitScheme in request
	componentTree := repository.ConstructComponentNodeByMap(req.ExportInitScheme())
	_ = impl.TreeStateService.CreateComponentTree(newApp, 0, componentTree)

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
	_, errInInitAppSnapSHot := impl.InitAppSnapshot(c, teamID, newApp.ExportID())
	if errInInitAppSnapSHot != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "error in create app snapshot: "+errInInitAppSnapSHot.Error())

	}

	// get all modifier user ids from all apps
	allUserIDs := repository.ExtractAllEditorIDFromApps([]*repository.App{newApp})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewAppForExport(newApp, usersLT))
	return
}

func (impl AppRestHandlerImpl) DeleteApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
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

	// fetch app
	app, err := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
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
	_ = impl.TreeStateRepository.DeleteAllTypeTreeStatesByApp(teamID, appID)
	_ = impl.KVStateRepository.DeleteAllTypeKVStatesByApp(teamID, appID)
	_ = impl.ActionRepository.DeleteActionsByApp(teamID, appID)
	_ = impl.SetStateRepository.DeleteAllTypeSetStatesByApp(teamID, appID)
	errInDeleteApp := impl.AppRepository.Delete(teamID, appID)
	if errInDeleteApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_DELETE_APP, "delete app error: "+errInDeleteApp.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewDeleteAppResponse(appID))
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

	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+errInRetrieveApp.Error())
		return
	}

	// update app config
	appConfig := app.ExportConfig()
	errInNewAppConfig := appConfig.UpdateAppConfigByConfigAppRawRequest(rawRequest)
	if errInNewAppConfig != nil {
		FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new app config failed: "+errInNewAppConfig.Error())
		return
	}

	// update app all field
	app.UpdateAppConfig(appConfig, userID)
	app.UpdateAppByConfigAppRawRequest(rawRequest) // for app name the field which not in config struct

	// execute update
	errInUpdateApp := impl.AppRepository.Update(app)
	if errInUpdateApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "config app error: "+errInUpdateApp.Error())
		return
	}

	// Call `action service` update action public config (the action follows the app config)
	actionConfig, errInNewActionConfig := repository.NewActionConfigByConfigAppRawRequest(rawRequest)
	if errInNewActionConfig != nil {
		FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new action config failed: "+errInNewActionConfig.Error())
		return
	}
	errInUpdatePublic := impl.ActionService.UpdatePublic(teamID, appID, userID, actionConfig)
	if errInUpdatePublic != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "config action error: "+errInUpdatePublic.Error())
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
	allUserIDs := repository.ExtractAllEditorIDFromApps([]*repository.App{app})

	fmt.Printf("[DUMP] allUserIDs: %+v\n", allUserIDs)

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewAppForExport(app, usersLT))
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

	// get all apps
	allApps, errInRetrieveAllApps := impl.AppRepository.RetrieveAllByUpdatedTime(teamID)
	if errInRetrieveAllApps != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get apps by team id failed: "+errInRetrieveAllApps.Error())
		return
	}

	// build user look up table
	usersLT := make(map[int]*repository.User)
	if len(allApps) > 0 {
		// get all modifier user ids from all apps
		allUserIDs := repository.ExtractAllEditorIDFromApps(allApps)
		fmt.Printf("[DUMP] GetAllApps.allUserIDs: %+v\n", allUserIDs)

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	c.JSON(http.StatusOK, repository.GenerateGetAllAppsResponse(allApps, usersLT))
}

func (impl AppRestHandlerImpl) GetMegaData(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	version, errInGetVersion := GetIntParamFromRequest(c, PARAM_VERSION)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetVersion != nil || errInGetAuthToken != nil || errInGetUserID != nil {
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

	// do get app for editor method
	app, _ := impl.GetTargetVersionApp(c, teamID, appID, version)

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
	req := repository.NewDuplicateAppRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// Validate request body
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// Call `app service` to duplicate app
	duplicatedAppID, errInDuplicateApp := impl.AppService.DuplicateApp(teamID, appID, userID, req.ExportAppName())
	if errInDuplicateApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_DUPLICATE_APP, "duplicate app error: "+errInDuplicateApp.Error())
		return
	}

	// get duplicated app
	duplicatedApp, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(duplicatedAppID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get user info failed: "+errInRetrieveApp.Error())
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
	_, errInInitAppSnapSHot := impl.InitAppSnapshot(c, teamID, duplicatedApp.ExportID())
	if errInInitAppSnapSHot != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "error in create app snapshot: "+errInInitAppSnapSHot.Error())

	}

	// get all modifier user ids from all apps
	allUserIDs := repository.ExtractAllEditorIDFromApps([]*repository.App{duplicatedApp})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewAppForExport(duplicatedApp, usersLT))
	return
}

func (impl AppRestHandlerImpl) ReleaseApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
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

	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// config app
	app.MainlineVersion += 1
	app.ReleaseVersion = app.MainlineVersion

	// release app following components & actions
	_ = impl.AppService.ReleaseTreeStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.AppService.ReleaseKVStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.AppService.ReleaseSetStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.AppService.ReleaseActionsByApp(teamID, appID, app.MainlineVersion)

	// config app & action public status
	if publicApp {
		app.SetPublic(userID)
		impl.ActionRepository.MakeActionPublicByTeamIDAndAppID(teamID, appID, userID)
	} else {
		app.SetPrivate(userID)
		impl.ActionRepository.MakeActionPrivateByTeamIDAndAppID(teamID, appID, userID)
	}

	// release app
	errInUpdateApp := impl.AppRepository.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
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
	FeedbackOK(c, repository.NewReleaseAppResponse(app.ReleaseVersion))
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

func (impl AppRestHandlerImpl) TakeSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
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

	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// config app
	app.BumpMainlineVersion()

	// do snapshot for app following components and actions
	if impl.SnapshotTreeState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotKVState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotSetState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotAction(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}

	// save snapshot
	_, errInTakeSnapshot := impl.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), repository.SNAPSHOT_TRIGGER_MODE_MANUAL)
	if errInTakeSnapshot != nil {
		return
	}

	// update app for version bump
	errInUpdateApp := impl.AppRepository.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// feedback
	FeedbackOK(c, nil)
	return

}

func (impl AppRestHandlerImpl) GetSnapshotList(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	pageLimit, errInGetPageLimit := GetIntParamFromRequest(c, PARAM_PAGE_LIMIT)
	page, errInGetPage := GetIntParamFromRequest(c, PARAM_PAGE)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetPageLimit != nil || errInGetPage != nil {
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

	// retrieve by page
	pagination := repository.NewPagiNation(pageLimit, page)
	snapshots, errInRetrieveSnapshot := impl.AppSnapshotRepository.RetrieveByTeamIDAppIDAndPage(teamID, appID, pagination)
	if errInRetrieveSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewGetSnapshotListResponse(snapshots))
	return

}

func (impl AppRestHandlerImpl) GetSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	snapshotID, errInGetSnapshotID := GetIntParamFromRequest(c, PARAM_SNAPSHOT_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetSnapshotID != nil || errInGetAuthToken != nil {
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

	// get target snapshot
	snapshot, errInRetrieveSnapshot := impl.AppSnapshotRepository.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}

	// get app
	impl.GetTargetVersionApp(c, teamID, appID, snapshot.ExportTargetVersion())
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

func (impl AppRestHandlerImpl) RecoverSnapshot(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	snapshotID, errInGetSnapshotID := GetMagicIntParamFromRequest(c, PARAM_SNAPSHOT_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetSnapshotID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
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

	// phrase 1: take snapshot for current edit version

	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// bump app mainline versoin
	app.BumpMainlineVersion()

	// do snapshot for app following components and actions
	if impl.SnapshotTreeState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotKVState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotSetState(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}
	if impl.SnapshotAction(c, teamID, appID, app.ExportMainlineVersion()) != nil {
		return
	}

	// save app snapshot
	newAppSnapshot, errInTakeSnapshot := impl.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), repository.SNAPSHOT_TRIGGER_MODE_MANUAL)
	if errInTakeSnapshot != nil {
		return
	}

	// update app version
	errInUpdateApp := impl.AppRepository.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// phrase 2: clean edit version app following components & actions
	impl.TreeStateRepository.DeleteAllTypeTreeStatesByTeamIDAppIDAndVersion(teamID, appID, repository.APP_EDIT_VERSION)
	impl.KVStateRepository.DeleteAllTypeKVStatesByTeamIDAppIDAndVersion(teamID, appID, repository.APP_EDIT_VERSION)
	impl.SetStateRepository.DeleteAllTypeSetStatesByTeamIDAppIDAndVersion(teamID, appID, repository.APP_EDIT_VERSION)
	impl.ActionRepository.DeleteAllActionsByTeamIDAppIDAndVersion(teamID, appID, repository.APP_EDIT_VERSION)

	// phrase 3: duplicate target version app data to edit version

	// get target snapshot
	targetSnapshot, errInRetrieveSnapshot := impl.AppSnapshotRepository.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}
	targetVersion := targetSnapshot.ExportTargetVersion()

	// copy target version app following components & actions to edit version
	impl.DuplicateTreeStateByVersion(c, teamID, appID, targetVersion, repository.APP_EDIT_VERSION)
	impl.DuplicateKVStateByVersion(c, teamID, appID, targetVersion, repository.APP_EDIT_VERSION)
	impl.DuplicateSetStateByVersion(c, teamID, appID, targetVersion, repository.APP_EDIT_VERSION)
	impl.DuplicateActionByVersion(c, teamID, appID, targetVersion, repository.APP_EDIT_VERSION)

	// create a snapshot.ModifyHistory for recover snapshot
	modifyHistoryLog := repository.NewRecoverAppSnapshotModifyHistory(userID)
	newAppSnapshot.PushModifyHistory(modifyHistoryLog)

	// update app snapshot
	errInUpdateSnapshot := impl.AppSnapshotRepository.UpdateWholeSnapshot(newAppSnapshot)
	if errInUpdateSnapshot != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update app snapshot failed: "+errInUpdateSnapshot.Error())
		return
	}

	// feedback
	FeedbackOK(c, nil)
	return

}
