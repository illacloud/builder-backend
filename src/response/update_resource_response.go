package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type UpdateResourceResponse struct {
	ID        string                 `json:"resourceID"`
	UID       uuid.UUID              `json:"uid"`
	TeamID    string                 `json:"teamID"`
	Name      string                 `json:"resourceName" validate:"required"`
	Type      string                 `json:"resourceType" validate:"required"`
	Options   map[string]interface{} `json:"content" validate:"required"`
	CreatedAt time.Time              `json:"createdAt,omitempty"`
	CreatedBy string                 `json:"createdBy,omitempty"`
	UpdatedAt time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy string                 `json:"updatedBy,omitempty"`
}

func NewUpdateResourceResponse(resource *model.Resource) *UpdateResourceResponse {
	return &UpdateResourceResponse{
		ID:        idconvertor.ConvertIntToString(resource.ID),
		UID:       resource.UID,
		TeamID:    idconvertor.ConvertIntToString(resource.TeamID),
		Name:      resource.Name,
		Type:      resourcelist.GetResourceIDMappedType(resource.Type),
		Options:   resource.ExportOptionsInMap(),
		CreatedAt: resource.CreatedAt,
		CreatedBy: idconvertor.ConvertIntToString(resource.CreatedBy),
		UpdatedAt: resource.UpdatedAt,
		UpdatedBy: idconvertor.ConvertIntToString(resource.UpdatedBy),
	}
}

func (resp *UpdateResourceResponse) ExportForFeedback() interface{} {
	return resp
}
