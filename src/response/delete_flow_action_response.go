package response

import (
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

type DeleteFlowActionResponse struct {
	ID string `json:"flowActionID"`
}

func NewDeleteFlowActionResponse(id int) *DeleteFlowActionResponse {
	resp := &DeleteFlowActionResponse{
		ID: idconvertor.ConvertIntToString(id),
	}
	return resp
}

func (resp *DeleteFlowActionResponse) ExportForFeedback() interface{} {
	return resp
}
