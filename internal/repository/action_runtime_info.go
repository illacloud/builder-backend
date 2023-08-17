package repository

const (
	ACTION_RUNTIME_INFO_FIELD_TEAM_ID          = "teamID"
	ACTION_RUNTIME_INFO_FIELD_RESOURCE_ID      = "resourceID"
	ACTION_RUNTIME_INFO_FIELD_ACTION_ID        = "actionID"
	ACTION_RUNTIME_INFO_FIELD_AUTHORIZATION    = "authorization"
	ACTION_RUNTIME_INFO_FIELD_RUN_BY_ANONYMOUS = "runByAnonymous"
)

type ActionRuntimeInfo struct {
	TeamID         string `json:"teamID"`
	ResourceID     string `json:"resourceID"`
	ActionID       string `json:"actionID"`
	Authorization  string `json:"authorization"`
	RunByAnonymous bool   `json:"runByAnonymous"`
}

func NewActionRuntimeInfo(teamID string, resourceID string, actionID string, authorization string) *ActionRuntimeInfo {
	return &ActionRuntimeInfo{
		TeamID:         teamID,
		ResourceID:     resourceID,
		ActionID:       actionID,
		Authorization:  authorization,
		RunByAnonymous: (authorization == ""),
	}
}

func (i *ActionRuntimeInfo) AppendToActionTemplate(tmpl map[string]interface{}) map[string]interface{} {
	tmpl[ACTION_RUNTIME_INFO_FIELD_TEAM_ID] = i.TeamID
	tmpl[ACTION_RUNTIME_INFO_FIELD_RESOURCE_ID] = i.ResourceID
	tmpl[ACTION_RUNTIME_INFO_FIELD_ACTION_ID] = i.ActionID
	tmpl[ACTION_RUNTIME_INFO_FIELD_AUTHORIZATION] = i.Authorization
	tmpl[ACTION_RUNTIME_INFO_FIELD_RUN_BY_ANONYMOUS] = i.RunByAnonymous
	return tmpl
}
