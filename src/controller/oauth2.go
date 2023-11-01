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
	errorOAuth2Callback, _ := controller.TestFirstStringParamValueFromURI(c, PARAM_ERROR)

	fmt.Printf("[DUMP] GoogleOAuth2Exchange().state: %+v\n", state)
	fmt.Printf("[DUMP] GoogleOAuth2Exchange().errInGetState: %+v\n", errInGetState)
	fmt.Printf("[DUMP] GoogleOAuth2Exchange().code: %+v\n", code)
	fmt.Printf("[DUMP] GoogleOAuth2Exchange().errInGetCode: %+v\n", errInGetCode)
	fmt.Printf("[DUMP] GoogleOAuth2Exchange().errorOAuth2Callback: %+v\n", errorOAuth2Callback)

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
	if errorOAuth2Callback != "" || errInGetCode != nil || code == "" || errInExtract != nil {
		fmt.Printf("[FAILED] 1\n")
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		fmt.Printf("[FAILED] 2\n")

		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// check resource type for create OAuth token
	if !resource.CanCreateOAuthToken() {
		fmt.Printf("[FAILED] 3\n")

		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// new resource option
	resourceOptionGoogleSheets, errInNewGoogleSheetResourceOption := model.NewResourceOptionGoogleSheetsByResource(resource)
	if errInNewGoogleSheetResourceOption != nil {
		fmt.Printf("[FAILED] 4\n")
		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// validate resource option
	if !resourceOptionGoogleSheets.IsAvaliableAuthenticationMethod() {
		fmt.Printf("[FAILED] 5\n")

		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	// exchange access token
	exchangeTokenResponse, errInExchangeOAuthToken := oauthgoogle.ExchangeOAuthToken(code)
	fmt.Printf("[DUMP] GoogleOAuth2Exchange.exchangeTokenResponse: %+v\n", exchangeTokenResponse)
	if errInExchangeOAuthToken != nil {
		fmt.Printf("[FAILED] 6\n")

		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}
	resourceOptionGoogleSheets.UpdateByExchangeTokenResponse(exchangeTokenResponse)
	resource.UpdateGoogleSheetOAuth2Options(userID, resourceOptionGoogleSheets)

	// update resource
	errInUpdateResource := controller.Storage.ResourceStorage.UpdateWholeResource(resource)
	if errInUpdateResource != nil {
		fmt.Printf("[FAILED] 7\n")

		controller.FeedbackRedirect(c, redirectURIForFailed)
		return
	}

	fmt.Printf("[OK] 8\n")

	// redirect
	controller.FeedbackRedirect(c, redirectURIForSuccess)
	return
}
