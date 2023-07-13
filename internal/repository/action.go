// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/util/resourcelist"
	"github.com/illacloud/builder-backend/pkg/db"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Action struct {
	ID          int       `gorm:"column:id;type:bigserial;primary_key"`
	UID         uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	TeamID      int       `gorm:"column:team_id;type:bigserial"`
	App         int       `gorm:"column:app_ref_id;type:bigint;not null"`
	Version     int       `gorm:"column:version;type:bigint;not null"`
	Resource    int       `gorm:"column:resource_ref_id;type:bigint;not null"`
	Name        string    `gorm:"column:name;type:varchar;size:255;not null"`
	Type        int       `gorm:"column:type;type:smallint;not null"`
	TriggerMode string    `gorm:"column:trigger_mode;type:varchar;size:16;not null"`
	Transformer db.JSONB  `gorm:"column:transformer;type:jsonb"`
	Template    db.JSONB  `gorm:"column:template;type:jsonb"`
	Config      string    `gorm:"column:config;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy   int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy   int       `gorm:"column:updated_by;type:bigint;not null"`
}

func (action *Action) CleanID() {
	action.ID = 0
}

func (action *Action) InitUID() {
	action.UID = uuid.New()
}

func (action *Action) InitCreatedAt() {
	action.CreatedAt = time.Now().UTC()
}

func (action *Action) InitUpdatedAt() {
	action.UpdatedAt = time.Now().UTC()
}

func (action *Action) AppendNewVersion(newVersion int) {
	action.CleanID()
	action.InitUID()
	action.Version = newVersion
}

func (action *Action) ExportID() int {
	return action.ID
}

func (action *Action) UpdateAppConfig(actionConfig *ActionConfig, userID int) {
	action.Config = actionConfig.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *Action) ExportConfig() *ActionConfig {
	ac := NewActionConfig()
	json.Unmarshal([]byte(action.Config), ac)
	return ac
}

func (action *Action) ExportDisplayName() string {
	return action.Name
}

func (action *Action) ExportTypeInString() string {
	return resourcelist.GetResourceIDMappedType(action.Type)
}

func (action *Action) IsPublic() bool {
	ac := action.ExportConfig()
	return ac.Public
}

func (action *Action) SetPublic(userID int) {
	ac := action.ExportConfig()
	ac.Public = true
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *Action) SetPrivate(userID int) {
	ac := action.ExportConfig()
	ac.Public = false
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}
