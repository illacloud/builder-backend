package resthandler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/internal/repository"
)

const PARAM_AUTHORIZATION = "Authorization"
const PARAM_TEAM_ID = "teamID"
const PARAM_USER_ID = "userID"
const PARAM_TARGET_USER_ID = "targetUserID"
const PARAM_USER_ROLE = "userRole"
const PARAM_INVITE_LINK_HASH = "inviteLinkHash"
const PARAM_UNIT_TYPE = "unitType"
const PARAM_UNIT_ID = "unitID"
const PARAM_ATTRIBUTE_ID = "attributeID"
const PARAM_FROM_ID = "fromID"
const PARAM_TO_ID = "toID"
const PARAM_ACTION_ID = "actionID"
const PARAM_APP_ID = "appID"
const PARAM_VERSION = "version"
const PARAM_RESOURCE_ID = "resourceID"

const (
	// validate failed
	ERROR_FLAG_VALIDATE_ACCOUNT_FAILED = iota + 1 // start with 1
	ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED
	ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED
	ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED
	ERROR_FLAG_VALIDATE_VERIFICATION_CODE_FAILED
	ERROR_FLAG_PARSE_REQUEST_BODY_FAILED
	ERROR_FLAG_PARSE_REQUEST_URI_FAILED
	ERROR_FLAG_PARSE_INVITE_LINK_HASH_FAILED
	ERROR_FLAG_CAN_NOT_TRANSFER_OWNER_TO_PENDING_USER
	ERROR_FLAG_CAN_NOT_REMOVE_OWNER_FROM_TEAM
	ERROR_FLAG_SIGN_UP_EMAIL_MISMATCH
	ERROR_FLAG_OWNER_ROLE_MUST_BE_TRANSFERED
	ERROR_FLAG_PASSWORD_INVALIED
	ERROR_FLAG_TEAM_MUST_TRANSFERED_BEFORE_USER_SUSPEND

	// can note create
	ERROR_FLAG_CAN_NOT_CREATE_USER
	ERROR_FLAG_CAN_NOT_CREATE_TEAM
	ERROR_FLAG_CAN_NOT_CREATE_TEAM_MEMBER
	ERROR_FLAG_CAN_NOT_CREATE_INVITE
	ERROR_FLAG_CAN_NOT_CREATE_INVITATION_CODE
	ERROR_FLAG_CAN_NOT_CREATE_DOMAIN
	ERROR_FLAG_CAN_NOT_CREATE_ACTION
	ERROR_FLAG_CAN_NOT_CREATE_RESOURCE
	ERROR_FLAG_CAN_NOT_CREATE_APP

	// can not get resource
	ERROR_FLAG_CAN_NOT_GET_USER
	ERROR_FLAG_CAN_NOT_GET_TEAM
	ERROR_FLAG_CAN_NOT_GET_TEAM_MEMBER
	ERROR_FLAG_CAN_NOT_GET_INVITE
	ERROR_FLAG_CAN_NOT_GET_INVITATION_CODE
	ERROR_FLAG_CAN_NOT_GET_DOMAIN
	ERROR_FLAG_CAN_NOT_GET_ACTION
	ERROR_FLAG_CAN_NOT_GET_RESOURCE
	ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO
	ERROR_FLAG_CAN_NOT_GET_APP
	ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION

	// can not update resource
	ERROR_FLAG_CAN_NOT_UPDATE_USER
	ERROR_FLAG_CAN_NOT_UPDATE_TEAM
	ERROR_FLAG_CAN_NOT_UPDATE_TEAM_MEMBER
	ERROR_FLAG_CAN_NOT_UPDATE_INVITE
	ERROR_FLAG_CAN_NOT_UPDATE_INVITATION_CODE
	ERROR_FLAG_CAN_NOT_UPDATE_DOMAIN
	ERROR_FLAG_CAN_NOT_UPDATE_ACTION
	ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE
	ERROR_FLAG_CAN_NOT_UPDATE_APP

	// can not delete
	ERROR_FLAG_CAN_NOT_DELETE_USER
	ERROR_FLAG_CAN_NOT_DELETE_TEAM
	ERROR_FLAG_CAN_NOT_DELETE_TEAM_MEMBER
	ERROR_FLAG_CAN_NOT_DELETE_INVITE
	ERROR_FLAG_CAN_NOT_DELETE_INVITATION_CODE
	ERROR_FLAG_CAN_NOT_DELETE_DOMAIN
	ERROR_FLAG_CAN_NOT_DELETE_ACTION
	ERROR_FLAG_CAN_NOT_DELETE_RESOURCE
	ERROR_FLAG_CAN_NOT_DELETE_APP

	// can not other operation
	ERROR_FLAG_CAN_NOT_CHECK_TEAM_MEMBER
	ERROR_FLAG_CAN_NOT_DUPLICATE_APP
	ERROR_FLAG_CAN_NOT_RELEASE_APP
	ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION

	// permission failed
	ERROR_FLAG_ACCESS_DENIED
	ERROR_FLAG_TEAM_CLOSED_THE_PERMISSION
	ERROR_FLAG_EMAIL_ALREADY_USED
	ERROR_FLAG_EMAIL_HAS_BEEN_TAKEN
	ERROR_FLAG_INVITATION_CODE_ALREADY_USED
	ERROR_FLAG_INVITATION_LINK_UNAVALIABLE
	ERROR_FLAG_TEAM_IDENTIFIER_HAS_BEEN_TAKEN
	ERROR_FLAG_USER_ALREADY_JOINED_TEAM
	ERROR_FLAG_SIGN_IN_FAILED
	ERROR_FLAG_NO_SUCH_USER

	// call resource failed
	ERROR_FLAG_SEND_EMAIL_FAILED
	ERROR_FLAG_SEND_VERIFICATION_CODE_FAILED
	ERROR_FLAG_CREATE_LINK_FAILED
	ERROR_FLAG_CREATE_UPLOAD_URL_FAILED
	ERROR_FLAG_EXECUTE_ACTION_FAILED
	ERROR_FLAG_GENERATE_SQL_FAILED

	// internal failed
	ERROR_FLAG_BUILD_TEAM_MEMBER_LIST_FAILED
	ERROR_FLAG_BUILD_TEAM_CONFIG_FAILED
	ERROR_FLAG_BUILD_TEAM_PERMISSION_FAILED
	ERROR_FLAG_BUILD_USER_INFO_FAILED
	ERROR_FLAG_BUILD_APP_CONFIG_FAILED
	ERROR_FLAG_GENERATE_PASSWORD_FAILED
)

func GetUserAuthTokenFromHeader(c *gin.Context) (string, error) {
	// fetch token
	rawToken := c.Request.Header[PARAM_AUTHORIZATION]
	if len(rawToken) != 1 {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "HTTP request header missing request token.")
		return "", errors.New("missing request token.")
	}
	var token string
	token = rawToken[0]
	return token, nil
}

func GetMagicIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return 0, errors.New("input mission " + paramName + " field.")
	}
	paramValueInt := idconvertor.ConvertStringToInt(paramValue)
	return paramValueInt, nil
}

func GetIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return 0, errors.New("input mission " + paramName + " field.")
	}
	paramValueInt, okAssert := strconv.Atoi(paramValue)
	if okAssert != nil {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param in int format.")
		return 0, errors.New("input teamID in wrong format.")
	}
	return paramValueInt, nil
}

func GetStringParamFromRequest(c *gin.Context, paramName string) (string, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_PARAM_FAILED, "please input param for request.")
		return "", errors.New("input mission " + paramName + " field.")
	}
	return paramValue, nil
}

func GetStringParamFromHeader(c *gin.Context, paramName string) (string, error) {
	paramValue := c.Request.Header[paramName]
	var ret string
	if len(paramValue) != 1 {
		return "", errors.New("can not fetch param from header.")
	} else {
		ret = paramValue[0]
	}
	return ret, nil
}

// @note: this param was setted by authenticator.JWTAuth() method
func GetUserIDFromAuth(c *gin.Context) (int, error) {
	// get request param
	userID, ok := c.Get("userID")
	if !ok {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "auth token invalied, can not fetch user ID in it.")
		return 0, errors.New("input mission userID field.")
	}
	userIDInt, okAssert := userID.(int)
	if !okAssert {
		FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_TOKEN_FAILED, "auth token invalied,user ID is not int type in it.")
		return 0, errors.New("input userID in wrong format.")
	}
	return userIDInt, nil
}

func FeedbackOK(c *gin.Context, resp repository.Response) {
	if resp != nil {
		c.JSON(http.StatusOK, resp.ExportForFeedback())
		return
	}
	// HTTP 200 with empty response
	c.JSON(http.StatusOK, nil)
}

func FeedbackBadRequest(c *gin.Context, errorFlag int, errorMessage string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"errorCode":    400,
		"errorFlag":    errorFlag,
		"errorMessage": errorMessage,
	})
	return
}

func FeedbackInternalServerError(c *gin.Context, errorFlag int, errorMessage string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"errorCode":    500,
		"errorFlag":    errorFlag,
		"errorMessage": errorMessage,
	})
	return
}
