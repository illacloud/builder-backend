package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/illa-marketplace-backend/src/response"
)

const (
	PARAM_AUTHORIZATION    = "Authorization"
	PARAM_REQUEST_TOKEN    = "Request-Token"
	PARAM_TEAM_ID          = "teamID"
	PARAM_USER_ID          = "userID"
	PARAM_TARGET_USER_ID   = "targetUserID"
	PARAM_TEAM_IDENTIFIER  = "teamIdentifier"
	PARAM_USER_ROLE        = "userRole"
	PARAM_INVITE_LINK_HASH = "inviteLinkHash"
	PARAM_UNIT_TYPE        = "unitType"
	PARAM_UNIT_ID          = "unitID"
	PARAM_ATTRIBUTE_ID     = "attributeID"
	PARAM_FROM_ID          = "fromID"
	PARAM_TO_ID            = "toID"
	PARAM_ACTION_ID        = "actionID"
	PARAM_APP_ID           = "appID"
	PARAM_VERSION          = "version"
	PARAM_RESOURCE_ID      = "resourceID"
	PARAM_PAGE_LIMIT       = "pageLimit"
	PARAM_PAGE             = "page"
	PARAM_SNAPSHOT_ID      = "snapshotID"
	PARAM_STATE            = "state"
	PARAM_CODE             = "code"
	PARAM_ERROR            = "error"
)

const (
	// validate failed
	ERROR_FLAG_VALIDATE_ACCOUNT_FAILED                  = "ERROR_FLAG_VALIDATE_ACCOUNT_FAILED"
	ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED             = "ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED"
	ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED            = "ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED"
	ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED            = "ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED"
	ERROR_FLAG_VALIDATE_VERIFICATION_CODE_FAILED        = "ERROR_FLAG_VALIDATE_VERIFICATION_CODE_FAILED"
	ERROR_FLAG_VALIDATE_RESOURCE_FAILED                 = "ERROR_FLAG_VALIDATE_RESOURCE_FAILED"
	ERROR_FLAG_PARSE_REQUEST_BODY_FAILED                = "ERROR_FLAG_PARSE_REQUEST_BODY_FAILED"
	ERROR_FLAG_PARSE_REQUEST_URI_FAILED                 = "ERROR_FLAG_PARSE_REQUEST_URI_FAILED"
	ERROR_FLAG_PARSE_INVITE_LINK_HASH_FAILED            = "ERROR_FLAG_PARSE_INVITE_LINK_HASH_FAILED"
	ERROR_FLAG_CAN_NOT_TRANSFER_OWNER_TO_PENDING_USER   = "ERROR_FLAG_CAN_NOT_TRANSFER_OWNER_TO_PENDING_USER"
	ERROR_FLAG_CAN_NOT_REMOVE_OWNER_FROM_TEAM           = "ERROR_FLAG_CAN_NOT_REMOVE_OWNER_FROM_TEAM"
	ERROR_FLAG_SIGN_UP_EMAIL_MISMATCH                   = "ERROR_FLAG_SIGN_UP_EMAIL_MISMATCH"
	ERROR_FLAG_OWNER_ROLE_MUST_BE_TRANSFERED            = "ERROR_FLAG_OWNER_ROLE_MUST_BE_TRANSFERED"
	ERROR_FLAG_PASSWORD_INVALIED                        = "ERROR_FLAG_PASSWORD_INVALIED"
	ERROR_FLAG_TEAM_MUST_TRANSFERED_BEFORE_USER_SUSPEND = "ERROR_FLAG_TEAM_MUST_TRANSFERED_BEFORE_USER_SUSPEND"

	// can note create
	ERROR_FLAG_CAN_NOT_CREATE_USER            = "ERROR_FLAG_CAN_NOT_CREATE_USER"
	ERROR_FLAG_CAN_NOT_CREATE_TEAM            = "ERROR_FLAG_CAN_NOT_CREATE_TEAM"
	ERROR_FLAG_CAN_NOT_CREATE_TEAM_MEMBER     = "ERROR_FLAG_CAN_NOT_CREATE_TEAM_MEMBER"
	ERROR_FLAG_CAN_NOT_CREATE_INVITE          = "ERROR_FLAG_CAN_NOT_CREATE_INVITE"
	ERROR_FLAG_CAN_NOT_CREATE_INVITATION_CODE = "ERROR_FLAG_CAN_NOT_CREATE_INVITATION_CODE"
	ERROR_FLAG_CAN_NOT_CREATE_DOMAIN          = "ERROR_FLAG_CAN_NOT_CREATE_DOMAIN"
	ERROR_FLAG_CAN_NOT_CREATE_ACTION          = "ERROR_FLAG_CAN_NOT_CREATE_ACTION"
	ERROR_FLAG_CAN_NOT_CREATE_RESOURCE        = "ERROR_FLAG_CAN_NOT_CREATE_RESOURCE"
	ERROR_FLAG_CAN_NOT_CREATE_APP             = "ERROR_FLAG_CAN_NOT_CREATE_APP"
	ERROR_FLAG_CAN_NOT_CREATE_STATE           = "ERROR_FLAG_CAN_NOT_CREATE_STATE"
	ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT        = "ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT"
	ERROR_FLAG_CAN_NOT_CREATE_COMPONENT_TREE  = "ERROR_FLAG_CAN_NOT_CREATE_COMPONENT_TREE"

	// can not get resource
	ERROR_FLAG_CAN_NOT_GET_USER                = "ERROR_FLAG_CAN_NOT_GET_USER"
	ERROR_FLAG_CAN_NOT_GET_TEAM                = "ERROR_FLAG_CAN_NOT_GET_TEAM"
	ERROR_FLAG_CAN_NOT_GET_TEAM_MEMBER         = "ERROR_FLAG_CAN_NOT_GET_TEAM_MEMBER"
	ERROR_FLAG_CAN_NOT_GET_INVITE              = "ERROR_FLAG_CAN_NOT_GET_INVITE"
	ERROR_FLAG_CAN_NOT_GET_INVITATION_CODE     = "ERROR_FLAG_CAN_NOT_GET_INVITATION_CODE"
	ERROR_FLAG_CAN_NOT_GET_DOMAIN              = "ERROR_FLAG_CAN_NOT_GET_DOMAIN"
	ERROR_FLAG_CAN_NOT_GET_ACTION              = "ERROR_FLAG_CAN_NOT_GET_ACTION"
	ERROR_FLAG_CAN_NOT_GET_RESOURCE            = "ERROR_FLAG_CAN_NOT_GET_RESOURCE"
	ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO  = "ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO"
	ERROR_FLAG_CAN_NOT_GET_APP                 = "ERROR_FLAG_CAN_NOT_GET_APP"
	ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION = "ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION"
	ERROR_FLAG_CAN_NOT_GET_STATE               = "ERROR_FLAG_CAN_NOT_GET_STATE"
	ERROR_FLAG_CAN_NOT_GET_SNAPSHOT            = "ERROR_FLAG_CAN_NOT_GET_SNAPSHOT"

	// can not update resource
	ERROR_FLAG_CAN_NOT_UPDATE_USER            = "ERROR_FLAG_CAN_NOT_UPDATE_USER"
	ERROR_FLAG_CAN_NOT_UPDATE_TEAM            = "ERROR_FLAG_CAN_NOT_UPDATE_TEAM"
	ERROR_FLAG_CAN_NOT_UPDATE_TEAM_MEMBER     = "ERROR_FLAG_CAN_NOT_UPDATE_TEAM_MEMBER"
	ERROR_FLAG_CAN_NOT_UPDATE_INVITE          = "ERROR_FLAG_CAN_NOT_UPDATE_INVITE"
	ERROR_FLAG_CAN_NOT_UPDATE_INVITATION_CODE = "ERROR_FLAG_CAN_NOT_UPDATE_INVITATION_CODE"
	ERROR_FLAG_CAN_NOT_UPDATE_DOMAIN          = "ERROR_FLAG_CAN_NOT_UPDATE_DOMAIN"
	ERROR_FLAG_CAN_NOT_UPDATE_ACTION          = "ERROR_FLAG_CAN_NOT_UPDATE_ACTION"
	ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE        = "ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE"
	ERROR_FLAG_CAN_NOT_UPDATE_APP             = "ERROR_FLAG_CAN_NOT_UPDATE_APP"
	ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE      = "ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE"
	ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT        = "ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT"

	// can not delete
	ERROR_FLAG_CAN_NOT_DELETE_USER            = "ERROR_FLAG_CAN_NOT_DELETE_USER"
	ERROR_FLAG_CAN_NOT_DELETE_TEAM            = "ERROR_FLAG_CAN_NOT_DELETE_TEAM"
	ERROR_FLAG_CAN_NOT_DELETE_TEAM_MEMBER     = "ERROR_FLAG_CAN_NOT_DELETE_TEAM_MEMBER"
	ERROR_FLAG_CAN_NOT_DELETE_INVITE          = "ERROR_FLAG_CAN_NOT_DELETE_INVITE"
	ERROR_FLAG_CAN_NOT_DELETE_INVITATION_CODE = "ERROR_FLAG_CAN_NOT_DELETE_INVITATION_CODE"
	ERROR_FLAG_CAN_NOT_DELETE_DOMAIN          = "ERROR_FLAG_CAN_NOT_DELETE_DOMAIN"
	ERROR_FLAG_CAN_NOT_DELETE_ACTION          = "ERROR_FLAG_CAN_NOT_DELETE_ACTION"
	ERROR_FLAG_CAN_NOT_DELETE_RESOURCE        = "ERROR_FLAG_CAN_NOT_DELETE_RESOURCE"
	ERROR_FLAG_CAN_NOT_DELETE_APP             = "ERROR_FLAG_CAN_NOT_DELETE_APP"

	// can not other operation
	ERROR_FLAG_CAN_NOT_CHECK_TEAM_MEMBER        = "ERROR_FLAG_CAN_NOT_CHECK_TEAM_MEMBER"
	ERROR_FLAG_CAN_NOT_DUPLICATE_APP            = "ERROR_FLAG_CAN_NOT_DUPLICATE_APP"
	ERROR_FLAG_CAN_NOT_RELEASE_APP              = "ERROR_FLAG_CAN_NOT_RELEASE_APP"
	ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION = "ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION"

	// permission failed
	ERROR_FLAG_ACCESS_DENIED                  = "ERROR_FLAG_ACCESS_DENIED"
	ERROR_FLAG_TEAM_CLOSED_THE_PERMISSION     = "ERROR_FLAG_TEAM_CLOSED_THE_PERMISSION"
	ERROR_FLAG_EMAIL_ALREADY_USED             = "ERROR_FLAG_EMAIL_ALREADY_USED"
	ERROR_FLAG_EMAIL_HAS_BEEN_TAKEN           = "ERROR_FLAG_EMAIL_HAS_BEEN_TAKEN"
	ERROR_FLAG_INVITATION_CODE_ALREADY_USED   = "ERROR_FLAG_INVITATION_CODE_ALREADY_USED"
	ERROR_FLAG_INVITATION_LINK_UNAVALIABLE    = "ERROR_FLAG_INVITATION_LINK_UNAVALIABLE"
	ERROR_FLAG_TEAM_IDENTIFIER_HAS_BEEN_TAKEN = "ERROR_FLAG_TEAM_IDENTIFIER_HAS_BEEN_TAKEN"
	ERROR_FLAG_USER_ALREADY_JOINED_TEAM       = "ERROR_FLAG_USER_ALREADY_JOINED_TEAM"
	ERROR_FLAG_SIGN_IN_FAILED                 = "ERROR_FLAG_SIGN_IN_FAILED"
	ERROR_FLAG_NO_SUCH_USER                   = "ERROR_FLAG_NO_SUCH_USER"

	// call resource failed
	ERROR_FLAG_SEND_EMAIL_FAILED             = "ERROR_FLAG_SEND_EMAIL_FAILED"
	ERROR_FLAG_SEND_VERIFICATION_CODE_FAILED = "ERROR_FLAG_SEND_VERIFICATION_CODE_FAILED"
	ERROR_FLAG_CREATE_LINK_FAILED            = "ERROR_FLAG_CREATE_LINK_FAILED"
	ERROR_FLAG_CREATE_UPLOAD_URL_FAILED      = "ERROR_FLAG_CREATE_UPLOAD_URL_FAILED"
	ERROR_FLAG_EXECUTE_ACTION_FAILED         = "ERROR_FLAG_EXECUTE_ACTION_FAILED"
	ERROR_FLAG_GENERATE_SQL_FAILED           = "ERROR_FLAG_GENERATE_SQL_FAILED"

	// internal failed
	ERROR_FLAG_BUILD_TEAM_MEMBER_LIST_FAILED = "ERROR_FLAG_BUILD_TEAM_MEMBER_LIST_FAILED"
	ERROR_FLAG_BUILD_TEAM_CONFIG_FAILED      = "ERROR_FLAG_BUILD_TEAM_CONFIG_FAILED"
	ERROR_FLAG_BUILD_TEAM_PERMISSION_FAILED  = "ERROR_FLAG_BUILD_TEAM_PERMISSION_FAILED"
	ERROR_FLAG_BUILD_USER_INFO_FAILED        = "ERROR_FLAG_BUILD_USER_INFO_FAILED"
	ERROR_FLAG_BUILD_APP_CONFIG_FAILED       = "ERROR_FLAG_BUILD_APP_CONFIG_FAILED"
	ERROR_FLAG_GENERATE_PASSWORD_FAILED      = "ERROR_FLAG_GENERATE_PASSWORD_FAILED"

	// google sheets oauth2 failed
	ERROR_FLAG_CAN_NOT_CREATE_TOKEN            = "ERROR_FLAG_CAN_NOT_CREATE_TOKEN"
	ERROR_FLAG_CAN_NOT_AUTHORIZE_GOOGLE_SHEETS = "ERROR_FLAG_CAN_NOT_AUTHORIZE_GOOGLE_SHEETS"
	ERROR_FLAG_CAN_NOT_REFRESH_GOOGLE_SHEETS   = "ERROR_FLAG_CAN_NOT_REFRESH_GOOGLE_SHEETS"
)

var SKIPPING_MAGIC_ID = map[string]int{
	"0":  0,
	"-1": -1,
	"-2": -2,
	"-3": -3,
}

func (controller *Controller) GetUserAuthTokenFromHeader(c *gin.Context) (string, error) {
	// fetch token
	rawToken := c.Request.Header[PARAM_AUTHORIZATION]
	if len(rawToken) != 1 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "HTTP request header missing request token.")
		return "", errors.New("missing request token.")
	}
	var token string
	token = rawToken[0]
	return token, nil
}

func (controller *Controller) ValidateRequestTokenFromHeader(c *gin.Context, input ...string) (bool, error) {
	// fetch token
	rawToken := c.Request.Header[PARAM_REQUEST_TOKEN]
	if len(rawToken) != 1 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "HTTP request header missing request token.")
		return false, errors.New("missing request token.")
	}
	var token string
	token = rawToken[0]
	// validate
	tokenShouldBe := controller.RequestTokenValidator.GenerateValidateTokenBySliceParam(input)
	if token != tokenShouldBe {
		log.Println("Illegal internal request token detected: \"" + token + "\", the token should be: \"" + tokenShouldBe + "\"")
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "request token mismatch.")
		return false, errors.New("request token mismatch.")
	}
	return true, nil
}

func (controller *Controller) ValidateRequestTokenFromHeaderByStringMap(c *gin.Context, input []string) (bool, error) {
	// fetch token
	rawToken := c.Request.Header[PARAM_REQUEST_TOKEN]
	if len(rawToken) != 1 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "HTTP request header missing request token.")
		return false, errors.New("missing request token.")
	}
	var token string
	token = rawToken[0]
	// validate
	tokenShouldBe := controller.RequestTokenValidator.GenerateValidateTokenBySliceParam(input)
	if token != tokenShouldBe {
		log.Println("Illegal internal request token detected: \"" + token + "\", the token should be: \"" + tokenShouldBe + "\"")
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "request token mismatch.")
		return false, errors.New("request token mismatch.")
	}
	return true, nil
}

func (controller *Controller) GetStringFromFormData(c *gin.Context, paramName string) (string, error) {
	// get request param
	paramValue := c.PostFormArray(paramName)

	// ho hit, convert
	if len(paramValue) == 0 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return "", errors.New("input missing " + paramName + " field.")
	}
	return paramValue[0], nil
}

func (controller *Controller) GetOptionalStringFromFormData(c *gin.Context, paramName string) string {
	// get request param
	paramValue := c.PostFormArray(paramName)

	// ho hit, convert
	if len(paramValue) == 0 {
		return ""
	}
	return paramValue[0]
}

func (controller *Controller) GetMagicIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	// check skipping id
	if intID, hitSkippingID := SKIPPING_MAGIC_ID[paramValue]; hitSkippingID {
		return intID, nil
	}
	// ho hit, convert
	if len(paramValue) == 0 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return 0, errors.New("input missing " + paramName + " field.")
	}
	paramValueInt := idconvertor.ConvertStringToInt(paramValue)
	return paramValueInt, nil
}

// test if Magic int exists in param, if not ,return 0 and an error.
func (controller *Controller) TestMagicIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		return 0, errors.New("input missing " + paramName + " field.")
	}
	paramValueInt := idconvertor.ConvertStringToInt(paramValue)
	return paramValueInt, nil
}

func (controller *Controller) GetIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return 0, errors.New("input missing " + paramName + " field.")
	}
	paramValueInt, okAssert := strconv.Atoi(paramValue)
	if okAssert != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param in int format.")
		return 0, errors.New("input teamID in wrong format.")
	}
	return paramValueInt, nil
}

func (controller *Controller) GetStringParamFromRequest(c *gin.Context, paramName string) (string, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return "", errors.New("input missing " + paramName + " field.")
	}
	return paramValue, nil
}

func (controller *Controller) TestStringParamFromRequest(c *gin.Context, paramName string) (string, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		return "", errors.New("input missing " + paramName + " field.")
	}
	return paramValue, nil
}

func (controller *Controller) TestFirstStringParamValueFromURI(c *gin.Context, paramName string) (string, error) {
	valueMaps := c.Request.URL.Query()
	paramValues, hit := valueMaps[paramName]
	// get request param
	if !hit {
		return "", errors.New("input missing " + paramName + " field.")
	}
	return paramValues[0], nil
}

func (controller *Controller) GetFirstStringParamValueFromURI(c *gin.Context, paramName string) (string, error) {
	valueMaps := c.Request.URL.Query()
	paramValues, hit := valueMaps[paramName]
	// get request param
	if !hit {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return "", errors.New("input missing " + paramName + " field.")
	}
	return paramValues[0], nil
}

func (controller *Controller) GetStringParamValuesFromURI(c *gin.Context, paramName string) ([]string, error) {
	valueMaps := c.Request.URL.Query()
	paramValues, hit := valueMaps[paramName]
	// get request param
	if !hit {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return nil, errors.New("input missing " + paramName + " field.")
	}
	return paramValues, nil
}

func (controller *Controller) GetStringParamFromHeader(c *gin.Context, paramName string) (string, error) {
	paramValue := c.Request.Header[paramName]
	var ret string
	if len(paramValue) != 1 {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "can not fetch param from header.")
		return "", errors.New("can not fetch param from header.")
	} else {
		ret = paramValue[0]
	}
	return ret, nil
}

// @note: this param was setted by authenticator.JWTAuth() method
func (controller *Controller) GetUserIDFromAuth(c *gin.Context) (int, error) {
	// get request param
	userID, ok := c.Get("userID")
	if !ok {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "auth token invalied, can not fetch user ID in it.")
		return 0, errors.New("input missing userID field.")
	}
	userIDInt, okAssert := userID.(int)
	if !okAssert {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "auth token invalied,user ID is not int type in it.")
		return 0, errors.New("input userID in wrong format.")
	}
	return userIDInt, nil
}

func (controller *Controller) FeedbackOK(c *gin.Context, resp response.Response) {
	if resp != nil {
		c.JSON(http.StatusOK, resp.ExportForFeedback())
		return
	}
	// HTTP 200 with empty response
	c.JSON(http.StatusOK, nil)
}

func (controller *Controller) FeedbackCreated(c *gin.Context, resp response.Response) {
	if resp != nil {
		c.JSON(http.StatusCreated, resp.ExportForFeedback())
		return
	}
	// HTTP 201 with empty response
	c.JSON(http.StatusCreated, nil)
}

func (controller *Controller) FeedbackBadRequest(c *gin.Context, errorFlag string, errorMessage string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errorCode":    400,
		"errorFlag":    errorFlag,
		"errorMessage": errorMessage,
	})
	return
}

func (controller *Controller) FeedbackRedirect(c *gin.Context, uri string) {
	c.Redirect(302, uri)
	return
}

func (controller *Controller) FeedbackInternalServerError(c *gin.Context, errorFlag string, errorMessage string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"errorCode":    500,
		"errorFlag":    errorFlag,
		"errorMessage": errorMessage,
	})
	return
}
