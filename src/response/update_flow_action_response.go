package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type UpdateFlowActionResponse struct {
	FlowActionID      string                 `json:"flowActionID"`
	UID               uuid.UUID              `json:"uid"`
	TeamID            string                 `json:"teamID"`
	WorkflowID        string                 `json:"workflowID"`
	Version           int                    `json:"version"`
	ResourceID        string                 `json:"resourceID,omitempty"`
	DisplayName       string                 `json:"displayName"`
	FlowActionType    string                 `json:"actionType"`
	IsVirtualResource bool                   `json:"isVirtualResource"`
	Content           map[string]interface{} `json:"content"`
	Transformer       map[string]interface{} `json:"transformer"`
	TriggerMode       string                 `json:"triggerMode"`
	Config            map[string]interface{} `json:"config"`
	CreatedAt         time.Time              `json:"createdAt,omitempty"`
	CreatedBy         string                 `json:"createdBy,omitempty"`
	UpdatedAt         time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy         string                 `json:"updatedBy,omitempty"`
}

func NewUpdateFlowActionResponse(action *model.FlowAction) *UpdateFlowActionResponse {
	actionConfig := action.ExportConfig()
	resp := &UpdateFlowActionResponse{
		FlowActionID:      idconvertor.ConvertIntToString(action.ID),
		UID:               action.UID,
		TeamID:            idconvertor.ConvertIntToString(action.TeamID),
		WorkflowID:        idconvertor.ConvertIntToString(action.WorkflowID),
		Version:           action.Version,
		ResourceID:        idconvertor.ConvertIntToString(action.ResourceID),
		DisplayName:       action.Name,
		FlowActionType:    resourcelist.GetResourceIDMappedType(action.Type),
		IsVirtualResource: actionConfig.IsVirtualResource,
		Content:           action.ExportTemplateInMap(),
		Transformer:       action.ExportTransformerInMap(),
		TriggerMode:       action.TriggerMode,
		Config:            action.ExportConfigInMap(),
		CreatedAt:         action.CreatedAt,
		CreatedBy:         idconvertor.ConvertIntToString(action.CreatedBy),
		UpdatedAt:         action.UpdatedAt,
		UpdatedBy:         idconvertor.ConvertIntToString(action.UpdatedBy),
	}
	return resp
}

func (resp *UpdateFlowActionResponse) ExportForFeedback() interface{} {
	return resp
}
