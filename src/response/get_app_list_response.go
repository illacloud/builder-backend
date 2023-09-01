package response

import (
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/storage"
)

type AppListResponse struct {
	AppList       []*model.AppForExport `json:"appList"`
	TotalPages    int                   `json:"-"`
	TotalAppCount int64                 `json:"-"`
	HasMore       bool                  `json:"hasMore"`
}

func NewAppListResponse(appList []*model.App, pagination *storage.Pagination, userLT map[int]*model.User) (*AppListResponse, error) {
	// build AppList
	appListForExport := &AppListResponse{}
	appListForExport.AppList = make([]*model.AppForExport, 0)
	for _, app := range appList {
		appForExport := model.NewAppForExport(app, userLT)
		appListForExport.AppList = append(appListForExport.AppList, appForExport)
	}
	//build page and cotun
	appListForExport.TotalPages = pagination.TotalPages
	appListForExport.TotalAppCount = pagination.TotalRows
	// check hasMore
	if pagination.Page < pagination.TotalPages {
		appListForExport.HasMore = true
	}
	return appListForExport, nil
}

func (resp *AppListResponse) ExportForFeedback() interface{} {
	return resp
}
