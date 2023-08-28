package model

import (
	"encoding/json"
	"time"

	"github.com/illacloud/builder-backend/src/utils/builderoperation"
)

const SNAPSHOT_TARGET_APP = "app"

type AppModifyHistory struct {
	Operation                 int         `json:"operation"  	            gorm:"column:operation;type:smallint"`              // same as websocket protol signal
	OperationTarget           int         `json:"operationTarget"           gorm:"column:operation_target;type:smallint"`     // same as websocket protol target
	OperationTargetName       string      `json:"operationTargetName"       gorm:"column:operation_target_name;type:varchar"` // smae as app name or components display name
	OperationBroadcastType    string      `json:"operationBroadcastType"    gorm:"column:operation_broadcast_type;type:varchar"`
	OperationBroadcastPayload interface{} `json:"operationBroadcastPayload" gorm:"column:operation_broadcast_payload;type:varchar"`
	OperationTargetModifiedAt time.Time   `json:"operationTargetModifiedAt" gorm:"column:operation_target_modified_at;type:timestamp"`
	ModifiedBy                int         `json:"modifiedBy" 		        gorm:"column:modified_by;type:timestamp"`
	ModifiedAt                time.Time   `json:"modifiedAt" 		        gorm:"column:modified_at;type:timestamp"`
}

func (appModifyHistory *AppModifyHistory) ExportModifiedBy() int {
	return appModifyHistory.ModifiedBy
}

func NewAppModifyHistory(operation int, target int, name string, broadcastType string, broadcastpayload interface{}, modifyBy int) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:                 operation,
		OperationTarget:           target,
		OperationTargetName:       name,
		OperationBroadcastType:    broadcastType,
		OperationBroadcastPayload: broadcastpayload,
		ModifiedBy:                modifyBy,
	}
	appModifyHistory.InitModifiedAt()
	return appModifyHistory
}

func NewTakeAppSnapshotModifyHistory(modifyBy int) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:                 builderoperation.SIGNAL_TAKE_APP_SNAPSHOT,
		OperationTarget:           builderoperation.TARGET_APPS,
		OperationTargetName:       SNAPSHOT_TARGET_APP,
		OperationBroadcastType:    "",
		OperationBroadcastPayload: nil,
		ModifiedBy:                modifyBy,
	}
	appModifyHistory.InitModifiedAt()
	return appModifyHistory
}

func NewRecoverAppSnapshotModifyHistory(modifyBy int, targetAppSnapshot *AppSnapshot) *AppModifyHistory {
	appModifyHistory := &AppModifyHistory{
		Operation:                 builderoperation.SIGNAL_RECOVER_APP_SNAPSHOT,
		OperationTarget:           builderoperation.TARGET_APPS,
		OperationTargetName:       SNAPSHOT_TARGET_APP,
		OperationBroadcastType:    "",
		OperationBroadcastPayload: nil,
		OperationTargetModifiedAt: targetAppSnapshot.ExportCreatedAt(),
		ModifiedBy:                modifyBy,
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
