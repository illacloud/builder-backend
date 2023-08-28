package model

import (
	"time"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

type AppSnapshotForExport struct {
	ID            string                       `json:"snapshotID"`
	TeamID        string                       `json:"teamID"`
	AppRefID      string                       `json:"appID"`
	TargetVersion int                          `json:"targetVersion"`
	TriggerMode   int                          `json:"snapshotTriggerMode"`
	ModifyHistory []*AppModifyHistoryForExport `json:"modifyHistory"`
	CreatedAt     time.Time                    `json:"createdAt"`
}

func NewAppSnapshotForExport(appSnapshot *AppSnapshot, usersLT map[int]*User) *AppSnapshotForExport {
	// construct modify history for export
	modifyHistorys := appSnapshot.ExportModifyHistory()
	modifyHisotrysForExport := make([]*AppModifyHistoryForExport, 0)
	for _, modifyHisotry := range modifyHistorys {
		modifyHisotryForExport := NewAppModifyHistoryForExport(modifyHisotry, usersLT)
		if modifyHisotryForExport != nil {
			modifyHisotrysForExport = append(modifyHisotrysForExport, modifyHisotryForExport)
		}
	}
	return &AppSnapshotForExport{
		ID:            idconvertor.ConvertIntToString(appSnapshot.ID),
		TeamID:        idconvertor.ConvertIntToString(appSnapshot.TeamID),
		AppRefID:      idconvertor.ConvertIntToString(appSnapshot.AppRefID),
		TargetVersion: appSnapshot.TargetVersion,
		TriggerMode:   appSnapshot.TriggerMode,
		ModifyHistory: modifyHisotrysForExport,
		CreatedAt:     appSnapshot.CreatedAt,
	}
}
