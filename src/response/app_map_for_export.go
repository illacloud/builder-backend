package response

import (
	"github.com/illacloud/builder-backend/src/model"
)

type AppMapForExport struct {
	AppList map[int]*model.AppForExport `json:"appList"`
}

func NewAppMapForExport(appList []*model.App, userLT map[int]*model.User) *AppMapForExport {
	// build AppList
	appMapForExport := &AppMapForExport{}
	appMapForExport.AppList = make(map[int]*model.AppForExport, len(appList))
	for _, app := range appList {
		appForExport := model.NewAppForExport(app, userLT)
		appMapForExport.AppList[app.ID] = appForExport
	}
	return appMapForExport
}

func (resp *AppMapForExport) ExportForFeedback() interface{} {
	return resp
}
