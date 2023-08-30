package controller

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"
)

func (controller *Controller) PublishAppToMarketplaceInternal(c *gin.Context) {
	// get user input params
	teamID, errInGetTeamID := controller.GetIntParamFromRequest(c, PARAM_TEAM_ID)
	teamIDInString, errInGetTeamInString := controller.GetStringParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAppID := controller.GetIntParamFromRequest(c, PARAM_APP_ID)
	appIDInString, errInGetAppIDInString := controller.GetStringParamFromRequest(c, PARAM_APP_ID)
	if errInGetTeamID != nil || errInGetTeamInString != nil || errInGetAppID != nil || errInGetAppIDInString != nil {
		return
	}

	// get request body
	req := request.NewPublishAppToMarketplaceInternalRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}
	userID := req.UserID

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// validate request data
	validated, errInValidate := controller.ValidateRequestTokenFromHeader(c, teamIDInString, appIDInString, req.ExportInJSONString())
	if !validated && errInValidate != nil {
		return
	}

	// retrieve by id
	app, err := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_PUBLISH_APP_TO_MARKETPLACE, "retrieve App by id error: "+err.Error())
		return
	}

	// publish
	// - publish will set app.Config.Public & app.Config.PublishedToMarketplace both to true
	// - unpublish will set app.Config.Public to false
	app.SetPublishedToMarketplace(req.PublishedToMarketplace, userID)

	// deploy app, publish an app need deploy it
	if req.PublishedToMarketplace {
		// release app version
		app.Release()

		// release app following components & actions
		// release will copy following units from edit version to app mainline version
		controller.DuplicateTreeStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
		controller.DuplicateKVStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
		controller.DuplicateSetStateByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
		controller.DuplicateActionByVersion(c, teamID, teamID, appID, appID, model.APP_EDIT_VERSION, app.ExportMainlineVersion(), userID)
	}

	// save
	errInUpdateAppByID := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateAppByID != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_PUBLISH_APP_TO_MARKETPLACE, "update App error: "+errInUpdateAppByID.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return
}

func (controller *Controller) GetAllAppListByIDInternal(c *gin.Context) {
	// get request body
	req := request.NewGetAppByIDsInternalRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// validate request data
	validated, errInValidate := controller.ValidateRequestTokenFromHeader(c, req.ExportInJSONString())
	if !validated && errInValidate != nil {
		return
	}

	// fetch
	apps, err := controller.Storage.AppStorage.RetrieveByIDs(req.IDs)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user error: "+err.Error())
		return
	}

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps(apps)

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewAppMapForExport(apps, usersLT))
	return
}

func (controller *Controller) GetReleaseVersionAppInternal(c *gin.Context) {
	// get user input params
	appID, errInGetAppID := controller.GetIntParamFromRequest(c, PARAM_APP_ID)
	appIDInString, errInGetAppIDInString := controller.GetStringParamFromRequest(c, PARAM_APP_ID)
	if errInGetAppID != nil || errInGetAppIDInString != nil {
		return
	}

	// validate request data
	validated, errInValidate := controller.ValidateRequestTokenFromHeader(c, appIDInString)
	if !validated && errInValidate != nil {
		return
	}

	// retrieve by id
	fullAppForExport, errInGetTargetVersionFullApp := controller.GetTargetVersionFullAppByAppID(c, appID, model.APP_AUTO_RELEASE_VERSION)
	if errInGetTargetVersionFullApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "retrieve App by id error: "+errInGetTargetVersionFullApp.Error())
		return
	}

	controller.FeedbackOK(c, fullAppForExport)
	return
}
