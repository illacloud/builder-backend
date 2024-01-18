package response

import (
	"github.com/illacloud/builder-backend/src/model"
)

type UpdateFlowActionByBatchResponse struct {
	FlowActions []*UpdateFlowActionResponse `json:"flowActions"`
}

func NewUpdateFlowActionByBatchResponse(FlowActions []*model.FlowAction) *UpdateFlowActionByBatchResponse {
	resp := make([]*UpdateFlowActionResponse, 0)
	for _, FlowAction := range FlowActions {
		resp = append(resp, NewUpdateFlowActionResponse(FlowAction))
	}
	return &UpdateFlowActionByBatchResponse{FlowActions: resp}
}

func (resp *UpdateFlowActionByBatchResponse) ExportForFeedback() interface{} {
	return resp
}
