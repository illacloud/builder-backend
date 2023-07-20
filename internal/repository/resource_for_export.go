package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/internal/util/resourcelist"
)

type ResourceForExport struct {
	ID        string                 `json:"resourceId"`
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
		Options:   r.Options,
		CreatedAt: r.CreatedAt,
		CreatedBy: idconvertor.ConvertIntToString(r.CreatedBy),
		UpdatedAt: r.UpdatedAt,
		UpdatedBy: idconvertor.ConvertIntToString(r.UpdatedBy),
	}
}

func (resp *ResourceForExport) ExportName() string {
	return resp.Name
}

func (resp *ResourceForExport) ExportForFeedback() interface{} {
	return resp
}

func (resp *ResourceForExport) ExportResource() Resource {
	resourceDto := Resource{
		UID:       resp.UID,
		Name:      resp.Name,
		Type:      resourcelist.GetResourceNameMappedID(resp.Type),
		Options:   resp.Options,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
	if resp.TeamID != "" {
		resourceDto.TeamID = idconvertor.ConvertStringToInt(resp.TeamID)
	}
	if resp.ID != "" {
		resourceDto.ID = idconvertor.ConvertStringToInt(resp.ID)
	}
	if resp.CreatedBy != "" {
		resourceDto.CreatedBy = idconvertor.ConvertStringToInt(resp.CreatedBy)
	}
	if resp.UpdatedBy != "" {
		resourceDto.UpdatedBy = idconvertor.ConvertStringToInt(resp.UpdatedBy)
	}
	return resourceDto

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
