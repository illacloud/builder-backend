package repository

import (
	"github.com/illacloud/builder-backend/internal/idconvertor"
)

type DeleteAppResponse struct {
	ID string `json:"id"`
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
