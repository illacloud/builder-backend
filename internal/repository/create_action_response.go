package repository

import (
	"github.com/illacloud/builder-backend/internal/idconvertor"
)

type CreateActionResponse struct {
	ID string `json:"id"`
}

func NewCreateActionResponse(id int) *CreateActionResponse {
	resp := &CreateActionResponse{
		ID: idconvertor.ConvertIntToString(id),
	}
	return resp
}

func (resp *CreateActionResponse) ExportForFeedback() interface{} {
	return resp
}
