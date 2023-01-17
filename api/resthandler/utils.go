package resthandler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func GetUserAuthTokenFromHeader(c *gin.Context) (string, error) {
	// fetch token
	rawToken := c.Request.Header[PARAM_AUTHORIZATION]
	if len(rawToken) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "HTTP request header missing request token.",
		})
		return "", errors.New("missing request token.")
	}
	var token string
	token = rawToken[0]
	return token, nil
}

func GetIntParamFromRequest(c *gin.Context, paramName string) (int, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "please input team id for get team info.",
		})
		return 0, errors.New("input mission " + paramName + " field.")
	}
	paramValueInt, okAssert := strconv.Atoi(paramValue)
	if okAssert != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "please input team id in int format.",
		})
		return 0, errors.New("input teamID in wrong format.")
	}
	return paramValueInt, nil
}

func GetStringParamFromRequest(c *gin.Context, paramName string) (string, error) {
	// get request param
	paramValue := c.Param(paramName)
	if len(paramValue) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "please input team id for get team info.",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "authorization missing.",
		})
		return 0, errors.New("input mission userID field.")
	}
	userIDInt, okAssert := userID.(int)
	if !okAssert {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "assert authorization failed.",
		})
		return 0, errors.New("input userID in wrong format.")
	}
	return userIDInt, nil
}
