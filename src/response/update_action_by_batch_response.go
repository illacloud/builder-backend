package response

import (
	"github.com/illacloud/builder-backend/src/model"
)

type UpdateActionByBatchResponse struct {
	Actions []*UpdateActionResponse `json:"actions"`
}

func NewUpdateActionByBatchResponse(actions []*model.Action) *UpdateActionByBatchResponse {
	resp := make([]*UpdateActionResponse, 0)
	for _, action := range actions {
		resp = append(resp, NewUpdateActionResponse(action))
	}
	return &UpdateActionByBatchResponse{Actions: resp}
}

func (resp *UpdateActionByBatchResponse) ExportForFeedback() interface{} {
	return resp
}
