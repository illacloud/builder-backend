package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

func (controller *Controller) duplicateFlowActionByVersion(c *gin.Context, fromTeamID int, toTeamID int, fromWorkflowID int, toWorkflowID int, fromVersion int, toVersion int, modifierID int, isForkWorkflow bool) (map[int]int, error) {
	idMap := make(map[int]int, 0)
	// get target version flow action from database
	flowActions, errinRetrieveFlowAction := controller.Storage.FlowActionStorage.RetrieveFlowActionsByTeamIDWorkflowIDAndVersion(fromTeamID, fromWorkflowID, fromVersion)
	if errinRetrieveFlowAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errinRetrieveFlowAction.Error())
		return nil, errinRetrieveFlowAction
	}

	// set fork info
	oldIDList := make([]int, 0)
	for serial, flowAction := range flowActions {
		oldIDList = append(oldIDList, flowAction.ID)
		flowActions[serial].InitForFork(toTeamID, toWorkflowID, toVersion, modifierID)
	}

	// and put them to the database as duplicate
	resourceManagerSDK, errInNewResourceManagerSDK := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
	if errInNewResourceManagerSDK != nil {
		return nil, errInNewResourceManagerSDK
	}

	for serial, flowAction := range flowActions {
		// check if action is ai-agent, and if ai-agent is public, and we are forking app from marketplace (not publish app to marketplace) fork it automatically
		if flowAction.Type == resourcelist.TYPE_AI_AGENT_ID && isForkWorkflow {
			fmt.Printf("[DUMP] DuplicateFlowActionByVersion: hit AI_AGENT action\n")
			// call resource manager for for ai-agent
			forkedAIAgent, errInForkAiAgent := resourceManagerSDK.ForkMarketplaceAIAgent(flowAction.ExportResourceID(), toTeamID, modifierID)
			fmt.Printf("[DUMP] DuplicateFlowActionByVersion() forkedAIAgent: %+v\n", forkedAIAgent)
			fmt.Printf("[DUMP] DuplicateFlowActionByVersion() errInForkAiAgent: %+v\n", errInForkAiAgent)
			if errInForkAiAgent == nil {
				flowAction.SetResourceIDByAiAgent(forkedAIAgent)
			}
		}
		fmt.Printf("[DUMP] DuplicateFlowActionByVersion() action: %+v\n", flowAction)

		// create action
		newFlowActionID, errInCreateFlowAction := controller.Storage.FlowActionStorage.Create(flowAction)
		if errInCreateFlowAction != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action failed: "+errInCreateFlowAction.Error())
			return nil, errInCreateFlowAction
		}
		idMap[oldIDList[serial]] = newFlowActionID
	}
	return idMap, nil
}
