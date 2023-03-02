package repository

import (
	"github.com/illacloud/illa-builder-backend/internal/idconvertor"
)

type DeleteActionResponse struct {
	ID string `json:"actionID"`
}

func NewDeleteActionResponse(id int) *DeleteActionResponse {
	resp := &DeleteActionResponse{
		ID: idconvertor.ConvertIntToString(id),
	}
	return resp
}

func (resp *DeleteActionResponse) ExportForFeedback() interface{} {
	return resp
}
