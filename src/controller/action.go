package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
)

func (controller *Controller) CreateAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_CREATE_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch payload
	createActionRequest := request.NewCreateActionRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&createActionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// create
	newAction, errInCreateRequest := controller.createAction(c, teamID, appID, userID, createActionRequest)
	if errInCreateRequest != nil {
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewCreateActionResponse(newAction))
}

func (controller *Controller) CreateActionByBatch(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_CREATE_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch payload
	createActionByBatchRequest := request.NewCreateActionByBatchRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&createActionByBatchRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// create by batch
	newActions := make([]*model.Action, 0)
	for _, createActionRequest := range createActionByBatchRequest.ExportActions() {
		newAction, errInCreateRequest := controller.createAction(c, teamID, appID, userID, createActionRequest)
		if errInCreateRequest != nil {
			return
		}
		newActions = append(newActions, newAction)
	}

	// feedback
	controller.FeedbackOK(c, response.NewCreateActionByBatchResponse(newActions))
}

func (controller *Controller) UpdateAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	actionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		actionID,
		accesscontrol.ACTION_MANAGE_EDIT_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch payload
	updateActionRequest := request.NewUpdateActionRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&updateActionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// update
	newInDatabaseAction, errInUpdateAction := controller.updateAction(c, teamID, appID, userID, actionID, updateActionRequest)
	if errInUpdateAction != nil {
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewUpdateActionResponse(newInDatabaseAction))
}

func (controller *Controller) UpdateActionByBatch(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAPPID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		0,
		accesscontrol.ACTION_MANAGE_EDIT_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch payload
	updateActionByBatchRequest := request.NewUpdateActionByBatchRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&updateActionByBatchRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	inDatabaseActions := make([]*model.Action, 0)
	for _, updateActionRequest := range updateActionByBatchRequest.ExportActions() {
		newInDatabaseAction, errInUpdateAction := controller.updateAction(c, teamID, appID, userID, updateActionRequest.ExportActionIDInInt(), updateActionRequest)
		if errInUpdateAction != nil {
			return
		}
		inDatabaseActions = append(inDatabaseActions, newInDatabaseAction)
	}

	// feedback
	controller.FeedbackOK(c, response.NewUpdateActionByBatchResponse(inDatabaseActions))
}

func (controller *Controller) DeleteAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	actionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanDelete(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		actionID,
		accesscontrol.ACTION_DELETE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// delete
	errInDelete := controller.Storage.ActionStorage.DeleteActionByTeamIDAndActionID(teamID, actionID)
	if errInDelete != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_DELETE_ACTION, "delete action error: "+errInDelete.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewDeleteActionResponse(actionID))
	return
}

func (controller *Controller) GetAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	actionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		actionID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	action, errInGetAction := controller.Storage.ActionStorage.RetrieveActionByTeamIDAndID(teamID, actionID)
	if errInGetAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action error: "+errInGetAction.Error())
		return
	}

	// new response
	getActionResponse := response.NewGetActionResponse(action)

	// append remote virtual resource
	if action.IsRemoteVirtualAction() {
		api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
		if errInNewAPI != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
			return
		}
		virtualResource, errInGetVirtualResource := api.GetResource(action.ExportType(), action.ExportResourceID())
		if errInGetVirtualResource != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetVirtualResource.Error())
			return
		}
		getActionResponse.AppendVirtualResourceToTemplate(virtualResource)
	}

	// feedback
	controller.FeedbackOK(c, getActionResponse)
	return
}

func (controller *Controller) RunAction(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	appID, errInGetAppID := controller.GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	actionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAppID != nil || errInGetActionID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		actionID,
		accesscontrol.ACTION_MANAGE_RUN_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// set resource timing header
	// @see:
	// [Timing-Allow-Origin](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Timing-Allow-Origin)
	// [Resource_timing](https://developer.mozilla.org/en-US/docs/Web/API/Performance_API/Resource_timing)
	c.Header("Timing-Allow-Origin", "*")

	// execute
	runActionRequest := request.NewRunActionRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&runActionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error"+err.Error())
		return
	}

	// get action
	action := model.NewAction()
	fmt.Printf("[RetrieveActionsByTeamIDActionID] teamID: %d, actionID: %d\n", teamID, actionID)

	// actionID has not been created (like actionID is 0 'ILAfx4p1C7d0'), but we still can run it (onboarding case)
	if !model.DoesActionHasBeenCreated(actionID) {
		// ok, action was not created, fetch app and build a temporary action.
		app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
		if errInRetrieveApp != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
			return
		}
		action = model.NewAcitonByRunActionRequest(app, userID, runActionRequest)
	} else {
		// ok, we retrieve action from database
		var errInRetrieveAction error
		action, errInRetrieveAction = controller.Storage.ActionStorage.RetrieveActionByTeamIDActionID(teamID, actionID)
		if errInRetrieveAction != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errInRetrieveAction.Error())
			return
		}
	}

	// update action data with run action reqeust
	action.UpdateWithRunActionRequest(runActionRequest, userID)
	fmt.Printf("[DUMP] action: %+v\n", action)

	// assembly action
	actionFactory := model.NewActionFactoryByAction(action)
	actionAssemblyLine, errInBuild := actionFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return
	}

	// get resource
	resource := model.NewResource()
	if !action.IsVirtualAction() {
		// process normal resource action
		var errInRetrieveResource error
		resource, errInRetrieveResource = controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, action.ExportResourceID())
		if errInRetrieveResource != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resource failed: "+errInRetrieveResource.Error())
			return
		}
		// resource option validate only happend in create or update phrase
		// note that validate will set resprce options to actionAssemblyLine
		_, errInValidateResourceOptions := actionAssemblyLine.ValidateResourceOptions(resource.ExportOptionsInMap())
		if errInValidateResourceOptions != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_RESOURCE_FAILED, "validate resource failed: "+errInValidateResourceOptions.Error())
			return
		}
	} else {
		// process virtual resource action
		action.AppendRuntimeInfoForVirtualResource(userAuthToken, teamID)
	}

	// check action template
	fmt.Printf("[DUMP] action.ExportTemplateInMap(): %+v\n", action.ExportTemplateInMap())
	fmt.Printf("[DUMP] action.ExportRawTemplateInMap(): %+v\n", action.ExportRawTemplateInMap())
	_, errInValidate := actionAssemblyLine.ValidateActionTemplate(action.ExportTemplateInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action template error: "+errInValidate.Error())
		return
	}

	// run
	log.Printf("[DUMP]action: %+v\n", action)
	log.Printf("[DUMP] resource.ExportOptionsInMap(): %+v, action.ExportTemplateInMap(): %+v\n", resource.ExportOptionsInMap(), action.ExportTemplateInMap())
	actionRunResult, errInRunAction := actionAssemblyLine.Run(resource.ExportOptionsInMap(), action.ExportTemplateInMap(), action.ExportRawTemplateInMap())
	if errInRunAction != nil {
		if strings.HasPrefix(errInRunAction.Error(), "Error 1064:") {
			lineNumber, _ := strconv.Atoi(errInRunAction.Error()[len(errInRunAction.Error())-1:])
			message := ""
			regexp, _ := regexp.Compile(`to use`)
			match := regexp.FindStringIndex(errInRunAction.Error())
			if len(match) == 2 {
				message = errInRunAction.Error()[match[1]:]
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"errorCode":    400,
				"errorFlag":    ERROR_FLAG_EXECUTE_ACTION_FAILED,
				"errorMessage": errors.New("SQL syntax error").Error(),
				"errorData": map[string]interface{}{
					"lineNumber": lineNumber,
					"message":    "SQL syntax error" + message,
				},
			})
			return
		}
		controller.FeedbackBadRequest(c, ERROR_FLAG_EXECUTE_ACTION_FAILED, "run action error: "+errInRunAction.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, actionRunResult)
}
