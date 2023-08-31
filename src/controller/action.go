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

	// append remote virtual resource (like aiagent, but the transformet is local virtual resource)
	if createActionRequest.IsRemoteVirtualAction() {
		api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
		if errInNewAPI != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
			return
		}
		virtualResource, errInGetVirtualResource := api.GetResource(createActionRequest.ExportActionTypeInInt(), createActionRequest.ExportResourceIDInInt())
		if errInGetVirtualResource != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetVirtualResource.Error())
			return
		}
		createActionRequest.AppendVirtualResourceToTemplate(virtualResource)
	}

	// get action mapped app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// init action instace
	action, errorInNewAction := model.NewAcitonByCreateActionRequest(app, userID, createActionRequest)
	if errorInNewAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in create action instance: "+errorInNewAction.Error())
		return
	}

	// validate action options
	errInValidateActionOptions := controller.ValidateActionTemplate(c, action)
	if errInValidateActionOptions != nil {
		return
	}

	// create action
	_, errInCreateAction := controller.Storage.ActionStorage.Create(action)
	if errInCreateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action error: "+errInCreateAction.Error())
		return
	}

	// update app updatedAt, updatedBy, editedBy field
	app.Modify(userID)
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app modify info error: "+errInUpdateApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewCreateActionResponse(action))
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

	// append remote virtual resource (like aiagent, but the transformet is local virtual resource)
	if updateActionRequest.IsRemoteVirtualAction() {
		api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
		if errInNewAPI != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
			return
		}
		virtualResource, errInGetVirtualResource := api.GetResource(updateActionRequest.ExportActionTypeInInt(), updateActionRequest.ExportResourceIDInInt())
		if errInGetVirtualResource != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetVirtualResource.Error())
			return
		}
		updateActionRequest.AppendVirtualResourceToTemplate(virtualResource)
	}

	// get action mapped app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return
	}

	// get action
	inDatabaseAction, errInRetrieveAction := controller.Storage.ActionStorage.RetrieveActionByTeamIDActionID(teamID, actionID)
	if errInRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveAction.Error())
		return
	}

	// update inDatabaseAction instance
	inDatabaseAction.UpdateAcitonByUpdateActionRequest(app, userID, updateActionRequest)

	// validate action options
	errInValidateActionOptions := controller.ValidateActionTemplate(c, inDatabaseAction)
	if errInValidateActionOptions != nil {
		return
	}

	// update action
	errInUpdateAction := controller.Storage.ActionStorage.UpdateWholeAction(inDatabaseAction)
	if errInUpdateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "update action error: "+errInUpdateAction.Error())
		return
	}

	// update app updatedAt, updatedBy, editedBy field
	app.Modify(userID)
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app modify info error: "+errInUpdateApp.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewUpdateActionResponse(inDatabaseAction))
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
	if model.DoesActionHasBeenCreated(actionID) {
		// fetch app
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
		action.AppendRuntimeInfoForVirtualResource(userAuthToken)
	}

	// check action template
	fmt.Printf("[DUMP] action.ExportTemplateInMap(): %+v\n", action.ExportTemplateInMap())
	_, errInValidate := actionAssemblyLine.ValidateActionTemplate(action.ExportTemplateInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action template error: "+errInValidate.Error())
		return
	}

	// run
	log.Printf("[DUMP]action: %+v\n", action)
	log.Printf("[DUMP] resource.ExportOptionsInMap(): %+v, action.ExportTemplateInMap(): %+v\n", resource.ExportOptionsInMap(), action.ExportTemplateInMap())
	actionRunResult, errInRunAction := actionAssemblyLine.Run(resource.ExportOptionsInMap(), action.ExportTemplateInMap())
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
