package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type ActionSummaryForExport struct {
	ID                string    `json:"actionID"`
	UID               uuid.UUID `json:"uid"`
	TeamID            string    `json:"teamID"`
	Version           int       `json:"version"`
	Resource          string    `json:"resourceID,omitempty"`
	DisplayName       string    `json:"name"`
	Icon              string    `json:"icon"`
	Type              string    `json:"type"`
	IsVirtualResource bool      `json:"isVirtualResource"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt,omitempty"`
	UpdatedBy         string    `json:"updatedBy,omitempty"`
}

func NewActionSummaryForExportByAction(action *Action) *ActionSummaryForExport {
	return &ActionSummaryForExport{
		ID:                idconvertor.ConvertIntToString(action.ID),
		UID:               action.UID,
		TeamID:            idconvertor.ConvertIntToString(action.TeamID),
		Version:           action.Version,
		Resource:          idconvertor.ConvertIntToString(action.ResourceRefID),
		DisplayName:       action.ExportDisplayName(),
		Icon:              action.ExportIcon(),
		Type:              action.ExportTypeInString(),
		IsVirtualResource: resourcelist.IsVirtualResourceByIntType(action.ExportType()),
		CreatedAt:         action.CreatedAt,
		CreatedBy:         idconvertor.ConvertIntToString(action.CreatedBy),
		UpdatedAt:         action.UpdatedAt,
		UpdatedBy:         idconvertor.ConvertIntToString(action.UpdatedBy),
	}
}
