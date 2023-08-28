package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type Resource struct {
	ID        int       `gorm:"column:id;type:bigserial;primary_key"`
	UID       uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `gorm:"column:team_id;type:bigserial"`
	Name      string    `gorm:"column:name;type:varchar;size:200;not null"`
	Type      int       `gorm:"column:type;type:smallint;not null"`
	Options   string    `gorm:"column:options;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy int       `gorm:"column:updated_by;type:bigint;not null"`
}

func NewResource() *Resource {
	return &Resource{}
}

func NewResourceByCreateResourceRequest(teamID int, userID int, req *request.CreateResourceRequest) *Resource {
	resource := &Resource{
		TeamID:    teamID,
		Name:      req.ResourceName,
		Type:      resourcelist.GetResourceNameMappedID(req.ResourceType),
		Options:   req.ExportOptionsInString(),
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	resource.InitUID()
	resource.InitCreatedAt()
	resource.InitUpdatedAt()
	return resource
}

func NewResourceByTestResourceConnectionRequest(teamID int, userID int, req *request.TestResourceConnectionRequest) *Resource {
	resource := &Resource{
		TeamID:    teamID,
		Name:      req.ResourceName,
		Type:      resourcelist.GetResourceNameMappedID(req.ResourceType),
		Options:   req.ExportOptionsInString(),
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	resource.InitUID()
	resource.InitCreatedAt()
	resource.InitUpdatedAt()
	return resource
}

func (resource *Resource) UpdateByUpdateResourceRequest(userID int, req *request.UpdateResourceRequest) {
	resource.Name = req.ResourceName
	resource.Type = resourcelist.GetResourceNameMappedID(req.ResourceType)
	resource.Options = req.ExportOptionsInString()
	resource.UpdatedBy = userID
	resource.InitUpdatedAt()
}

func (resource *Resource) UpdateGoogleSheetOAuth2Options(userID int, options *ResourceOptionGoogleSheets) {
	resource.Options = options.ExportInString()
	resource.UpdatedBy = userID
	resource.InitUpdatedAt()
}

func (resource *Resource) CleanID() {
	resource.ID = 0
}

func (resource *Resource) InitUID() {
	resource.UID = uuid.New()
}

func (resource *Resource) InitCreatedAt() {
	resource.CreatedAt = time.Now().UTC()
}

func (resource *Resource) InitUpdatedAt() {
	resource.UpdatedAt = time.Now().UTC()
}

func (resource *Resource) ExportUpdatedAt() time.Time {
	return resource.UpdatedAt
}

func (resource *Resource) ExportType() int {
	return resource.Type
}

func (resource *Resource) ExportTypeInString() string {
	return resourcelist.GetResourceIDMappedType(resource.Type)
}

func (resource *Resource) ExportOptionsInMap() map[string]interface{} {
	var options map[string]interface{}
	json.Unmarshal([]byte(resource.Options), &options)
	return options
}

func (resource *Resource) CanCreateOAuthToken() bool {
	return resourcelist.CanCreateOAuthToken(resource.Type)
}
