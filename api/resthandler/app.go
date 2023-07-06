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
	Logger              *zap.SugaredLogger
	AttributeGroup      *ac.AttributeGroup
	AppService          app.AppService
	ActionService       action.ActionService
	TreeStateService    state.TreeStateService
	AppRepository       repository.AppRepository
	ActionRepository    repository.ActionRepository
	TreeStateRepository repository.TreeStateRepository
	KVStateRepository   repository.KVStateRepository
	SetStateRepository  repository.SetStateRepository
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
) *AppRestHandlerImpl {
	return &AppRestHandlerImpl{
		Logger:              Logger,
		AttributeGroup:      attrg,
		AppService:          AppService,
		ActionService:       ActionService,
		TreeStateService:    TreeStateService,
		AppRepository:       AppRepository,
		ActionRepository:    ActionRepository,
		TreeStateRepository: TreeStateRepository,
		KVStateRepository:   KVStateRepository,
		SetStateRepository:  SetStateRepository,
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

	// storage app
	_, errInCreateApp := impl.AppRepository.Create(newApp)
	if errInCreateApp != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// fill component node by given init schema
	// @NOTE: that the root node will created by InitScheme in request
	componentTree := repository.ConstructComponentNodeByMap(req.ExportAppName())
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

	// get all modifier user ids from all apps
	allUserIDs := repository.ExtractAllEditorIDFromApps(allApps)

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
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

	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error: "+errInRetrieveApp.Error())
		return
	}

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

	// check app version
	if app.ID == 0 || version > app.MainlineVersion {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error, app version invalied.")
		return
	}
	if version == repository.APP_AUTO_MAINLINE_VERSION {
		version = app.MainlineVersion
	}
	if version == repository.APP_AUTO_RELEASE_VERSION {
		version = app.ReleaseVersion
	}

	// form editor object field appForExport
	//
	// We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// get all modifier user ids from all apps
	allUserIDs := repository.ExtractAllEditorIDFromApps([]*repository.App{app})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	appForExport := repository.NewAppForExport(app, usersLT)

	// form editor object field actions
	actions, errInRetrieveActions := impl.ActionRepository.RetrieveActionsByAppVersion(teamID, appID, version)
	if errInRetrieveActions != nil {
		actions = []*repository.Action{}
	}
	actionsForExport := make([]*repository.ActionForExport, 0)
	for _, action := range actions {
		actionsForExport = append(actionsForExport, repository.NewActionForExport(action))
	}

	// form editor object field components
	treeStateComponents, errInRetrieveTreeStateComponents := impl.TreeStateRepository.RetrieveTreeStatesByApp(teamID, appID, repository.TREE_STATE_TYPE_COMPONENTS, version)
	if errInRetrieveTreeStateComponents != nil {
		treeStateComponents = []*repository.TreeState{}
	}
	treeStateLT := repository.BuildTreeStateLookupTable(treeStateComponents)
	rootOfTreeState := repository.PickUpTreeStatesRootNode(treeStateComponents)
	componentTree, _ := repository.BuildComponentTree(rootOfTreeState, treeStateLT, nil)

	// form editor object field dependenciesState
	dependenciesState := map[string][]string{}
	dependenciesKVStates, errInRetrieveDependenciesKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DEPENDENCIES, version)
	if errInRetrieveDependenciesKVStates != nil {
		dependenciesKVStates = []*repository.KVState{}
	}
	for _, dependency := range dependenciesKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dependency.Value), &revMsg)
		dependenciesState[dependency.Key] = revMsg // value convert to []string
	}

	// form editor object field dragShadowState
	dragShadowState := map[string]interface{}{}
	dragShadowKVStates, errInRetrieveDragShadowKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DRAG_SHADOW, version)
	if errInRetrieveDragShadowKVStates != nil {
		dragShadowKVStates = []*repository.KVState{}
	}
	for _, dragShadow := range dragShadowKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dragShadow.Value), &revMsg)
		dragShadowState[dragShadow.Key] = revMsg // value convert to []string
	}

	// form editor object field dottedLineSquareState
	dottedLineSquareState := map[string]interface{}{}
	dottedLineSquareKVStates, errInRetrieveDottedLineSquareKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE, version)
	if errInRetrieveDottedLineSquareKVStates != nil {
		dottedLineSquareKVStates = []*repository.KVState{}
	}
	for _, dottedLineSquare := range dottedLineSquareKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dottedLineSquare.Value), &revMsg)
		dottedLineSquareState[dottedLineSquare.Key] = revMsg // value convert to []string
	}

	// form editor object field displayNameState
	displayNameSetStates, errInRetrieveDisplayNameSetState := impl.SetStateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, version)
	if errInRetrieveDisplayNameSetState != nil {
		displayNameSetStates = []*repository.SetState{}
	}
	displayNameState := make([]string, 0, len(displayNameSetStates))
	for _, displayName := range displayNameSetStates {
		displayNameState = append(displayNameState, displayName.Value)
	}

	// finally, make a brand new editor object
	editorForExport := repository.NewEditorForExport(appForExport, actionsForExport, componentTree, dependenciesState, dragShadowState, dottedLineSquareState, displayNameState)

	// feedback
	FeedbackOK(c, editorForExport)
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
	appDTO, err := impl.AppService.FetchAppByID(teamID, appID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
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
		AppName:   appDTO.Name,
	})

	// Call `app service` to release app
	version, err := impl.AppService.ReleaseApp(teamID, appID, userID, publicApp)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_RELEASE_APP, "release app error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewReleaseAppResponse(version))
	return
}
