package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/oauthgoogle"
)

func (controller *Controller) CreateGoogleOAuthToken(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_MANAGE_EDIT_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	createOAuthTokenRequest := request.NewCreateOAuthTokenRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&createOAuthTokenRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate request body fields
	validate := validator.New()
	if err := validate.Struct(createOAuthTokenRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// check resource type for create OAuth token
	if !resource.CanCreateOAuthToken() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "unsupported resource type")
		return
	}

	// new resource option
	resourceOptionGoogleSheets, errInNewGoogleSheetResourceOption := model.NewResourceOptionGoogleSheetsByResource(resource)
	if errInNewGoogleSheetResourceOption != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "unsupported resource type: "+errInNewGoogleSheetResourceOption.Error())
		return
	}

	// validate resource option
	if !resourceOptionGoogleSheets.IsAvaliableAuthenticationMethod() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "unsupported authentication type")
		return
	}

	// generate access token
	token, err := model.GenerateGoogleSheetsOAuth2Token(teamID, userID, resourceID, createOAuthTokenRequest)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "generate token error: "+err.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewCreateOAuthTokenResponse(token))
	return
}

func (controller *Controller) GetGoogleSheetsOAuth2Token(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	accessToken, errInGetAccessToken := controller.GetFirstStringParamValueFromURI(c, PARAM_ACCESS_TOKEN)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAccessToken != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_MANAGE_EDIT_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// check resource type for create OAuth token
	if !resource.CanCreateOAuthToken() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_TOKEN, "unsupported resource type")
		return
	}

	// new resource option
	resourceOptionGoogleSheets, errInNewGoogleSheetResourceOption := model.NewResourceOptionGoogleSheetsByResource(resource)
	if errInNewGoogleSheetResourceOption != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_TOKEN, "unsupported resource type: "+errInNewGoogleSheetResourceOption.Error())
		return
	}

	// validate resource option
	if !resourceOptionGoogleSheets.IsAvaliableAuthenticationMethod() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_TOKEN, "unsupported authentication type")
		return
	}

	// validate access token
	googleSheetsOAuth2Claims := model.NewGoogleSheetsOAuth2Claims()
	accessType, errinValidateAccessToken := googleSheetsOAuth2Claims.ValidateAccessToken(accessToken)
	if errinValidateAccessToken != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GOOGLE_SHEETS, "validate token error: "+errinValidateAccessToken.Error())
		return
	}

	// return new url
	controller.FeedbackOK(c, response.NewGoogleSheetsOAuth2Response(accessType, accessToken))
	return
}

func (controller *Controller) RefreshGoogleSheetsOAuth(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_MANAGE_EDIT_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}
	fmt.Printf("[DUMP] RefreshGoogleSheetsOAuth.resource: %+v\n", resource)

	// check resource type for create OAuth token
	if !resource.CanCreateOAuthToken() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_TOKEN, "unsupported resource type")
		return
	}

	// new resource option
	resourceOptionGoogleSheets, errInNewGoogleSheetResourceOption := model.NewResourceOptionGoogleSheetsByResource(resource)
	if errInNewGoogleSheetResourceOption != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_TOKEN, "unsupported resource type: "+errInNewGoogleSheetResourceOption.Error())
		return
	}

	// validate resource option
	if !resourceOptionGoogleSheets.IsAvaliableAuthenticationMethod() {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_TOKEN, "unsupported authentication type")
		return
	}

	fmt.Printf("[DUMP] RefreshGoogleSheetsOAuth.resourceOptionGoogleSheets: %+v\n", resourceOptionGoogleSheets)

	// refresh access token
	refreshTokenResponse, errInRefreshOAuthToken := oauthgoogle.RefreshOAuthToken(resourceOptionGoogleSheets.ExportRefreshToken())
	if errInRefreshOAuthToken != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GOOGLE_SHEETS, "fresh google sheets oauth token error: "+errInRefreshOAuthToken.Error())
		return
	}
	fmt.Printf("[DUMP] RefreshGoogleSheetsOAuth.refreshTokenResponse: %+v\n", refreshTokenResponse)

	resourceOptionGoogleSheets.SetAccessToken(refreshTokenResponse.ExportAccessToken())
	resource.UpdateGoogleSheetOAuth2Options(userID, resourceOptionGoogleSheets)

	// update resource
	errInUpdateResource := controller.Storage.ResourceStorage.UpdateWholeResource(resource)
	if errInUpdateResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE, "update resources error: "+errInUpdateResource.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewUpdateResourceResponse(resource))
	return
}
