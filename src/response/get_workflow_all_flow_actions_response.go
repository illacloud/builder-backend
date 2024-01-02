package response

import (
	"github.com/illacloud/builder-backend/src/model"
)

type GetWorkflowAllFlowActionsResponse struct {
	AllFlowActions []*GetFlowActionResponse `json:"allFlowActions"`
}

func NewGetWorkflowAllFlowActionsResponse(flowActions []*model.FlowAction, virtualResourceLT map[int]map[string]interface{}) *GetWorkflowAllFlowActionsResponse {
	flowActionsRet := make([]*GetFlowActionResponse, 0)
	for _, flowAction := range flowActions {
		flowActionResp := NewGetFlowActionResponse(flowAction)
		virtualResource, hitVirtualResource := virtualResourceLT[flowAction.ExportID()]
		if hitVirtualResource {
			flowActionResp.AppendVirtualResourceToTemplate(virtualResource)
		}
		flowActionsRet = append(flowActionsRet, flowActionResp)
	}
	ret := &GetWorkflowAllFlowActionsResponse{AllFlowActions: flowActionsRet}
	return ret
}

func (resp *GetWorkflowAllFlowActionsResponse) ExportForFeedback() interface{} {
	return resp
}

func NewEmptyGetWorkflowAllFlowActionsResponse() *GetWorkflowAllFlowActionsResponse {
	flowActionsRet := make([]*GetFlowActionResponse, 0)
	ret := &GetWorkflowAllFlowActionsResponse{AllFlowActions: flowActionsRet}
	return ret
}
