package response

import "github.com/illacloud/builder-backend/src/model"

func GenerateGetAllAppsResponse(allApps []*model.App, usersLT map[int]*model.User) []*model.AppForExport {
	appDtoForExportSlice := make([]*model.AppForExport, 0, len(allApps))
	for _, app := range allApps {
		appForExport := model.NewAppForExport(app, usersLT)
		appDtoForExportSlice = append(appDtoForExportSlice, appForExport)

	}
	return appDtoForExportSlice
}
