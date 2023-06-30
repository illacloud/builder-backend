// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resthandler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/mitchellh/mapstructure"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ResourceRestHandler interface {
	FindAllResources(c *gin.Context)
	CreateResource(c *gin.Context)
	GetResource(c *gin.Context)
	UpdateResource(c *gin.Context)
	DeleteResource(c *gin.Context)
	TestConnection(c *gin.Context)
	GetMetaInfo(c *gin.Context)
	CreateOAuthToken(c *gin.Context)
	GoogleSheetsOAuth2(c *gin.Context)
	RefreshGSOAuth(c *gin.Context)
}

type ResourceRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService resource.ResourceService
	AttributeGroup  *ac.AttributeGroup
}

func NewResourceRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService, attrg *ac.AttributeGroup) *ResourceRestHandlerImpl {
	return &ResourceRestHandlerImpl{
		logger:          logger,
		resourceService: resourceService,
		AttributeGroup:  attrg,
	}
}

func (impl ResourceRestHandlerImpl) FindAllResources(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	res, err := impl.resourceService.FindAllResources(teamID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, res)
	return
}

func (impl ResourceRestHandlerImpl) CreateResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_CREATE_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	var rsc resource.ResourceDto
	rsc.InitUID()
	if err := json.NewDecoder(c.Request.Body).Decode(&rsc); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}
	if err := impl.resourceService.ValidateResourceOptions(rsc.Type, rsc.Options); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	rsc.SetTeamID(teamID)
	rsc.CreatedAt = time.Now().UTC()
	rsc.CreatedBy = userID
	rsc.UpdatedAt = time.Now().UTC()
	rsc.UpdatedBy = userID
	res, err := impl.resourceService.CreateResource(rsc)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_RESOURCE, "create resources error: "+err.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_CREATE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   idconvertor.ConvertStringToInt(res.ID),
		ResourceName: res.Name,
		ResourceType: res.Type,
	})

	// feedback
	FeedbackOK(c, res)
	return
}

func (impl ResourceRestHandlerImpl) GetResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, res)
	return
}

func (impl ResourceRestHandlerImpl) UpdateResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	var rscForExport resource.ResourceDtoForExport
	if err := json.NewDecoder(c.Request.Body).Decode(&rscForExport); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}
	rsc := rscForExport.ExportResourceDto()
	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}
	if err := impl.resourceService.ValidateResourceOptions(rsc.Type, rsc.Options); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// update
	rsc.ID = resourceID
	rsc.UpdatedBy = userID
	rsc.UpdatedAt = time.Now().UTC()
	res, err := impl.resourceService.UpdateResource(rsc)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE, "update resources error: "+err.Error())
		return
	}
	originInfo, _ := impl.resourceService.GetResource(teamID, rsc.ID)
	res.CreatedAt = originInfo.CreatedAt
	res.CreatedBy = originInfo.CreatedBy

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_UPDATE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   rsc.ID,
		ResourceName: res.Name,
		ResourceType: res.Type,
	})

	// feedback
	FeedbackOK(c, res)
	return
}

func (impl ResourceRestHandlerImpl) DeleteResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canDelete, errInCheckAttr := impl.AttributeGroup.CanDelete(ac.ACTION_DELETE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canDelete {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_DELETE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   resourceID,
		ResourceName: res.Name,
		ResourceType: res.Type,
	})

	if err := impl.resourceService.DeleteResource(teamID, resourceID); err != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_DELETE_RESOURCE, "delete resources error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, repository.NewDeleteResourceResponse(resourceID))
	return

}

func (impl ResourceRestHandlerImpl) TestConnection(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// format data to DTO struct
	var rscForExport resource.ResourceDtoForExport
	if err := json.NewDecoder(c.Request.Body).Decode(&rscForExport); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}
	rsc := rscForExport.ExportResourceDto()

	// validate `resource` valid required fields
	validate := validator.New()
	if err := validate.Struct(rsc); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	connRes, err := impl.resourceService.TestConnection(rsc)
	if err != nil || !connRes {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION, "test connection failed: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, nil)
	return
}

func (impl ResourceRestHandlerImpl) GetMetaInfo(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	res, err := impl.resourceService.GetMetaInfo(teamID, resourceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

type CreateOAuthTokenRequest struct {
	RedirectURL string `json:"redirectURL" validate:"required"`
	AccessType  string `json:"accessType" validate:"oneof=rw r"`
}

type GoogleSheetsResource struct {
	Authentication string
	Opts           OAuth2Opts
}
type OAuth2Opts struct {
	AccessType   string
	AccessToken  string
	TokenType    string
	RefreshToken string
	Status       int
}

func (impl ResourceRestHandlerImpl) CreateOAuthToken(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	var createOAuthTokenRequest CreateOAuthTokenRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&createOAuthTokenRequest); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}
	// validate request body fields
	validate := validator.New()
	if err := validate.Struct(createOAuthTokenRequest); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// validate the resource id
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}
	if res.Type != "googlesheets" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "unsupported resource type")
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "get resource error: "+err.Error())
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "unsupported authentication type")
		return
	}

	// generate access token
	access := 0
	if createOAuthTokenRequest.AccessType == "rw" {
		access = 1
	} else {
		access = 2
	}
	token, err := generateGSOAuth2Token(teamID, userID, resourceID, access, createOAuthTokenRequest.RedirectURL)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_TOKEN, "generate token error: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
	})
	return
}

type GoogleSheetsOAuth2Request struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

func (impl ResourceRestHandlerImpl) GoogleSheetsOAuth2(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	_, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	var gsOAuth2Request GoogleSheetsOAuth2Request
	gsOAuth2Request.AccessToken = c.Query("accessToken")
	// validate request body fields
	validate := validator.New()
	if err := validate.Struct(gsOAuth2Request); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// validate the resource id
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+err.Error())
		return
	}
	if res.Type != "googlesheets" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "unsupported resource type")
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "get resource error: "+err.Error())
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "unsupported authentication type")
		return
	}

	// validate access token
	access, err := validateGSOAuth2Token(gsOAuth2Request.AccessToken)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_AUTHORIZE_GS, "validate token error: "+err.Error())
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

func (impl ResourceRestHandlerImpl) RefreshGSOAuth(c *gin.Context) {
	// fetch needed params
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_RESOURCE)
	impl.AttributeGroup.SetUnitID(resourceID)
	canManage, errInCheckAttr := impl.AttributeGroup.CanManage(ac.ACTION_MANAGE_EDIT_RESOURCE)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// validate the resource id
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "get resources error: "+err.Error())
		return
	}
	if res.Type != "googlesheets" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "unsupported resource type")
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "get resource error: "+err.Error())
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "unsupported authentication type")
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
			FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_REFRESH_GS, "fresh google sheets error: "+err.Error())
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
	updateRes, err := impl.resourceService.UpdateResource(resource.ResourceDto{
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
	FeedbackOK(c, res)
	return
}
