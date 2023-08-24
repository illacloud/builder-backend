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
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/pkg/db"
)

type Resource struct {
	ID        int       `gorm:"column:id;type:bigserial;primary_key"`
	UID       uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `gorm:"column:team_id;type:bigserial"`
	Name      string    `gorm:"column:name;type:varchar;size:200;not null"`
	Type      int       `gorm:"column:type;type:smallint;not null"`
	Options   db.JSONB  `gorm:"column:options;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy int       `gorm:"column:updated_by;type:bigint;not null"`
}

func (resource *Resource) ExportUpdatedAt() time.Time {
	return resource.UpdatedAt
}
