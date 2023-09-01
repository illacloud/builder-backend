package controller

import (
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) GetFullPublicApp(c *gin.Context) {
	// fetch needed param
	userID := model.ANONYMOUS_USER_ID
	userAuthToken := accesscontrol.ANONYMOUS_AUTH_TOKEN
	teamIdentifier, errInGetTeamIdentifier := controller.GetStringParamFromRequest(c, PARAM_TEAM_IDENTIFIER)
	publicAppID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	version, errInGetVersion := controller.GetIntParamFromRequest(c, PARAM_VERSION)
	if errInGetTeamIdentifier != nil || errInGetAPPID != nil || errInGetVersion != nil {
		return
	}

	// check version, the version must be model.APP_AUTO_RELEASE_VERSION
	if version != model.APP_AUTO_RELEASE_VERSION {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you only can access release version of app.")
		return
	}

	// get team id by team teamIdentifier
	team, errInGetTeamInfo := datacontrol.GetTeamInfoByIdentifier(teamIdentifier)
	if errInGetTeamInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_TEAM, "get target team by identifier error: "+errInGetTeamInfo.Error())
		return
	}
	teamID := team.GetID()

	// get app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, publicAppID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get target app by id error: "+errInRetrieveApp.Error())
		return
	}

	// check if app publishedToMarketplace, if it is, everyone can access app and do not need authorization and access controll check
	if !app.IsPublishedToMarketplace() {
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
	}

	// check if app is public app
	fullAppForExport, errInGenerateFullApp := controller.GetTargetVersionFullApp(c, teamID, publicAppID, version, true)
	if errInGenerateFullApp != nil {
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_VIEW_APP,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		AppID:     publicAppID,
		AppName:   fullAppForExport.AppInfo.Name,
	})

	// feedback
	controller.FeedbackOK(c, fullAppForExport)
	return
}

func (controller *Controller) IsPublicApp(c *gin.Context) {
	// fetch needed param
	teamIdentifier, errInGetTeamIdentifier := controller.GetStringParamFromRequest(c, PARAM_TEAM_IDENTIFIER)
	publicAppID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	if errInGetTeamIdentifier != nil || errInGetAPPID != nil {
		return
	}

	// get team id by team teamIdentifier
	team, errInGetTeamInfo := datacontrol.GetTeamInfoByIdentifier(teamIdentifier)
	if errInGetTeamInfo != nil {
		controller.FeedbackOK(c, response.NewIsPublicAppResponse(false))
		return
	}
	teamID := team.GetID()
	fmt.Printf("[DUMP] IsPublicApp.teamIdentifier: %+v, teamID: %+v, IsPublicApp.publicAppID: %+v \n", teamIdentifier, teamID, publicAppID)
	fmt.Printf("[DUMP] IsPublicApp.team: %+v\n", team)

	// check if app is public app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, publicAppID)
	fmt.Printf("[DUMP] IsPublicApp.app: %+v\n", app)
	fmt.Printf("[DUMP] IsPublicApp.errInRetrieveApp: %+v\n", errInRetrieveApp)
	if errInRetrieveApp != nil {
		controller.FeedbackOK(c, response.NewIsPublicAppResponse(false))
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewIsPublicAppResponse(app.IsPublic()))
	return
}
