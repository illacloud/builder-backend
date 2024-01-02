package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type CreateFlowActionResponse struct {
	FlowActionID      string                 `json:"flowActionID"`
	UID               uuid.UUID              `json:"uid"`
	TeamID            string                 `json:"teamID"`
	WorkflowID        string                 `json:"workflowID"`
	Version           int                    `json:"version"`
	ResourceID        string                 `json:"resourceID,omitempty"`
	DisplayName       string                 `json:"displayName"`
	ActionType        string                 `json:"flowActionType"`
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

func NewCreateFlowActionResponse(flowAction *model.FlowAction) *CreateFlowActionResponse {
	flowActionConfig := flowAction.ExportConfig()
	resp := &CreateFlowActionResponse{
		FlowActionID:      idconvertor.ConvertIntToString(flowAction.ID),
		UID:               flowAction.UID,
		TeamID:            idconvertor.ConvertIntToString(flowAction.TeamID),
		WorkflowID:        idconvertor.ConvertIntToString(flowAction.WorkflowID),
		Version:           flowAction.Version,
		ResourceID:        idconvertor.ConvertIntToString(flowAction.ResourceID),
		DisplayName:       flowAction.Name,
		ActionType:        resourcelist.GetResourceIDMappedType(flowAction.Type),
		IsVirtualResource: flowActionConfig.IsVirtualResource,
		Content:           flowAction.ExportTemplateInMap(),
		Transformer:       flowAction.ExportTransformerInMap(),
		TriggerMode:       flowAction.TriggerMode,
		Config:            flowAction.ExportConfigInMap(),
		CreatedAt:         flowAction.CreatedAt,
		CreatedBy:         idconvertor.ConvertIntToString(flowAction.CreatedBy),
		UpdatedAt:         flowAction.UpdatedAt,
		UpdatedBy:         idconvertor.ConvertIntToString(flowAction.UpdatedBy),
	}
	return resp
}

func (resp *CreateFlowActionResponse) ExportForFeedback() interface{} {
	return resp
}
