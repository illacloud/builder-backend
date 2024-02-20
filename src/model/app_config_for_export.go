package model

type AppConfigForExport struct {
	Public                 bool                      `json:"public"` // switch for public app (which can view by anonymous user)
	WaterMark              bool                      `json:"waterMark"`
	Description            string                    `json:"description"`
	PublishedToMarketplace bool                      `json:"publishedToMarketplace"`
	PublishWithAIAgent     bool                      `json:"publishWithAIAgent"`
	Cover                  string                    `json:"cover"`
	AppType                string                    `json:"appType"`
	Components             []string                  `json:"components"`
	Actions                []*ActionSummaryForExport `json:"actions"`
}

func NewAppConfigForExport(appConfig *AppConfig, treeStates []*TreeState, actions []*Action) *AppConfigForExport {
	return &AppConfigForExport{
		Public:                 appConfig.Public,
		WaterMark:              appConfig.WaterMark,
		Description:            appConfig.Description,
		PublishedToMarketplace: appConfig.PublishedToMarketplace,
		PublishWithAIAgent:     appConfig.PublishWithAIAgent,
		Cover:                  appConfig.Cover,
		AppType:                appConfig.ExportAppTypeToString(),
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
		PublishWithAIAgent:     appConfig.PublishWithAIAgent,
		Cover:                  appConfig.Cover,
		AppType:                appConfig.ExportAppTypeToString(),
		Components:             make([]string, 0),
		Actions:                make([]*ActionSummaryForExport, 0),
	}
}
