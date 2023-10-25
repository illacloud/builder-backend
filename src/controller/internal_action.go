package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
	"github.com/illacloud/builder-backend/src/utils/illacloudperipheralapisdk"
)

func (controller *Controller) GenerateSQL(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// fetch payload
	generateSQLRequest := request.NewGenerateSQLRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&generateSQLRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(generateSQLRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}
	resourceID := generateSQLRequest.ExportResourceIDInInt()

	// validate sql generate special management
	canManageSpecial, errInCheckAttr := controller.AttributeGroup.CanManageSpecial(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_PERIPHERAL_SERVICE,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_SPECIAL_GENERATE_SQL,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManageSpecial {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// validate resource access
	canAccessResource, errInCheckResourceAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckResourceAttr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    500,
			"errorMessage": "error in check attribute: " + errInCheckResourceAttr.Error(),
		})
		return
	}
	if !canAccessResource {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType: auditlogger.AUDIT_LOG_TRIGGER_TASK,
		TeamID:    teamID,
		UserID:    userID,
		IP:        c.ClientIP(),
		TaskName:  auditlogger.TASK_GENERATE_SQL,
		TaskInput: map[string]interface{}{"content": generateSQLRequest.Description},
	})

	// fetch resource
	resource, errInGetResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInGetResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "error in fetch resource: "+errInGetResource.Error())
		return
	}

	// fetch resource meta info
	actionFactory := model.NewActionFactoryByResource(resource)
	actionAssemblyLine, errInBuild := actionFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return
	}
	resourceMetaInfo, errInGetMetaInfo := actionAssemblyLine.GetMetaInfo(resource.ExportOptionsInMap())
	if errInGetMetaInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE_META_INFO, "error in fetch resource meta info: "+errInGetMetaInfo.Error())
		return
	}

	// form request payload
	generateSQLPeripheralRequest, errInNewReq := illacloudperipheralapisdk.NewGenerateSQLPeripheralRequest(resource.ExportTypeInString(), resourceMetaInfo.ExportSchema(), generateSQLRequest.Description, generateSQLRequest.GetActionInString())
	if errInNewReq != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate request failed: "+errInNewReq.Error())
		return
	}

	// call remote generate sql API
	peripheralAPI := illacloudperipheralapisdk.NewIllaCloudPeriphearalAPI()
	generateSQLResponse, errInGGenerateSQL := peripheralAPI.GenerateSQL(generateSQLPeripheralRequest)
	if errInGGenerateSQL != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_GENERATE_SQL_FAILED, "generate sql failed: "+errInGGenerateSQL.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, generateSQLResponse)
}
