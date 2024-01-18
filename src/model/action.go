package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

const (
	ACTION_RUNTIME_INFO_FIELD_TEAM_ID          = "teamID"
	ACTION_RUNTIME_INFO_FIELD_APP_ID           = "appID"
	ACTION_RUNTIME_INFO_FIELD_RESOURCE_ID      = "resourceID"
	ACTION_RUNTIME_INFO_FIELD_ACTION_ID        = "actionID"
	ACTION_RUNTIME_INFO_FIELD_AUTHORIZATION    = "authorization"
	ACTION_RUNTIME_INFO_FIELD_RUN_BY_ANONYMOUS = "runByAnonymous"
	ACTION_RUNTIME_INFO_FIELD_CONTEXT          = "context"
)

const (
	INVALIED_ACTION_ID = 0
)

type Action struct {
	ID            int       `gorm:"column:id;type:bigserial;primary_key"`
	UID           uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	TeamID        int       `gorm:"column:team_id;type:bigserial"`
	AppRefID      int       `gorm:"column:app_ref_id;type:bigint;not null"`
	Version       int       `gorm:"column:version;type:bigint;not null"`
	ResourceRefID int       `gorm:"column:resource_ref_id;type:bigint;not null"`
	Name          string    `gorm:"column:name;type:varchar;size:255;not null"`
	Type          int       `gorm:"column:type;type:smallint;not null"`
	TriggerMode   string    `gorm:"column:trigger_mode;type:varchar;size:16;not null"`
	Transformer   string    `gorm:"column:transformer;type:jsonb"`
	Template      string    `gorm:"column:template;type:jsonb"`
	RawTemplate   string    `gorm:"-" sql:"-"`
	Config        string    `gorm:"column:config;type:jsonb"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy     int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy     int       `gorm:"column:updated_by;type:bigint;not null"`
}

func NewAction() *Action {
	return &Action{}
}

func NewAcitonByCreateActionRequest(app *App, userID int, req *request.CreateActionRequest) (*Action, error) {
	action := &Action{
		TeamID:        app.ExportTeamID(),
		AppRefID:      app.ExportID(),
		Version:       APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceRefID: idconvertor.ConvertStringToInt(req.ResourceID),
		Name:          req.DisplayName,
		Type:          resourcelist.GetResourceNameMappedID(req.ActionType),
		TriggerMode:   req.TriggerMode,
		Transformer:   req.ExportTransformerInString(),
		Template:      req.ExportTemplateInString(),
		Config:        req.ExportConfigInString(),
		CreatedBy:     userID,
		UpdatedBy:     userID,
	}
	if app.IsPublic() {
		action.SetPublic(userID)
	} else {
		action.SetPrivate(userID)
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action, nil
}

func NewAcitonByUpdateActionRequest(app *App, userID int, req *request.UpdateActionRequest) (*Action, error) {
	action := &Action{
		TeamID:        app.ExportTeamID(),
		AppRefID:      app.ExportID(),
		Version:       APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceRefID: idconvertor.ConvertStringToInt(req.ResourceID),
		Name:          req.DisplayName,
		Type:          resourcelist.GetResourceNameMappedID(req.ActionType),
		TriggerMode:   req.TriggerMode,
		Transformer:   req.ExportTransformerInString(),
		Template:      req.ExportTemplateInString(),
		Config:        req.ExportConfigInString(),
		CreatedBy:     userID,
		UpdatedBy:     userID,
	}
	if app.IsPublic() {
		action.SetPublic(userID)
	} else {
		action.SetPrivate(userID)
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action, nil
}

func NewAcitonByRunActionRequest(app *App, userID int, req *request.RunActionRequest) *Action {
	action := &Action{
		TeamID:        app.ExportTeamID(),
		AppRefID:      app.ExportID(),
		Version:       APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceRefID: idconvertor.ConvertStringToInt(req.ResourceID),
		Name:          req.DisplayName,
		Type:          resourcelist.GetResourceNameMappedID(req.ActionType),
		Template:      req.ExportTemplateInString(),
		RawTemplate:   req.ExportTemplateWithContextInString(),
		CreatedBy:     userID,
		UpdatedBy:     userID,
	}
	if app.IsPublic() {
		action.SetPublic(userID)
	} else {
		action.SetPrivate(userID)
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action
}

func (action *Action) CleanID() {
	action.ID = 0
}

func (action *Action) InitUID() {
	action.UID = uuid.New()
}

func (action *Action) InitCreatedAt() {
	action.CreatedAt = time.Now().UTC()
}

func (action *Action) InitUpdatedAt() {
	action.UpdatedAt = time.Now().UTC()
}

func (action *Action) InitForFork(teamID int, appID int, version int, userID int) {
	action.TeamID = teamID
	action.AppRefID = appID
	action.Version = version
	action.CreatedBy = userID
	action.UpdatedBy = userID
	action.CleanID()
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
}

func (action *Action) AppendNewVersion(newVersion int) {
	action.CleanID()
	action.InitUID()
	action.Version = newVersion
}

func (action *Action) ExportID() int {
	return action.ID
}

func (action *Action) ExportType() int {
	return action.Type
}

func (action *Action) ExportResourceID() int {
	return action.ResourceRefID
}

func (action *Action) ExportConfig() *ActionConfig {
	ac := NewActionConfig()
	json.Unmarshal([]byte(action.Config), ac)
	return ac
}

func (action *Action) ExportDisplayName() string {
	return action.Name
}

func (action *Action) ExportIcon() string {
	content := action.ExportTemplateInMap()
	virtualResource, hitVirtualResource := content["virtualResource"]
	if !hitVirtualResource {
		return ""
	}
	virtualResourceAsserted, virtualResourceAssertPass := virtualResource.(map[string]interface{})
	if !virtualResourceAssertPass {
		return ""
	}
	icon, hitIcon := virtualResourceAsserted["icon"]
	if !hitIcon {
		return ""
	}
	iconAsserted, iconAssertPass := icon.(string)
	if !iconAssertPass {
		return ""
	}
	return iconAsserted
}

func (action *Action) ExportTypeInString() string {
	return resourcelist.GetResourceIDMappedType(action.Type)
}

func (action *Action) IsPublic() bool {
	ac := action.ExportConfig()
	return ac.Public
}

func (action *Action) SetPublic(userID int) {
	ac := action.ExportConfig()
	ac.SetPublic()
	action.Config = ac.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *Action) SetPrivate(userID int) {
	ac := action.ExportConfig()
	ac.SetPrivate()
	action.Config = ac.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *Action) SetTutorialLink(link string, userID int) {
	ac := action.ExportConfig()
	ac.SetTutorialLink(link)
	action.Config = ac.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

// WARRING! this is an view-level method, do not use this method to sync database changes, just for export data.
func (action *Action) RewritePublicSettings(isPublic bool) {
	actionConfig := action.ExportConfig()
	actionConfig.Public = isPublic
	action.Config = actionConfig.ExportToJSONString()
}

func (action *Action) UpdateAppConfig(actionConfig *ActionConfig, userID int) {
	action.Config = actionConfig.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *Action) UpdateWithRunActionRequest(req *request.RunActionRequest, userID int) {
	action.MergeRunActionContextToRawTemplate(req.ExportContext())
	action.Template = req.ExportTemplateInString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()

	// check if is onboarding action (which have no action storaged in database)
	if len(action.Template) == 0 {
		action.Template = action.RawTemplate
	}
}

func (action *Action) UpdateAcitonByUpdateActionRequest(app *App, userID int, req *request.UpdateActionRequest) {
	action.TeamID = app.ExportTeamID()
	action.AppRefID = app.ExportID()
	action.Version = APP_EDIT_VERSION // new action always created in builder edit mode, and it is edit version.
	action.ResourceRefID = idconvertor.ConvertStringToInt(req.ResourceID)
	action.Name = req.DisplayName
	action.Type = resourcelist.GetResourceNameMappedID(req.ActionType)
	action.TriggerMode = req.TriggerMode
	action.Transformer = req.ExportTransformerInString()
	action.Template = req.ExportTemplateInString()
	action.Config = req.ExportConfigInString()
	action.UpdatedBy = userID
	if app.IsPublic() {
		action.SetPublic(userID)
	} else {
		action.SetPrivate(userID)
	}
	action.InitUpdatedAt()
}

func (action *Action) IsVirtualAction() bool {
	return resourcelist.IsVirtualResourceByIntType(action.Type)
}

func (action *Action) IsLocalVirtualAction() bool {
	return resourcelist.IsLocalVirtualResourceByIntType(action.Type)
}

func (action *Action) IsRemoteVirtualAction() bool {
	return resourcelist.IsRemoteVirtualResourceByIntType(action.Type)
}

func (action *Action) ExportTransformerInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.Transformer), &payload)
	return payload
}

func (action *Action) ExportTemplateInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.Template), &payload)
	return payload
}

func (action *Action) ExportRawTemplateInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.RawTemplate), &payload)
	return payload
}

func (action *Action) ExportConfigInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.Config), &payload)
	return payload
}

// the action runtime does not pass the env info for virtual resource, so add them.
func (action *Action) AppendRuntimeInfoForVirtualResource(authorization string, teamID int) {
	template := action.ExportTemplateInMap()
	template[ACTION_RUNTIME_INFO_FIELD_TEAM_ID] = teamID // the action.TeamID will invalied when onboarding
	template[ACTION_RUNTIME_INFO_FIELD_APP_ID] = action.AppRefID
	template[ACTION_RUNTIME_INFO_FIELD_RESOURCE_ID] = action.ResourceRefID
	template[ACTION_RUNTIME_INFO_FIELD_ACTION_ID] = action.ID
	template[ACTION_RUNTIME_INFO_FIELD_AUTHORIZATION] = authorization
	template[ACTION_RUNTIME_INFO_FIELD_RUN_BY_ANONYMOUS] = (authorization == "")
	templateInByte, _ := json.Marshal(template)
	action.Template = string(templateInByte)
}

func (action *Action) MergeRunActionContextToRawTemplate(context map[string]interface{}) {
	template := action.ExportTemplateInMap()
	template[ACTION_RUNTIME_INFO_FIELD_CONTEXT] = context
	templateJsonByte, _ := json.Marshal(template)
	action.RawTemplate = string(templateJsonByte)
}

func (action *Action) SetResourceIDByAiAgent(aiAgent *illaresourcemanagersdk.AIAgentForExport) {
	action.ResourceRefID = aiAgent.ExportIDInInt()
}

func ExportAllActionASActionSummary(actions []*Action) []*ActionSummaryForExport {
	ret := make([]*ActionSummaryForExport, 0)
	for _, action := range actions {
		ret = append(ret, NewActionSummaryForExportByAction(action))
	}
	return ret
}

func DoesActionHasBeenCreated(actionID int) bool {
	return actionID > INVALIED_ACTION_ID
}
