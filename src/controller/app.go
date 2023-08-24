package controller

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"
)

func (controller *Controller) CreateApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := controller.controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// Parse request body
	req := request.NewCreateAppRequest()
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
	newApp := model.NewAppByCreateAppRequest(req.ExportAppName(), teamID, userID)

	// create app
	_, errInCreateApp := controller.Storage.AppStorage.Create(newApp)
	if errInCreateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// fill component node by given init schema
	// @NOTE: the root node will created by InitScheme in request
	componentTree := model.ConstructComponentNodeByMap(req.ExportInitScheme())
	errInCreateComponentTree := controller.controller.CreateComponentTree(newApp, 0, componentTree)
	if errInCreateComponentTree != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CREATE_COMPONENT_TREE_FAILED, "error in create component tree: "+errInCreateComponentTree.Error())
		return
	}

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
	teamID, errInGetTeamID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := controller.controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.controller.GetUserIDFromAuth(c)
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
	_ = controller.Storage.AppSnapshotStorage.DeleteAllAppSnapshotByTeamIDAndAppID(teamID, appID)
	errInDeleteApp := controller.Storage.AppStorage.Delete(teamID, appID)
	if errInDeleteApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_DELETE_APP, "delete app error: "+errInDeleteApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewDeleteAppResponse(appID))
	return
}

func (controller *Controller) ConfigApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.controller.GetUserAuthTokenFromHeader(c)
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

	// update app field & app config
	appConfig := app.ExportConfig()
	errInNewAppConfig := appConfig.UpdateAppConfigByConfigAppRawRequest(rawRequest)
	if errInNewAppConfig != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new app config failed: "+errInNewAppConfig.Error())
		return
	}
	app.UpdateAppConfig(appConfig, userID)
	app.UpdateAppByConfigAppRawRequest(rawRequest) // for "app name" field which not in config struct

	// update app
	errInUpdateApp := controller.Storage.AppStorage.Update(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "config app error: "+errInUpdateApp.Error())
		return
	}

	// update action public settings
	actionConfig, errInNewActionConfig := model.NewActionConfigByConfigAppRawRequest(rawRequest)
	if errInNewActionConfig != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_BUILD_APP_CONFIG_FAILED, "new action config failed: "+errInNewActionConfig.Error())
		return
	}
	errInUpdatePublic := controller.Storage.ActionStorage.UpdatePublicByTeamIDAndAppIDAndUserID(teamID, appID, userID, actionConfig)
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
	teamID, errInGetTeamID := controller.controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.controller.GetUserAuthTokenFromHeader(c)
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

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	controller.FeedbackOK(c, response.GenerateGetAllAppsResponse(allApps, usersLT))
}

func (controller *Controller) GetFullApp(c *gin.Context) {
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
	fullAppForExport, errInGenerateFullApp := controller.GetTargetVersionFullApp(c, teamID, appID, version)
	if errInGenerateFullApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "build full app failed: "+errInGenerateFullApp.Error())
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
		AppName:   fullAppForExport.ExportAppName(),
	})

	// feedback
	controller.FeedbackOK(c, fullAppForExport)
	return
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
	req := request.NewDuplicateAppRequest()
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

	// get target app
	targetApp, errInRetrieveTargetApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveTargetApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveTargetApp.Error())
		return
	}

	// create new app for duplicate
	duplicatedApp := model.NewAppForDuplicate(targetApp, req.ExportAppName(), userID)
	duplicatedAppID, errInCreateApp := controller.Storage.AppStorage.Create(duplicatedApp)
	if errInCreateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// duplicate app following units
	// - TreeState
	// - KVState
	// - SetState
	// - Action
	controller.DuplicateTreeStates(teamID, appID, duplicatedAppID, userID)
	controller.DuplicateKVStates(teamID, appID, duplicatedAppID, userID)
	controller.DuplicateSetStates(teamID, appID, duplicatedAppID, userID)
	controller.DuplicateActions(teamID, appID, duplicatedAppID, userID)

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
