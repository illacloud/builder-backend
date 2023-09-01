package response

import (
	"github.com/illacloud/builder-backend/src/model"
)

type GetSnapshotListResponse struct {
	SnapshotList []*model.AppSnapshotForExport `json:"snapshotList"`
	TotalPages   int                           `json:"totalPages"`
}

func NewGetSnapshotListResponse(appSnapshots []*model.AppSnapshot, totalPages int, usersLT map[int]*model.User) *GetSnapshotListResponse {
	resp := &GetSnapshotListResponse{
		TotalPages: totalPages,
	}
	resp.SnapshotList = make([]*model.AppSnapshotForExport, 0)
	for _, appSnapshot := range appSnapshots {
		appSnapshotForExport := model.NewAppSnapshotForExport(appSnapshot, usersLT)
		resp.SnapshotList = append(resp.SnapshotList, appSnapshotForExport)
	}
	return resp
}

func (resp *GetSnapshotListResponse) ExportForFeedback() interface{} {
	return resp
}
