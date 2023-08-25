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

package model

import (
	"time"

	"github.com/google/uuid"
)

type SetState struct {
	ID        int       `json:"id" 		   gorm:"column:id;type:bigserial"`
	UID       uuid.UUID `json:"uid" 	   gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `json:"teamID"    gorm:"column:team_id;type:bigserial"`
	StateType int       `json:"state_type" gorm:"column:state_type;type:bigint"`
	AppRefID  int       `json:"app_ref_id" gorm:"column:app_ref_id;type:bigint"`
	Version   int       `json:"version"    gorm:"column:version;type:bigint"`
	Value     string    `json:"value" 	   gorm:"column:value;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamp"`
	CreatedBy int       `json:"created_by" gorm:"column:created_by;type:bigint"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	UpdatedBy int       `json:"updated_by" gorm:"column:updated_by;type:bigint"`
}

func (setState *SetState) CleanID() {
	setState.ID = 0
}

func (setState *SetState) InitUID() {
	setState.UID = uuid.New()
}

func (setState *SetState) InitCreatedAt() {
	setState.CreatedAt = time.Now().UTC()
}

func (setState *SetState) InitUpdatedAt() {
	setState.UpdatedAt = time.Now().UTC()
}

func (setState *SetState) InitForFork(teamID int, appID int, version int, userID int) {
	setState.TeamID = teamID
	setState.AppRefID = appID
	setState.Version = version
	setState.CreatedBy = userID
	setState.UpdatedBy = userID
	setState.CleanID()
	setState.InitUID()
	setState.InitCreatedAt()
	setState.InitUpdatedAt()
}

func (setState *SetState) AppendNewVersion(newVersion int) {
	setState.CleanID()
	setState.InitUID()
	setState.Version = newVersion
}

func (setState *SetState) ExportID() int {
	return setState.ID
}