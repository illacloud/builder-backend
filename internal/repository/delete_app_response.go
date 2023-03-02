package repository

import (
	"github.com/illacloud/illa-builder-backend/internal/idconvertor"
)

type DeleteAppResponse struct {
	ID string `json:"appID"`
}

func NewDeleteAppResponse(id int) *DeleteAppResponse {
	resp := &DeleteAppResponse{
		ID: idconvertor.ConvertIntToString(id),
	}
	return resp
}

func (resp *DeleteAppResponse) ExportForFeedback() interface{} {
	return resp
}
