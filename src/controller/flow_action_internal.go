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
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
)

func (controller *Controller) GetWorkflowAllFlowActionsInternal(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	teamIDInString, errInGetTeamIDInString := controller.GetStringParamFromRequest(c, PARAM_TEAM_ID)
	workflowID, errInGetWorkflowID := controller.GetMagicIntParamFromRequest(c, PARAM_WORKFLOW_ID)
	workflowIDInString, errInGetworkflowIDInString := controller.GetStringParamFromRequest(c, PARAM_WORKFLOW_ID)
	if errInGetTeamID != nil || errInGetWorkflowID != nil || errInGetTeamIDInString != nil || errInGetworkflowIDInString != nil {
		return
	}

	// validate request data
	validated, errInValidate := controller.ValidateRequestTokenFromHeader(c, teamIDInString, workflowIDInString)
	if !validated && errInValidate != nil {
		return
	}

	// fetch data
	flowActions, errInGetActions := controller.Storage.FlowActionStorage.RetrieveAll(teamID, workflowID)
	if errInGetActions != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_FLOW_ACTION, "get workflow all flowActions error: "+errInGetActions.Error())
		return
	}

	// build remote virtual resource lookup table
	virtualResourceLT := make(map[int]map[string]interface{}, 0)
	api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
	if errInNewAPI != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_FLOW_ACTION, "error in fetch flowAction mapped virtual resource: "+errInNewAPI.Error())
		return
	}
	for _, flowAction := range flowActions {
		if flowAction.IsRemoteVirtualFlowAction() {
			virtualResource, errInGetVirtualResource := api.GetResource(flowAction.ExportType(), flowAction.ExportResourceID())
			if errInGetVirtualResource != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_FLOW_ACTION, "error in fetch flowAction mapped virtual resource: "+errInGetVirtualResource.Error())
				return
			}
			virtualResourceLT[flowAction.ExportID()] = virtualResource
		}
	}

	// new response
	getAllFlowActionResponse := response.NewGetWorkflowAllFlowActionsResponse(flowActions, virtualResourceLT)

	// feedback
	controller.FeedbackOK(c, getAllFlowActionResponse)
	return
}

func (controller *Controller) RunFlowActionInternal(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	workflowID, errInGetAppID := controller.GetMagicIntParamFromRequest(c, PARAM_WORKFLOW_ID)
	flowActionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_FLOW_ACTION_ID)
	if errInGetTeamID != nil || errInGetAppID != nil || errInGetActionID != nil {
		return
	}

	// execute
	runFlowActionRequest := request.NewRunFlowActionRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&runFlowActionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error"+err.Error())
		return
	}

	// get flowAction
	fmt.Printf("[RetrieveActionsByTeamIDActionID] teamID: %d, flowActionID: %d\n", teamID, flowActionID)
	var errInRetrieveAction error
	flowAction, errInRetrieveAction := controller.Storage.FlowActionStorage.RetrieveFlowActionByTeamIDFlowActionID(teamID, flowActionID)
	if errInRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_FLOW_ACTION, "get flowAction failed: "+errInRetrieveAction.Error())
		return
	}

	// update flowAction data with run flowAction reqeust
	flowAction.UpdateWithRunFlowActionRequest(runFlowActionRequest, model.ANONYMOUS_USER_ID)
	fmt.Printf("[DUMP] flowAction: %+v\n", flowAction)

	// assembly flowAction
	flowActionFactory := model.NewFlowActionFactoryByFlowAction(flowAction)
	flowActionAssemblyLine, errInBuild := flowActionFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate flowAction type error: "+errInBuild.Error())
		return
	}

	// get resource
	resource := model.NewResource()
	if !flowAction.IsVirtualFlowAction() {
		// process normal resource flowAction
		var errInRetrieveResource error
		resource, errInRetrieveResource = controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, flowAction.ExportResourceID())
		if errInRetrieveResource != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resource failed: "+errInRetrieveResource.Error())
			return
		}
		// resource option validate only happend in create or update phrase
		// note that validate will set resprce options to flowActionAssemblyLine
		_, errInValidateResourceOptions := flowActionAssemblyLine.ValidateResourceOptions(resource.ExportOptionsInMap())
		if errInValidateResourceOptions != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_RESOURCE_FAILED, "validate resource failed: "+errInValidateResourceOptions.Error())
			return
		}
	} else {
		// process virtual resource flowAction
		flowAction.AppendRuntimeInfoForVirtualResource(userAuthToken, teamID)
	}

	// check flowAction template
	fmt.Printf("[DUMP] flowAction.ExportTemplateInMap(): %+v\n", flowAction.ExportTemplateInMap())
	fmt.Printf("[DUMP] flowAction.ExportRawTemplateInMap(): %+v\n", flowAction.ExportRawTemplateInMap())
	_, errInValidate := flowActionAssemblyLine.ValidateActionTemplate(flowAction.ExportTemplateInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate flowAction template error: "+errInValidate.Error())
		return
	}

	// run
	log.Printf("[DUMP]flowAction: %+v\n", flowAction)
	log.Printf("[DUMP] resource.ExportOptionsInMap(): %+v, flowAction.ExportTemplateInMap(): %+v\n", resource.ExportOptionsInMap(), flowAction.ExportTemplateInMap())
	flowActionRunResult, errInRunAction := flowActionAssemblyLine.Run(resource.ExportOptionsInMap(), flowAction.ExportTemplateInMap(), flowAction.ExportRawTemplateInMap())
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
				"errorFlag":    ERROR_FLAG_EXECUTE_FLOW_ACTION_FAILED,
				"errorMessage": errors.New("SQL syntax error").Error(),
				"errorData": map[string]interface{}{
					"lineNumber": lineNumber,
					"message":    "SQL syntax error" + message,
				},
			})
			return
		}
		controller.FeedbackBadRequest(c, ERROR_FLAG_EXECUTE_FLOW_ACTION_FAILED, "run flowAction error: "+errInRunAction.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, flowActionRunResult)
}
