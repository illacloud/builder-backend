package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type ResourceForExport struct {
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

func NewResourceForExport(r *Resource) *ResourceForExport {
	return &ResourceForExport{
		ID:        idconvertor.ConvertIntToString(r.ID),
		UID:       r.UID,
		TeamID:    idconvertor.ConvertIntToString(r.TeamID),
		Name:      r.Name,
		Type:      resourcelist.GetResourceIDMappedType(r.Type),
		Options:   r.ExportOptionsInMap(),
		CreatedAt: r.CreatedAt,
		CreatedBy: idconvertor.ConvertIntToString(r.CreatedBy),
		UpdatedAt: r.UpdatedAt,
		UpdatedBy: idconvertor.ConvertIntToString(r.UpdatedBy),
	}
}

func BatchNewResourceForExport(r []*Resource) []*ResourceForExport {
	resourcesForExport := make([]*ResourceForExport, 0, len(r))
	for _, resource := range r {
		resourcesForExport = append(resourcesForExport, NewResourceForExport(resource))
	}
	return resourcesForExport
}

func (resp *ResourceForExport) ExportName() string {
	return resp.Name
}

func (resp *ResourceForExport) ExportOptionsInString() string {
	optionsByte, _ := json.Marshal(resp.Options)
	return string(optionsByte)
}

func (resp *ResourceForExport) ExportForFeedback() interface{} {
	return resp
}

func NewResourceForExportByMap(data interface{}) (*ResourceForExport, error) {
	udata, ok := data.(map[string]interface{})
	if !ok {
		err := errors.New("NewResourceForExportByMap failed, please check your input.")
		return nil, err
	}
	name, mapok := udata["name"].(string)
	if !mapok {
		err := errors.New("NewResourceForExportByMap failed, can not find name field.")
		return nil, err
	}
	// fill field
	resourceForExport := &ResourceForExport{
		Name: name,
	}
	return resourceForExport, nil
}
