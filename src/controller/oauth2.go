package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

func (controller *Controller) GoogleOAuth2Exchange(c *gin.Context) {
	state, errInGetState := controller.TestFirstStringParamValueFromURI(c, PARAM_STATE)
	code, errInGetCode := controller.TestFirstStringParamValueFromURI(c, PARAM_CODE)
	errorOAuth2Callback, errInGetError := controller.TestFirstStringParamValueFromURI(c, PARAM_ERROR)

	// check input
	if errInGetState != nil {
		controller.FeedbackBadRequest(c, nil)
		return
	}

	// new OAuth claims
	googleSheetsOAuth2Claims := model.NewGoogleSheetsOAuth2Claims()
	teamID, userID, resourceID, url, errInExtract := googleSheetsOAuth2Claims.ExtractGoogleSheetsOAuth2TokenInfo(state)
	if errInExtract != nil {
		c.Redirect(302)
		return
	}
	redirectURIForFailed := fmt.Sprintf("%s?status=%d&resourceID=%s", url, model.GOOGLE_SHEETS_OAUTH_STATUS_FAILED, idconvertor.ConvertIntToString(resourceID))
	redirectURIForSuccess := fmt.Sprintf("%s?status=%d&resourceID=%s", url, model.GOOGLE_SHEETS_OAUTH_STATUS_SUCCESS, idconvertor.ConvertIntToString(resourceID))
	if errorOAuth2Callback != "" || code == "" {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// check resource type for create OAuth token
	if !resource.CanCreateOAuthToken() {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// new resource option
	resourceOptionGoogleSheets, errInNewGoogleSheetResourceOption := model.NewResourceOptionGoogleSheetsByResource(resource)
	if errInNewGoogleSheetResourceOption != nil {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// validate resource option
	if !resourceOptionGoogleSheets.IsAvaliableAuthenticationMethod() {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// exchange access token
	refreshTokenResponse, errInRefreshOAuthToken := oauthgoogle.ExchangeOAuthToken(code)
	if errInRefreshOAuthToken != nil {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}
	resourceOptionGoogleSheets.SetAccessToken(refreshTokenResponse.ExportAccessToken())
	resource.UpdateGoogleSheetOAuth2Options(userID, resourceOptionGoogleSheets)

	// update resource
	errInUpdateResource := controller.Storage.ResourceStorage.UpdateWholeResource(resource)
	if errInUpdateResource != nil {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// redirect
	controller.FeedbackRedirect(c, redirectURIForSuccess)
	return
}
