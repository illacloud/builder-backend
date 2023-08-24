package response

import (
	"github.com/illacloud/builder-backend/internal/idconvertor"
)

type DeleteResourceResponse struct {
	ID string `json:"resourceId"`
}

func NewDeleteResourceResponse(id int) *DeleteResourceResponse {
	resp := &DeleteResourceResponse{
		ID: idconvertor.ConvertIntToString(id),
	}
	return resp
}

func (resp *DeleteResourceResponse) ExportForFeedback() interface{} {
	return resp
}
