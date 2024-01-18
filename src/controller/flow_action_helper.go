package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
)

func (controller *Controller) updateFlowAction(c *gin.Context, teamID int, workflowID int, userID int, flowActionID int, updateFlowActionRequest *request.UpdateFlowActionRequest) (*model.FlowAction, error) {
	// append remote virtual resource (like aiagent, but the transformet is local virtual resource)
	if updateFlowActionRequest.IsRemoteVirtualAction() {
		// the AI_Agent need fetch resource info from resource manager, but illa drive does not need that
		if updateFlowActionRequest.NeedFetchResourceInfoFromSourceManager() {
			api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
			if errInNewAPI != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_FLOW_ACTION, "error in fetch flowAction mapped virtual resource: "+errInNewAPI.Error())
				return nil, errInNewAPI
			}
			virtualResource, errInGetVirtualResource := api.GetResource(updateFlowActionRequest.ExportFlowActionTypeInInt(), updateFlowActionRequest.ExportResourceIDInInt())
			if errInGetVirtualResource != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_FLOW_ACTION, "error in fetch flowAction mapped virtual resource: "+errInGetVirtualResource.Error())
				return nil, errInGetVirtualResource
			}
			updateFlowActionRequest.AppendVirtualResourceToTemplate(virtualResource)
		}
	}

	// get flowAction
	inDatabaseFlowAction, errInRetrieveFlowAction := controller.Storage.FlowActionStorage.RetrieveFlowActionByTeamIDFlowActionID(teamID, flowActionID)
	if errInRetrieveFlowAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_FLOW_ACTION, "get app failed: "+errInRetrieveFlowAction.Error())
		return nil, errInRetrieveFlowAction
	}

	// update inDatabaseFlowAction instance
	inDatabaseFlowAction.UpdateFlowAcitonByUpdateFlowActionRequest(teamID, workflowID, userID, updateFlowActionRequest)

	// validate flowAction options
	errInValidateActionOptions := controller.ValidateFlowActionTemplate(c, inDatabaseFlowAction)
	if errInValidateActionOptions != nil {
		return nil, errInValidateActionOptions
	}

	// update flowAction
	errInUpdateAction := controller.Storage.FlowActionStorage.UpdateWholeFlowAction(inDatabaseFlowAction)
	if errInUpdateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_FLOW_ACTION, "update flowAction error: "+errInUpdateAction.Error())
		return nil, errInUpdateAction
	}

	return inDatabaseFlowAction, nil
}
