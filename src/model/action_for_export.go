package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

const (
	ACTION_FOR_EXPORT_FIELD_VIRTUAL_RESOURCE = "virtualResource"
)

type ActionForExport struct {
	ID                string                 `json:"actionID"`
	UID               uuid.UUID              `json:"uid"`
	TeamID            string                 `json:"teamID"`
	App               string                 `json:"-"`
	Version           int                    `json:"-"`
	Resource          string                 `json:"resourceID,omitempty"`
	DisplayName       string                 `json:"displayName"`
	Type              string                 `json:"actionType"`
	IsVirtualResource bool                   `json:"isVirtualResource"`
	Template          map[string]interface{} `json:"content"`
	Transformer       map[string]interface{} `json:"transformer"`
	TriggerMode       string                 `json:"triggerMode"`
	Config            *ActionConfig          `json:"config"`
	CreatedAt         time.Time              `json:"createdAt,omitempty"`
	CreatedBy         string                 `json:"createdBy,omitempty"`
	UpdatedAt         time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy         string                 `json:"updatedBy,omitempty"`
}

func NewActionForExport(action *Action) *ActionForExport {
	return &ActionForExport{
		ID:                idconvertor.ConvertIntToString(action.ID),
		UID:               action.UID,
		TeamID:            idconvertor.ConvertIntToString(action.TeamID),
		App:               idconvertor.ConvertIntToString(action.AppRefID),
		Version:           action.Version,
		Resource:          idconvertor.ConvertIntToString(action.ResourceRefID),
		DisplayName:       action.ExportDisplayName(),
		Type:              action.ExportTypeInString(),
		IsVirtualResource: resourcelist.IsVirtualResourceByIntType(action.ExportType()),
		Template:          action.ExportTemplateInMap(),
		Transformer:       action.ExportTransformerInMap(),
		TriggerMode:       action.TriggerMode,
		Config:            action.ExportConfig(),
		CreatedAt:         action.CreatedAt,
		CreatedBy:         idconvertor.ConvertIntToString(action.CreatedBy),
		UpdatedAt:         action.UpdatedAt,
		UpdatedBy:         idconvertor.ConvertIntToString(action.UpdatedBy),
	}
}

func (resp *ActionForExport) ExportDisplayName() string {
	return resp.DisplayName
}

func (resp *ActionForExport) ExportTeamID() string {
	return resp.TeamID
}

func (resp *ActionForExport) ExportResourceID() string {
	return resp.Resource
}

func (resp *ActionForExport) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(resp.Resource)
}

func (resp *ActionForExport) AppendVirtualResourceToTemplate(value interface{}) {
	resp.Template[ACTION_FOR_EXPORT_FIELD_VIRTUAL_RESOURCE] = value
}

func (resp *ActionForExport) ExportForFeedback() interface{} {
	return resp
}

func NewActionForExportByMap(data interface{}) (*ActionForExport, error) {
	udata, ok := data.(map[string]interface{})
	if !ok {
		err := errors.New("NewActionForExportByMap failed, please check your input.")
		return nil, err
	}
	displayName, mapok := udata["displayName"].(string)
	if !mapok {
		err := errors.New("NewActionForExportByMap failed, can not find displayName field.")
		return nil, err
	}
	// fill field
	actionForExport := &ActionForExport{
		DisplayName: displayName,
	}
	return actionForExport, nil
}
