package repository

import (
	"time"
)

type AppModifyHistoryForExport struct {
	Operation           int                `json:"operation"  	        gorm:"column:operation;type:smallint"`              // same as websocket protol signal
	OperationTarget     int                `json:"operationTarget"       gorm:"column:operation_target;type:smallint"`     // same as websocket protol target
	OperationTargetName string             `json:"operationTargetName"   gorm:"column:operation_target_name;type:varchar"` // smae as app name or components display name
	ModifiedBy          *UserForModifiedBy `json:"modifiedBy" 		    gorm:"column:modified_by;type:timestamp"`
	ModifiedAt          time.Time          `json:"modifiedAt" 		    gorm:"column:modified_at;type:timestamp"`
}

func NewAppModifyHistoryForExport(appModifyHistory *AppModifyHistory, usersLT map[int]*User) *AppModifyHistoryForExport {
	targetUser, hit := usersLT[appModifyHistory.ModifiedBy]
	if !hit {
		return nil
	}
	return &AppModifyHistoryForExport{
		Operation:           appModifyHistory.Operation,
		OperationTarget:     appModifyHistory.OperationTarget,
		OperationTargetName: appModifyHistory.OperationTargetName,
		ModifiedBy:          NewUserForModifiedBy(targetUser),
		ModifiedAt:          appModifyHistory.ModifiedAt,
	}
}
