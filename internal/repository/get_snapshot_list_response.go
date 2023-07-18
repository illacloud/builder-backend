package repository

import (
	"time"

	"github.com/illacloud/builder-backend/internal/idconvertor"
)

type GetSnapshotListResponse struct {
	SnapshotList []*AppSnapshotForExport `json:"snapshotList"`
}

type AppSnapshotForExport struct {
	ID            string              `json:"snapshotID"`
	TeamID        string              `json:"teamID"`
	AppRefID      string              `json:"appRefID"`
	TargetVersion int                 `json:"targetVersion"`
	TriggerMode   int                 `json:"snapshotTriggerMode"`
	ModifyHistory []*AppModifyHistory `json:"modifyHistory"`
	CreatedAt     time.Time           `json:"createdAt"`
}

func NewAppSnapshotForExport(appSnapshot *AppSnapshot) *AppSnapshotForExport {
	return &AppSnapshotForExport{
		ID:            idconvertor.ConvertIntToString(appSnapshot.ID),
		TeamID:        idconvertor.ConvertIntToString(appSnapshot.TeamID),
		AppRefID:      idconvertor.ConvertIntToString(appSnapshot.AppRefID),
		TargetVersion: appSnapshot.TargetVersion,
		TriggerMode:   appSnapshot.TriggerMode,
		ModifyHistory: appSnapshot.ExportModifyHistory(),
		CreatedAt:     appSnapshot.CreatedAt,
	}
}

func NewGetSnapshotListResponse(appSnapshots []*AppSnapshot) *GetSnapshotListResponse {
	resp := &GetSnapshotListResponse{}
	resp.SnapshotList = make([]*AppSnapshotForExport, len(appSnapshots))
	for _, appSnapshot := range appSnapshots {
		appSnapshotForExport := NewAppSnapshotForExport(appSnapshot)
		resp.SnapshotList = append(resp.SnapshotList, appSnapshotForExport)
	}
	return resp
}

func (resp *GetSnapshotListResponse) ExportForFeedback() interface{} {
	return resp
}
