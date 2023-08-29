package model

type AppConfigForExport struct {
	Public                 bool                      `json:"public"` // switch for public app (which can view by anonymous user)
	WaterMark              bool                      `json:"waterMark"`
	Description            string                    `json:"description"`
	PublishedToMarketplace bool                      `json:"publishedToMarketplace"`
	Cover                  string                    `json:"cover"`
	Components             []string                  `json:"components"`
	Actions                []*ActionSummaryForExport `json:"actions"`
}

func NewAppConfigForExport(appConfig *AppConfig, treeStates []*TreeState, actions []*Action) *AppConfigForExport {
	return &AppConfigForExport{
		Public:                 appConfig.Public,
		WaterMark:              appConfig.WaterMark,
		Description:            appConfig.Description,
		PublishedToMarketplace: appConfig.PublishedToMarketplace,
		Cover:                  appConfig.Cover,
		Components:             ExtractComponentsNameList(treeStates),
		Actions:                ExportAllActionASActionSummary(actions),
	}
}

func NewAppConfigForExportWithoutComponentsAndActions(appConfig *AppConfig) *AppConfigForExport {
	return &AppConfigForExport{
		Public:                 appConfig.Public,
		WaterMark:              appConfig.WaterMark,
		Description:            appConfig.Description,
		PublishedToMarketplace: appConfig.PublishedToMarketplace,
		Cover:                  appConfig.Cover,
		Components:             make([]string, 0),
		Actions:                make([]*ActionSummaryForExport, 0),
	}
}
