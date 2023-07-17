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
)

const APP_MODIFY_HISTORY_MAX_LEN = 10

const (
	SNAPSHOT_TRIGGER_MODE_AUTO   = 1
	SNAPSHOT_TRIGGER_MODE_MANUAL = 2
)

type AppSnapshot struct {
	ID            int       `json:"id" 				gorm:"column:id;type:bigserial;primary_key;unique"`
	UID           uuid.UUID `json:"uid"   		   	gorm:"column:uid;type:uuid;not null"`
	TeamID        int       `json:"teamID" 		   	gorm:"column:team_id;type:bigserial"`
	AppRefID      int       `json:"appID" 		gorm:"column:app_ref_id;type:bigserial"`
	TargetVersion int       `json:"targetVersion" 	gorm:"column:target_version;type:bigserial"`
	TriggerMode   int       `json:"triggerMode"     gorm:"column:trigger_mode;type:smallint"`
	ModifyHistory string    `json:"modifyHistory" 	gorm:"column:modify_history;type:jsonb"`
	CreatedAt     time.Time `json:"createdAt" 		gorm:"column:created_at;type:timestamp"`
}

func NewAppSnapshot(teamID int, appID int, targetVersion int, triggerMode int) *AppSnapshot {
	appSnapshot := &AppSnapshot{
		TeamID:        teamID,
		AppRefID:      appName,
		TargetVersion: targetVersion,
		TriggerMode:   triggerMode,
	}
	appSnapshot.InitUID()
	appSnapshot.InitCreatedAt()
	appSnapshot.InitModifyHistory()
	return app
}

func (appSnapshot *AppSnapshot) InitUID() {
	appSnapshot.UID = uuid.New()
}

func (appSnapshot *AppSnapshot) InitCreatedAt() {
	appSnapshot.CreatedAt = time.Now().UTC()
}

func (appSnapshot *AppSnapshot) InitModifyHistory() {
	enptyModifyHistory := make([]interface{}, 0)
	appSnapshot.ModifyHistory, _ = json.Marshal(enptyModifyHistory)
}

func (appSnapshot *AppSnapshot) SetTargetVersion(targetVersion int) {
	appSnapshot.TargetVersion = targetVersion
}

func (appSnapshot *AppSnapshot) ExportModifyHistory() []*AppModifyHistory {
	appModifyHistorys := make([]*AppModifyHistory, 0)
	json.Unmarshal([]byte(app.ModifyHistory), &appModifyHistorys)
	return appModifyHistorys
}

func (appSnapshot *AppSnapshot) ExportTargetVersion() int {
	return appSnapshot.TargetVersion
}

func (appSnapshot *AppSnapshot) ImportModifyHistory(appModifyHistorys []*AppModifyHistory) {
	payload, _ := json.Marshal(appModifyHistorys)
	app.ModifyHistory = string(payload)
}

func (app *App) PushModifyHistory(currentAppModifyHistory *AppModifyHistory) {
	appModifyHistoryList := app.ExportModifyHistory()

	// insert
	appModifyHistoryList = append([]*AppModifyHistory{currentAppModifyHistory}, appModifyHistoryList...)

	// check length
	if len(appModifyHistoryList) > APP_MODIFY_HISTORY_MAX_LEN {
		appModifyHistoryList = appModifyHistoryList[:len(appModifyHistoryList)-1]
	}

	// ok, set it
	app.ImportModifyHistory(appModifyHistoryList)
}
