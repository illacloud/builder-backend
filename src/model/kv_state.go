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
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	KV_STATE_FIELD_DISPLAY_NAME = "displayName"
)

type KVState struct {
	ID        int       `json:"id" 		   gorm:"column:id;type:bigserial"`
	UID       uuid.UUID `json:"uid" 	   gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `json:"teamID"    gorm:"column:team_id;type:bigserial"`
	StateType int       `json:"state_type" gorm:"column:state_type;type:bigint"`
	AppRefID  int       `json:"app_ref_id" gorm:"column:app_ref_id;type:bigint"`
	Version   int       `json:"version"    gorm:"column:version;type:bigint"`
	Key       string    `json:"key" 	   gorm:"column:key;type:text"`
	Value     string    `json:"value" 	   gorm:"column:value;type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamp"`
	CreatedBy int       `json:"created_by" gorm:"column:created_by;type:bigint"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	UpdatedBy int       `json:"updated_by" gorm:"column:updated_by;type:bigint"`
}

func (kvState *KVState) CleanID() {
	kvState.ID = 0
}

func (kvState *KVState) InitUID() {
	kvState.UID = uuid.New()
}

func (kvState *KVState) InitCreatedAt() {
	kvState.CreatedAt = time.Now().UTC()
}

func (kvState *KVState) InitUpdatedAt() {
	kvState.UpdatedAt = time.Now().UTC()
}

func (kvState *KVState) InitForFork(teamID int, appID int, version int, userID int) {
	kvState.TeamID = teamID
	kvState.AppRefID = appID
	kvState.Version = version
	kvState.CreatedBy = userID
	kvState.UpdatedBy = userID
	kvState.CleanID()
	kvState.InitUID()
	kvState.InitCreatedAt()
	kvState.InitUpdatedAt()
}

func NewKVStateByApp(app *App, stateType int) *KVState {
	kvState := &KVState{
		TeamID:    app.ExportTeamID(),
		StateType: stateType,
		AppRefID:  app.ExportID(),
		Version:   APP_EDIT_VERSION,
		CreatedBy: app.ExportUpdatedBy(),
		UpdatedBy: app.ExportUpdatedBy(),
	}
	kvState.InitUID()
	kvState.InitCreatedAt()
	kvState.InitUpdatedAt()
	return kvState
}

func NewKVStateByWebsocketMessage(app *App, stateType int, data interface{}) (*KVState, error) {
	kvState := NewKVStateByApp(app, stateType)
	udata, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("KVStateDto ConstructByMap failed, please check your input.")
	}
	displayName, mapok := udata[KV_STATE_FIELD_DISPLAY_NAME].(string)
	if !mapok {
		return nil, errors.New("KVStateDto ConstructByMap failed, can not find displayName field.")
	}
	// fild
	kvState.Key = displayName
	jsonbyte, _ := json.Marshal(udata)
	kvState.Value = string(jsonbyte)
	return kvState, nil
}

func NewKVStateByWebsocketMessageWithGivenKey(app *App, stateType int, key string, data interface{}) (*KVState, error) {
	kvState := NewKVStateByApp(app, stateType)
	// fild
	kvState.Key = key
	jsonbyte, _ := json.Marshal(data)
	kvState.Value = string(jsonbyte)
	return kvState, nil
}

func (kvState *KVState) UpdateByNewKVState(newKVState *KVState) {
	kvState.Key = newKVState.Key
	kvState.Value = newKVState.Value
	kvState.UpdatedBy = newKVState.UpdatedBy
	kvState.InitUpdatedAt()
}

func (kvState *KVState) AppendNewVersion(newVersion int) {
	kvState.CleanID()
	kvState.InitUID()
	kvState.Version = newVersion
}

func (kvState *KVState) ExportID() int {
	return kvState.ID
}

func (kvState *KVState) ExportKey() string {
	return kvState.Key
}

func (kvState *KVState) ExportVersion() int {
	return kvState.Version
}
