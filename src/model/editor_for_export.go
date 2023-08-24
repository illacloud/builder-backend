package repository

type EditorForExport struct {
	AppInfo               *AppForExport          `json:"appInfo"`
	Actions               []*ActionForExport     `json:"actions"`
	Components            *ComponentNode         `json:"components"`
	DependenciesState     map[string][]string    `json:"dependenciesState"`
	DragShadowState       map[string]interface{} `json:"dragShadowState"`
	DottedLineSquareState map[string]interface{} `json:"dottedLineSquareState"`
	DisplayNameState      []string               `json:"displayNameState"`
}

func (resp *EditorForExport) ExportForFeedback() interface{} {
	return resp
}

func NewEditorForExport(appInfo *AppForExport, actions []*ActionForExport, components *ComponentNode, dependenciesState map[string][]string, dragShadowState map[string]interface{}, dottedLineSquareState map[string]interface{}, displayNameState []string) *EditorForExport {
	return &EditorForExport{
		AppInfo:               appInfo,
		Actions:               actions,
		Components:            components,
		DependenciesState:     dependenciesState,
		DragShadowState:       dragShadowState,
		DottedLineSquareState: dottedLineSquareState,
		DisplayNameState:      displayNameState,
	}
}
