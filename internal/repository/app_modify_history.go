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

	"github.com/illacloud/builder-backend/internal/util/builderoperation"
)

const SNAPSHOT_TARGET_APP = "app"

type AppModifyHistory struct {
	Operation           int       `json:"operation"  	        gorm:"column:operation;type:smallint"`              // same as websocket protol signal
	OperationTarget     int       `json:"operationTarget"       gorm:"column:operation_target;type:smallint"`     // same as websocket protol target
	OperationTargetName string    `json:"operationTargetName"   gorm:"column:operation_target_name;type:varchar"` // smae as app name or components display name
	ModifiedBy          int       `json:"modifiedBy" 		    gorm:"column:modified_by;type:timestamp"`
	ModifiedAt          time.Time `json:"modifiedAt" 		    gorm:"column:modified_at;type:timestamp"`
}

func (appModifyHistory *AppModifyHistory) ExportModifiedBy() int {
	return appModifyHistory.ModifiedBy
}

func NewAppModifyHistory(operation int, target int, name string, modifyBy int) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:           operation,
		OperationTarget:     target,
		OperationTargetName: name,
		ModifiedBy:          modifyBy,
	}
	appModifyHistory.InitModifiedAt()
	return appModifyHistory
}

func NewTakeAppSnapshotModifyHistory(modifyBy int) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:           builderoperation.SIGNAL_TAKE_APP_SNAPSHOT,
		OperationTarget:     builderoperation.TARGET_APPS,
		OperationTargetName: SNAPSHOT_TARGET_APP,
		ModifiedBy:          modifyBy,
	}
	appModifyHistory.InitModifiedAt()
	return appModifyHistory
}

func NewRecoverAppSnapshotModifyHistory(modifyBy int) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:           builderoperation.SIGNAL_RECOVER_APP_SNAPSHOT,
		OperationTarget:     builderoperation.TARGET_APPS,
		OperationTargetName: SNAPSHOT_TARGET_APP,
		ModifiedBy:          modifyBy,
	}
	appModifyHistory.InitModifiedAt()
	return appModifyHistory
}

func (appModifyHistory *AppModifyHistory) InitModifiedAt() {
	appModifyHistory.ModifiedAt = time.Now().UTC()
}

func (a *AppModifyHistory) ExportToJSONString() string {
	r, _ := json.Marshal(a)
	return string(r)
}
