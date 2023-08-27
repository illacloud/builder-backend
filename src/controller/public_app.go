package controller

import (
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) GetFullPublicApp(c *gin.Context) {
	// fetch needed param
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

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(accesscontrol.ANONYMOUS_AUTH_TOKEN)
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

	// check if app is public app
	fullAppForExport, errInGenerateFullApp := controller.GetTargetVersionFullApp(c, teamID, appID, version, true)
	if errInGenerateFullApp != nil {
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_VIEW_APP,
		TeamID:    teamID,
		UserID:    -1,
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
		controller.FeedbackOK(c, model.NewIsPublicAppResponse(false))
		return
	}
	teamID := team.GetID()

	// check if app is public app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackOK(c, model.NewIsPublicAppResponse(false))
		return
	}
	if !app.IsPublic() {
		controller.FeedbackOK(c, model.NewIsPublicAppResponse(false))
		return
	}

	// feedback
	controller.FeedbackOK(c, model.NewIsPublicAppResponse(true))
	return
}
