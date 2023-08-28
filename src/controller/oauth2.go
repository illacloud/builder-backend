package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/oauthgoogle"
)

func (controller *Controller) GoogleOAuth2Exchange(c *gin.Context) {
	state, errInGetState := controller.TestFirstStringParamValueFromURI(c, PARAM_STATE)
	code, errInGetCode := controller.TestFirstStringParamValueFromURI(c, PARAM_CODE)
	errorOAuth2Callback, errInGetError := controller.TestFirstStringParamValueFromURI(c, PARAM_ERROR)

	// check input
	if errInGetState != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_URI_FAILED, "")
		return
	}

	// new OAuth claims
	googleSheetsOAuth2Claims := model.NewGoogleSheetsOAuth2Claims()
	teamID, userID, resourceID, url, errInExtract := googleSheetsOAuth2Claims.ExtractGoogleSheetsOAuth2TokenInfo(state)
	redirectURIForFailed := fmt.Sprintf("%s?status=%d&resourceID=%s", url, model.GOOGLE_SHEETS_OAUTH_STATUS_FAILED, idconvertor.ConvertIntToString(resourceID))
	redirectURIForSuccess := fmt.Sprintf("%s?status=%d&resourceID=%s", url, model.GOOGLE_SHEETS_OAUTH_STATUS_SUCCESS, idconvertor.ConvertIntToString(resourceID))
	if errorOAuth2Callback != "" || errInGetCode != nil || errInGetError != nil || code == "" || errInExtract != nil {
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
	exchangeTokenResponse, errInExchangeOAuthToken := oauthgoogle.ExchangeOAuthToken(code)
	if errInExchangeOAuthToken != nil {
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}
	resourceOptionGoogleSheets.SetAccessToken(exchangeTokenResponse.ExportAccessToken())
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
