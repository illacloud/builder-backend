package repository

func GenerateGetAllAppsResponse(allApps []*App, usersLT map[int]*User) []*AppForExport {
	appDtoForExportSlice := make([]*AppForExport, 0, len(allApps))
	for _, app := range allApps {
		appForExport := NewAppForExport(app, usersLT)
		appDtoForExportSlice = append(appDtoForExportSlice, appForExport)

	}
	return appDtoForExportSlice
}
