package controller

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/request"
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
	app.SetPublishedToMarketplace(req.PublishedToMarketplace, req.UserID)

	// save
	errInUpdateAppByID := controller.Storage.AppStorage.Update(app)
	if errInUpdateAppByID != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_PUBLISH_APP_TO_MARKETPLACE, "update App error: "+errInUpdateAppByID.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return
}

func (controller *Controller) GetAllAppListByIDInternal(c *gin.Context) {

}

func (controller *Controller) GetAppInternal(c *gin.Context) {

}
