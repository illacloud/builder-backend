package repository

import (
	"time"

	"github.com/illacloud/builder-backend/internal/idconvertor"
)

type GetSnapshotListResponse struct {
	SnapshotList []*AppSnapshotForExport `json:"snapshotList"`
	TotalPages   int                     `json:"totalPages"`
}

type AppSnapshotForExport struct {
	ID            string                       `json:"snapshotID"`
	TeamID        string                       `json:"teamID"`
	AppRefID      string                       `json:"appRefID"`
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

func NewGetSnapshotListResponse(appSnapshots []*AppSnapshot, usersLT map[int]*User) *GetSnapshotListResponse {
	resp := &GetSnapshotListResponse{}
	resp.SnapshotList = make([]*AppSnapshotForExport, 0)
	for _, appSnapshot := range appSnapshots {
		appSnapshotForExport := NewAppSnapshotForExport(appSnapshot, usersLT)
		resp.SnapshotList = append(resp.SnapshotList, appSnapshotForExport)
	}
	return resp
}

func (resp *GetSnapshotListResponse) ExportForFeedback() interface{} {
	return resp
}
