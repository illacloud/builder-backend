package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const APP_MODIFY_HISTORY_MAX_LEN = 10

const APP_SNAPSHOT_PERIOD = time.Second * 300 // 5 min

const (
	SNAPSHOT_TRIGGER_MODE_AUTO   = 1
	SNAPSHOT_TRIGGER_MODE_MANUAL = 2
)

type AppSnapshot struct {
	ID            int       `json:"id" 						gorm:"column:id;type:bigserial;primary_key;unique"`
	UID           uuid.UUID `json:"uid"   		   			gorm:"column:uid;type:uuid;not null"`
	TeamID        int       `json:"teamID" 		   			gorm:"column:team_id;type:bigserial"`
	AppRefID      int       `json:"appID" 		    		gorm:"column:app_ref_id;type:bigserial"`
	TargetVersion int       `json:"targetVersion" 			gorm:"column:target_version;type:bigserial"`
	TriggerMode   int       `json:"snapshotTriggerMode"     gorm:"column:trigger_mode;type:smallint"`
	ModifyHistory string    `json:"modifyHistory" 			gorm:"column:modify_history;type:jsonb"`
	CreatedAt     time.Time `json:"createdAt" 				gorm:"column:created_at;type:timestamp"`
}

func NewAppSnapshot(teamID int, appID int, targetVersion int, triggerMode int) *AppSnapshot {
	appSnapshot := &AppSnapshot{
		TeamID:        teamID,
		AppRefID:      appID,
		TargetVersion: targetVersion,
		TriggerMode:   triggerMode,
	}
	appSnapshot.InitUID()
	appSnapshot.InitCreatedAt()
	appSnapshot.InitModifyHistory()
	return appSnapshot
}

func (appSnapshot *AppSnapshot) InitUID() {
	appSnapshot.UID = uuid.New()
}

func (appSnapshot *AppSnapshot) InitCreatedAt() {
	appSnapshot.CreatedAt = time.Now().UTC()
}

func (appSnapshot *AppSnapshot) InitModifyHistory() {
	emptyModifyHistory := make([]interface{}, 0)
	encodingByte, _ := json.Marshal(emptyModifyHistory)
	appSnapshot.ModifyHistory = string(encodingByte)
}

func (appSnapshot *AppSnapshot) SetTargetVersion(targetVersion int) {
	appSnapshot.TargetVersion = targetVersion
}

func (appSnapshot *AppSnapshot) ExportCreatedAt() time.Time {
	return appSnapshot.CreatedAt
}

func (appSnapshot *AppSnapshot) ExportModifyHistory() []*AppModifyHistory {
	appModifyHistorys := make([]*AppModifyHistory, 0)
	json.Unmarshal([]byte(appSnapshot.ModifyHistory), &appModifyHistorys)
	return appModifyHistorys
}

func (appSnapshot *AppSnapshot) ExportTargetVersion() int {
	return appSnapshot.TargetVersion
}

func (appSnapshot *AppSnapshot) SetTriggerMode(triggerMode int) {
	appSnapshot.TriggerMode = triggerMode
}

func (appSnapshot *AppSnapshot) SetTriggerModeAuto() {
	appSnapshot.TriggerMode = SNAPSHOT_TRIGGER_MODE_AUTO
}

func (appSnapshot *AppSnapshot) SetTriggerModeManual() {
	appSnapshot.TriggerMode = SNAPSHOT_TRIGGER_MODE_MANUAL
}

func (appSnapshot *AppSnapshot) ImportModifyHistory(appModifyHistorys []*AppModifyHistory) {
	payload, _ := json.Marshal(appModifyHistorys)
	appSnapshot.ModifyHistory = string(payload)
}

func (appSnapshot *AppSnapshot) DoesActiveSnapshotNeedArchive() bool {
	return time.Now().UTC().After(appSnapshot.CreatedAt.Add(APP_SNAPSHOT_PERIOD))
}

func (appSnapshot *AppSnapshot) PushModifyHistory(currentAppModifyHistory *AppModifyHistory) {
	appModifyHistoryList := appSnapshot.ExportModifyHistory()

	// insert
	appModifyHistoryList = append([]*AppModifyHistory{currentAppModifyHistory}, appModifyHistoryList...)

	// check length
	if len(appModifyHistoryList) > APP_MODIFY_HISTORY_MAX_LEN {
		appModifyHistoryList = appModifyHistoryList[:len(appModifyHistoryList)-1]
	}

	// ok, set it
	appSnapshot.ImportModifyHistory(appModifyHistoryList)
}

func ExtractAllModifierIDFromAppSnapshot(appSnapshots []*AppSnapshot) []int {
	allUserIDsHashT := make(map[int]int, 0)
	userIDs := make([]int, 0)
	for _, appSnapshot := range appSnapshots {
		modifyHistorys := appSnapshot.ExportModifyHistory()
		for _, modifyHistory := range modifyHistorys {
			modifiedBy := modifyHistory.ExportModifiedBy()
			allUserIDsHashT[modifiedBy] = modifiedBy
		}
	}
	for _, userID := range allUserIDsHashT {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
