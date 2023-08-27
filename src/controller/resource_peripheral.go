package controller

import (
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/mitchellh/mapstructure"
)

type OAuth2Opts struct {
	AccessType   string
	AccessToken  string
	TokenType    string
	RefreshToken string
	Status       int
}

func (controller *Controller) CreateOAuthToken(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_RESOURCE)
	controller.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_EDIT_RESOURCE)
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
	controller.FeedbackOK(c, NewCreateOAuthTokenResponse(token))
	return
}

type GoogleSheetsOAuth2Request struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

func (controller *Controller) GoogleSheetsOAuth2(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	_, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_RESOURCE)
	controller.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	var gsOAuth2Request GoogleSheetsOAuth2Request
	gsOAuth2Request.AccessToken = c.Query("accessToken")
	// validate request body fields
	validate := validator.New()
	if err := validate.Struct(gsOAuth2Request); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// validate the resource id
	res, err := controller.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}
	if res.Type != "googlesheets" {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "unsupported resource type")
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "get resource error: "+err.Error())
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "unsupported authentication type")
		return
	}

	// validate access token
	access, err := validateGSOAuth2Token(gsOAuth2Request.AccessToken)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "validate token error: "+err.Error())
		return
	}

	// return new url
	googleOAuthClientID := os.Getenv("ILLA_GS_CLIENT_ID")
	redirectURI := os.Getenv("ILLA_GS_REDIRECT_URI")
	u := url.URL{}
	if access == 1 {
		u = url.URL{
			Scheme:   "https",
			Host:     "accounts.google.com",
			Path:     "o/oauth2/v2/auth/oauthchooseaccount",
			RawQuery: "response_type=" + "code" + "&client_id=" + googleOAuthClientID + "&redirect_uri=" + redirectURI + "&state=" + gsOAuth2Request.AccessToken + "&scope=" + "https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/spreadsheets" + "&access_type=" + "offline" + "&prompt=" + "consent" + "&service=" + "lso" + "&o2v=" + "2" + "&flowName=" + "GeneralOAuthFlow",
		}
	} else {
		u = url.URL{
			Scheme:   "https",
			Host:     "accounts.google.com",
			Path:     "o/oauth2/v2/auth/oauthchooseaccount",
			RawQuery: "response_type=" + "code" + "&client_id=" + googleOAuthClientID + "&redirect_uri=" + redirectURI + "&state=" + gsOAuth2Request.AccessToken + "&scope=" + "https://www.googleapis.com/auth/spreadsheets.readonly https://www.googleapis.com/auth/drive.readonly" + "&access_type=" + "offline" + "&prompt=" + "consent" + "&service=" + "lso" + "&o2v=" + "2" + "&flowName=" + "GeneralOAuthFlow",
		}
	}
	c.JSON(200, gin.H{
		"url": u.String(),
	})
	return
}

func (controller *Controller) RefreshGSOAuth(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	controller.AttributeGroup.Init()
	controller.AttributeGroup.SetTeamID(teamID)
	controller.AttributeGroup.SetUserAuthToken(userAuthToken)
	controller.AttributeGroup.SetUnitType(accesscontrol.UNIT_TYPE_RESOURCE)
	controller.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(accesscontrol.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// validate the resource id
	res, err := controller.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "get resources error: "+err.Error())
		return
	}
	if res.Type != "googlesheets" {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "unsupported resource type")
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "get resource error: "+err.Error())
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "unsupported authentication type")
		return
	}

	// get new access token
	client := resty.New()
	resp, err := client.R().
		SetFormData(map[string]string{
			"client_id":     os.Getenv("ILLA_GS_CLIENT_ID"),
			"client_secret": os.Getenv("ILLA_GS_CLIENT_SECRET"),
			"refresh_token": googleSheetsResource.Opts.RefreshToken,
			"grant_type":    "refresh_token",
		}).
		Post("https://oauth2.googleapis.com/token")
	if resp.IsSuccess() {
		type RefreshTokenSuccessResponse struct {
			AccessToken string `json:"access_token"`
			Expiry      int    `json:"expires_in"`
			Scope       string `json:"scope"`
			TokenType   string `json:"token_type"`
		}
		var refreshTokenSuccessResponse RefreshTokenSuccessResponse
		if err := json.Unmarshal(resp.Body(), &refreshTokenSuccessResponse); err != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "fresh google sheets error: "+err.Error())
			return
		}
		googleSheetsResource.Opts.AccessToken = refreshTokenSuccessResponse.AccessToken
	} else if resp.IsError() {
		googleSheetsResource.Opts.RefreshToken = ""
		googleSheetsResource.Opts.Status = 1
		googleSheetsResource.Opts.AccessToken = ""
		googleSheetsResource.Opts.TokenType = ""
	}

	// update resource and return response
	updateRes, err := controller.resourceService.UpdateResource(resource.ResourceDto{
		ID:   idconvertor.ConvertStringToInt(res.ID),
		Name: res.Name,
		Type: res.Type,
		Options: map[string]interface{}{
			"authentication": googleSheetsResource.Authentication,
			"opts": map[string]interface{}{
				"accessType":   googleSheetsResource.Opts.AccessType,
				"accessToken":  googleSheetsResource.Opts.AccessToken,
				"tokenType":    googleSheetsResource.Opts.TokenType,
				"refreshToken": googleSheetsResource.Opts.RefreshToken,
				"status":       googleSheetsResource.Opts.Status,
			},
		},
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: userID,
	})
	res.Options = updateRes.Options
	controller.FeedbackOK(c, res)
	return
}
