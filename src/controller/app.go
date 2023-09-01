package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/storage"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"
	"github.com/illacloud/builder-backend/src/utils/illamarketplacesdk"
)

func (controller *Controller) CreateApp(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
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
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_CREATE_APP,
	)
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
	errInCreateComponentTree := controller.BuildComponentTree(newApp, 0, componentTree)
	if errInCreateComponentTree != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_COMPONENT_TREE, "error in create component tree: "+errInCreateComponentTree.Error())
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
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	canDelete, errInCheckAttr := controller.AttributeGroup.CanDelete(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_DELETE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canDelete {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, err := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app error: "+err.Error())
		return
	}

	// init parallel counter
	var wg sync.WaitGroup
	wg.Add(1)

	// add counter to marketpalce
	go func() {
		defer wg.Done()
		marketplaceAPI := illamarketplacesdk.NewIllaMarketplaceRestAPI()
		marketplaceAPI.OpenDebug()
		errInDeleteProduct := marketplaceAPI.DeleteProduct(illamarketplacesdk.PRODUCT_TYPE_APPS, appID)
		if errInDeleteProduct != nil {
			log.Printf("[DUMP] DeleteApp.errInDeleteProduct: %+v\n", errInDeleteProduct)
		}
	}()

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
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_MANAGE_EDIT_APP,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
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
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "config app error: "+errInUpdateApp.Error())
		return
	}

	// update action public settings
	errInUpdatePublic := controller.Storage.ActionStorage.UpdatePrivacyByTeamIDAndAppIDAndUserID(teamID, appID, userID, app.IsPublic())
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
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// check if user is viewer (the viewer role can not access undeployed app aka "edit app" and have no ACTION_MANAGE_EDIT_APP attribute)
	canManage, errInCheckAttrManage := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_EDIT_APP,
	)
	if errInCheckAttrManage != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}

	// get all apps
	var allApps []*model.App
	var errInRetrieveAllApps error
	if canManage {
		allApps, errInRetrieveAllApps = controller.Storage.AppStorage.RetrieveByTeamIDOrderByUpdatedTime(teamID)
		fmt.Printf("[DUMP] canManage: %+v\n", canManage)
		fmt.Printf("[DUMP] RetrieveByTeamID.allApps: %+v\n", allApps)
	} else {
		allApps, errInRetrieveAllApps = controller.Storage.AppStorage.RetrieveDeployedAppByTeamID(teamID)
		fmt.Printf("[DUMP] canManage: %+v\n", canManage)
		fmt.Printf("[DUMP] RetrieveDeployedAppByTeamID.allApps: %+v\n", allApps)
	}
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
	c.JSON(http.StatusOK, response.GenerateGetAllAppsResponse(allApps, usersLT))
}

func (controller *Controller) SearchAppByKeywordsByPage(c *gin.Context) {
	// get user input params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	keywords, errInGetKeywords := controller.GetStringParamFromRequest(c, PARAM_KEYWORDS)
	limit, errInGetlimit := controller.GetIntParamFromRequest(c, PARAM_LIMIT)
	page, errInGetPage := controller.GetIntParamFromRequest(c, PARAM_PAGE)
	sortBy, errInGetSortBy := controller.GetStringParamFromRequest(c, PARAM_SORT_BY)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetKeywords != nil || errInGetSortBy != nil || errInGetAuthToken != nil || errInGetlimit != nil || errInGetPage != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get all apps
	var appList []*model.App
	var errInRetrieveAppList error
	pagination := storage.NewPagination(limit, page)

	appTotalRows, errInRetrieveAppCount := controller.Storage.AppStorage.CountByTeamIDAndKeywords(teamID, keywords)
	if errInRetrieveAppCount != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent tital count error: "+errInRetrieveAppCount.Error())
		return
	}
	pagination.CalculateTotalPagesByTotalRows(appTotalRows)

	// retrieve
	switch sortBy {
	case request.ORDER_BY_CREATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByCreatedAtDesc(teamID, keywords)
	case request.ORDER_BY_UPDATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByUpdatedAtDesc(teamID, keywords)
	default:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByCreatedAtDesc(teamID, keywords)
	}
	if errInRetrieveAppList != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent list error: "+errInRetrieveAppList.Error())
		return
	}

	// build user look up table
	usersLT := make(map[int]*model.User)
	if len(appList) > 0 {
		// get all modifier user ids from all apps
		allUserIDs := model.ExtractAllEditorIDFromApps(appList)

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	appListForExport, errInNewAppForExport := response.NewAppListResponse(appList, pagination, usersLT)
	if errInNewAppForExport != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "build AI-Agent list error: "+errInNewAppForExport.Error())
		return
	}
	controller.FeedbackOK(c, appListForExport)
}

func (controller *Controller) SearchAppByKeywordsByPageUsingURIParam(c *gin.Context) {
	// get user input params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	keywords, errInGetKeywords := controller.GetFirstStringParamValueFromURI(c, PARAM_KEYWORDS)
	limitString, errInGetlimit := controller.GetFirstStringParamValueFromURI(c, PARAM_LIMIT)
	pageString, errInGetPage := controller.GetFirstStringParamValueFromURI(c, PARAM_PAGE)
	sortBy, errInGetSortBy := controller.GetFirstStringParamValueFromURI(c, PARAM_SORT_BY)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetKeywords != nil || errInGetSortBy != nil || errInGetAuthToken != nil || errInGetlimit != nil || errInGetPage != nil {
		return
	}

	// convert param
	limit, _ := strconv.Atoi(limitString)
	page, _ := strconv.Atoi(pageString)

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get all apps
	var appList []*model.App
	var errInRetrieveAppList error
	pagination := storage.NewPagination(limit, page)

	appTotalRows, errInRetrieveAppCount := controller.Storage.AppStorage.CountByTeamIDAndKeywords(teamID, keywords)
	if errInRetrieveAppCount != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent tital count error: "+errInRetrieveAppCount.Error())
		return
	}
	pagination.CalculateTotalPagesByTotalRows(appTotalRows)

	// retrieve
	switch sortBy {
	case request.ORDER_BY_CREATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByCreatedAtDesc(teamID, keywords)
	case request.ORDER_BY_UPDATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByUpdatedAtDesc(teamID, keywords)
	default:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByKeywordsAndSortByCreatedAtDesc(teamID, keywords)
	}
	if errInRetrieveAppList != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent list error: "+errInRetrieveAppList.Error())
		return
	}

	// build user look up table
	usersLT := make(map[int]*model.User)
	if len(appList) > 0 {
		// get all modifier user ids from all apps
		allUserIDs := model.ExtractAllEditorIDFromApps(appList)

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	appListForExport, errInNewAppForExport := response.NewAppListResponse(appList, pagination, usersLT)
	if errInNewAppForExport != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "build AI-Agent list error: "+errInNewAppForExport.Error())
		return
	}
	controller.FeedbackOK(c, appListForExport)
}

func (controller *Controller) GetAllAppByPage(c *gin.Context) {
	// get user input params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	limitString, errInGetlimit := controller.GetFirstStringParamValueFromURI(c, PARAM_LIMIT)
	pageString, errInGetPage := controller.GetFirstStringParamValueFromURI(c, PARAM_PAGE)
	sortBy, errInGetSortBy := controller.GetFirstStringParamValueFromURI(c, PARAM_SORT_BY)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetSortBy != nil || errInGetAuthToken != nil || errInGetlimit != nil || errInGetPage != nil {
		return
	}

	// convert param
	limit, _ := strconv.Atoi(limitString)
	page, _ := strconv.Atoi(pageString)

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get all apps
	var appList []*model.App
	var errInRetrieveAppList error
	pagination := storage.NewPagination(limit, page)

	appTotalRows, errInRetrieveAppCount := controller.Storage.AppStorage.CountByTeamID(teamID)
	if errInRetrieveAppCount != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent tital count error: "+errInRetrieveAppCount.Error())
		return
	}
	pagination.CalculateTotalPagesByTotalRows(appTotalRows)

	// retrieve
	switch sortBy {
	case request.ORDER_BY_CREATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByTeamIDAndSortByCreatedAtDescByPage(teamID, pagination)
	case request.ORDER_BY_UPDATED_AT:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByTeamIDAndSortByUpdatedAtDescByPage(teamID, pagination)
	default:
		appList, errInRetrieveAppList = controller.Storage.AppStorage.RetrieveByTeamIDAndSortByCreatedAtDescByPage(teamID, pagination)
	}
	if errInRetrieveAppList != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve AI-Agent list error: "+errInRetrieveAppList.Error())
		return
	}

	// build user look up table
	usersLT := make(map[int]*model.User)
	if len(appList) > 0 {
		// get all modifier user ids from all apps
		allUserIDs := model.ExtractAllEditorIDFromApps(appList)

		// fet all user id mapped user info, and build user info lookup table
		var errInGetMultiUserInfo error
		usersLT, errInGetMultiUserInfo = datacontrol.GetMultiUserInfo(allUserIDs)
		if errInGetMultiUserInfo != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
			return
		}
	}

	// feedback
	appListForExport, errInNewAppForExport := response.NewAppListResponse(appList, pagination, usersLT)
	if errInNewAppForExport != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "build AI-Agent list error: "+errInNewAppForExport.Error())
		return
	}
	controller.FeedbackOK(c, appListForExport)
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
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	log.Printf("[DUMP] GetFullApp teamID: %+v, appID: %+v, version: %+v\n", teamID, appID, version)

	// do get app for editor method
	fullAppForExport, errInGenerateFullApp := controller.GetTargetVersionFullApp(c, teamID, appID, version, false)
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
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_MANAGE_EDIT_APP,
	)
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
	targetApp, errInRetrieveTargetApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
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
	// duplicate will copy following units from target app mainline version to duplicated app edit version
	controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, duplicatedAppID, targetApp.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, duplicatedAppID, targetApp.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, duplicatedAppID, targetApp.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateActionByVersion(c, teamID, teamID, appID, duplicatedAppID, targetApp.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)

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
	req := request.NewReleaseAppRequest()
	if errInDecode := json.NewDecoder(c.Request.Body).Decode(&req); errInDecode != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_RELEASE_APP, "release app error: "+errInDecode.Error())
		return
	}

	// validate
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_SPECIAL_RELEASE_APP)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// check team can release public app, the free team can not release app as public.
	// but when publish app to marketplace, the can re-deploy this app as public.
	if req.ExportPublic() && !app.IsPublishedToMarketplace() {
		canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(
			teamID,
			userAuthToken,
			accesscontrol.UNIT_TYPE_APP,
			appID,
			accesscontrol.ACTION_SPECIAL_RELEASE_PUBLIC_APP,
		)
		if errInCheckAttr != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
			return
		}
		if !canManageSpecial {
			controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
			return
		}
	}

	// config app & action public status
	if req.ExportPublic() {
		// deploy app as public
		app.SetPublic(userID)
		controller.Storage.ActionStorage.MakeActionPublicByTeamIDAndAppID(teamID, appID, userID)
	} else {
		// marketplace app can not published as private
		if app.IsPublishedToMarketplace() {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_RELEASE_APP, "this app already published to marketplace, can not make it private.")
			return
		}
		// deploy app as private
		app.SetPrivate(userID)
		controller.Storage.ActionStorage.MakeActionPrivateByTeamIDAndAppID(teamID, appID, userID)
	}

	// release app version
	treeStateLatestVersion, _ := controller.Storage.TreeStateStorage.RetrieveTreeStatesLatestVersion(teamID, appID)
	app.SyncMainlineVersionWithTreeStateLatestVersion(treeStateLatestVersion)
	app.Release()

	// update app for version bump, we should update app first in case create tree state failed with mismatch release & mainline version
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// release app following components & actions
	// release will copy following units from edit version to app mainline version
	errInDuplicateTreeStateByVersion := controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	errInDuplicateKVStateByVersion := controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	errInDuplicateSetStateByVersion := controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	errInDuplicateActionByVersion := controller.DuplicateActionByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	if errInDuplicateTreeStateByVersion != nil || errInDuplicateKVStateByVersion != nil || errInDuplicateSetStateByVersion != nil || errInDuplicateActionByVersion != nil {
		return
	}

	// if app already published to marketplace, sync app to marketplace
	if app.IsPublishedToMarketplace() {
		// init parallel counter
		var wg sync.WaitGroup
		wg.Add(1)

		// add counter to marketpalce
		go func() {
			defer wg.Done()
			marketplaceAPI := illamarketplacesdk.NewIllaMarketplaceRestAPI()
			marketplaceAPI.UpdateProduct(illamarketplacesdk.PRODUCT_TYPE_APPS, appID, illamarketplacesdk.NewAppForMarketplace(app))
		}()

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
	controller.FeedbackOK(c, response.NewReleaseAppResponse(app))
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
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTOIN_SPECIAL_TAKE_SNAPSHOT,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// config app version
	treeStateLatestVersion, _ := controller.Storage.TreeStateStorage.RetrieveTreeStatesLatestVersion(teamID, appID)
	app.SyncMainlineVersionWithTreeStateLatestVersion(treeStateLatestVersion)
	app.BumpMainlineVersionOverReleaseVersion()

	// update app for version bump, we should update app first in case create tree state failed with mismatch release & mainline version
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// do snapshot for app following components and actions
	// do snapshot will copy following units from edit version to app mainline version
	controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateActionByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)

	// save snapshot
	_, errInTakeSnapshot := controller.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), model.SNAPSHOT_TRIGGER_MODE_MANUAL)
	if errInTakeSnapshot != nil {
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
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// retrieve by page
	pagination := storage.NewPagination(pageLimit, page)
	snapshotTotalRows, errInRetrieveSnapshotCount := controller.Storage.AppSnapshotStorage.RetrieveCountByTeamIDAndAppID(teamID, appID)
	if errInRetrieveSnapshotCount != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshotCount.Error())
		return
	}
	pagination.CalculateTotalPagesByTotalRows(snapshotTotalRows)
	snapshots, errInRetrieveSnapshot := controller.Storage.AppSnapshotStorage.RetrieveByTeamIDAppIDAndPage(teamID, appID, pagination)
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
	controller.FeedbackOK(c, response.NewGetSnapshotListResponse(snapshots, pagination.GetTotalPages(), usersLT))
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
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get target snapshot
	snapshot, errInRetrieveSnapshot := controller.Storage.AppSnapshotStorage.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}

	// get app
	fullAppForExport, errInGenerateFullApp := controller.GetTargetVersionFullApp(c, teamID, appID, snapshot.ExportTargetVersion(), false)
	if errInGenerateFullApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get target version app failed: "+errInGenerateFullApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, fullAppForExport)
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
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		appID,
		accesscontrol.ACTOIN_SPECIAL_RECOVER_SNAPSHOT,
	)
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
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// bump app mainline versoin
	app.BumpMainlineVersion()

	// update app for version bump, we should update app first in case create tree state failed with mismatch release & mainline version
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app failed: "+errInUpdateApp.Error())
		return
	}

	// do snapshot for app following components and actions
	// do snapshot will copy following units from edit version to app mainline version
	controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	controller.DuplicateActionByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)

	// save app snapshot
	newAppSnapshot, errInTakeSnapshot := controller.SaveAppSnapshot(c, teamID, appID, userID, app.ExportMainlineVersion(), model.SNAPSHOT_TRIGGER_MODE_AUTO)
	if errInTakeSnapshot != nil {
		return
	}

	// phrase 2: clean edit version app following components & actions
	controller.Storage.TreeStateStorage.DeleteAllTypeTreeStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.KVStateStorage.DeleteAllTypeKVStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.SetStateStorage.DeleteAllTypeSetStatesByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)
	controller.Storage.ActionStorage.DeleteAllActionsByTeamIDAppIDAndVersion(teamID, appID, model.APP_EDIT_VERSION)

	// phrase 3: duplicate target version app data to edit version

	// get target snapshot
	targetSnapshot, errInRetrieveSnapshot := controller.Storage.AppSnapshotStorage.RetrieveByID(snapshotID)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return
	}
	targetVersion := targetSnapshot.ExportTargetVersion()

	// copy target version app following components & actions to edit version
	controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, appID, targetVersion, model.APP_EDIT_VERSION, userID)
	controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, appID, targetVersion, model.APP_EDIT_VERSION, userID)
	controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, appID, targetVersion, model.APP_EDIT_VERSION, userID)
	controller.DuplicateActionByVersion(c, teamID, teamID, appID, appID, targetVersion, model.APP_EDIT_VERSION, userID)

	// create a snapshot.ModifyHistory for recover snapshot
	modifyHistoryLog := model.NewRecoverAppSnapshotModifyHistory(userID, targetSnapshot)
	newAppSnapshot.PushModifyHistory(modifyHistoryLog)

	// update app snapshot
	errInUpdateSnapshot := controller.Storage.AppSnapshotStorage.UpdateWholeSnapshot(newAppSnapshot)
	if errInUpdateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update app snapshot failed: "+errInUpdateSnapshot.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return

}

func (controller *Controller) ForkMarketplaceApp(c *gin.Context) {
	// get user input params
	toTeamID, errInGetToTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TO_TEAM_ID)
	appID, errInGetAppID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetToTeamID != nil || errInGetUserID != nil || errInGetAppID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		toTeamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_APP,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_FORK_APP,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get app, only published ai-agent can be fork.
	app, errInRetrieve := controller.Storage.AppStorage.RetrieveByID(appID)
	if errInRetrieve != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_FORK_APP, "target app nost exists failed: "+errInRetrieve.Error())
		return
	}
	if !app.IsPublishedToMarketplace() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_FORK_APP, "permission denied")
		return
	}

	// create new app for duplicate
	duplicatedApp := model.NewAppForDuplicate(app, app.ExportAppName(), userID)
	duplicatedApp.InitForFork(toTeamID, userID)
	duplicatedAppID, errInCreateApp := controller.Storage.AppStorage.Create(duplicatedApp)
	if errInCreateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_FORK_APP, "error in create app: "+errInCreateApp.Error())
		return
	}

	// duplicate app following units
	// duplicate will copy following units from target app mainline version to duplicated app edit version
	controller.DuplicateTreeStateByVersion(c, app.ExportTeamID(), toTeamID, appID, duplicatedAppID, app.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateKVStateByVersion(c, app.ExportTeamID(), toTeamID, appID, duplicatedAppID, app.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateSetStateByVersion(c, app.ExportTeamID(), toTeamID, appID, duplicatedAppID, app.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)
	controller.DuplicateActionByVersion(c, app.ExportTeamID(), toTeamID, appID, duplicatedAppID, app.ExportMainlineVersion(), model.APP_EDIT_VERSION, userID)

	// init parallel counter
	var wg sync.WaitGroup
	wg.Add(1)

	// add counter to marketpalce
	go func() {
		defer wg.Done()
		marketplaceAPI := illamarketplacesdk.NewIllaMarketplaceRestAPI()
		marketplaceAPI.ForkCounter(illamarketplacesdk.PRODUCT_TYPE_APPS, appID)
	}()

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_CREATE_APP,
		TeamID:    toTeamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     duplicatedApp.ExportID(),
		AppName:   duplicatedApp.ExportAppName(),
	})

	// init app snapshot
	_, errInInitAppSnapSHot := controller.InitAppSnapshot(c, toTeamID, duplicatedApp.ExportID())
	if errInInitAppSnapSHot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "error in create app snapshot: "+errInInitAppSnapSHot.Error())
		return
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
}
