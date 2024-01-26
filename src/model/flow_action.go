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
	FLOW_ACTION_EDIT_VERSION = 0
)

type FlowAction struct {
	ID          int                    `gorm:"column:id;type:bigserial;primary_key"`
	UID         uuid.UUID              `gorm:"column:uid;type:uuid;not null"`
	TeamID      int                    `gorm:"column:team_id;type:bigserial"`
	WorkflowID  int                    `gorm:"column:workflow_id;type:bigint;not null"`
	Version     int                    `gorm:"column:version;type:bigint;not null"`
	ResourceID  int                    `gorm:"column:resource_id;type:bigint;not null"`
	Name        string                 `gorm:"column:name;type:varchar;size:255;not null"`
	Type        int                    `gorm:"column:type;type:smallint;not null"`
	TriggerMode string                 `gorm:"column:trigger_mode;type:varchar;size:16;not null"`
	Transformer string                 `gorm:"column:transformer;type:jsonb"`
	Template    string                 `gorm:"column:template;type:jsonb"`
	RawTemplate string                 `gorm:"-" sql:"-"`
	Context     map[string]interface{} `gorm:"-" sql:"-"`
	Config      string                 `gorm:"column:config;type:jsonb"`
	CreatedAt   time.Time              `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy   int                    `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt   time.Time              `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy   int                    `gorm:"column:updated_by;type:bigint;not null"`
}

func NewFlowAction() *FlowAction {
	return &FlowAction{}
}

func NewFlowAcitonByCreateFlowActionRequest(teamID int, workflowID int, userID int, req *request.CreateFlowActionRequest) (*FlowAction, error) {
	action := &FlowAction{
		TeamID:      teamID,
		WorkflowID:  workflowID,
		Version:     APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceID:  idconvertor.ConvertStringToInt(req.ResourceID),
		Name:        req.DisplayName,
		Type:        resourcelist.GetResourceNameMappedID(req.FlowActionType),
		TriggerMode: req.TriggerMode,
		Transformer: req.ExportTransformerInString(),
		Template:    req.ExportTemplateInString(),
		Config:      req.ExportConfigInString(),
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action, nil
}

func NewFlowAcitonByUpdateFlowActionRequest(teamID int, workflowID int, userID int, req *request.UpdateFlowActionRequest) (*FlowAction, error) {
	action := &FlowAction{
		TeamID:      teamID,
		WorkflowID:  workflowID,
		Version:     APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceID:  idconvertor.ConvertStringToInt(req.ResourceID),
		Name:        req.DisplayName,
		Type:        resourcelist.GetResourceNameMappedID(req.FlowActionType),
		TriggerMode: req.TriggerMode,
		Transformer: req.ExportTransformerInString(),
		Template:    req.ExportTemplateInString(),
		Config:      req.ExportConfigInString(),
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action, nil
}

func NewFlowAcitonByRunFlowActionRequest(teamID int, workflowID int, userID int, req *request.RunFlowActionRequest) *FlowAction {
	action := &FlowAction{
		TeamID:      teamID,
		WorkflowID:  workflowID,
		Version:     APP_EDIT_VERSION, // new action always created in builder edit mode, and it is edit version.
		ResourceID:  idconvertor.ConvertStringToInt(req.ResourceID),
		Name:        req.DisplayName,
		Type:        resourcelist.GetResourceNameMappedID(req.FlowActionType),
		Template:    req.ExportTemplateInString(),
		RawTemplate: req.ExportTemplateWithContextInString(),
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
	return action
}

func (action *FlowAction) CleanID() {
	action.ID = 0
}

func (action *FlowAction) InitUID() {
	action.UID = uuid.New()
}

func (action *FlowAction) InitCreatedAt() {
	action.CreatedAt = time.Now().UTC()
}

func (action *FlowAction) InitUpdatedAt() {
	action.UpdatedAt = time.Now().UTC()
}

func (action *FlowAction) InitForFork(teamID int, workflowID int, version int, userID int) {
	action.TeamID = teamID
	action.WorkflowID = workflowID
	action.Version = version
	action.CreatedBy = userID
	action.UpdatedBy = userID
	action.CleanID()
	action.InitUID()
	action.InitCreatedAt()
	action.InitUpdatedAt()
}

func (action *FlowAction) SetTemplate(tempalte interface{}) {
	templateInJSONByte, _ := json.Marshal(tempalte)
	action.Template = string(templateInJSONByte)
}

func (action *FlowAction) AppendNewVersion(newVersion int) {
	action.CleanID()
	action.InitUID()
	action.Version = newVersion
}

func (action *FlowAction) ExportID() int {
	return action.ID
}

func (action *FlowAction) ExportType() int {
	return action.Type
}

func (action *FlowAction) ExportResourceID() int {
	return action.ResourceID
}

func (action *FlowAction) ExportConfig() *FlowActionConfig {
	ac := NewFlowActionConfig()
	json.Unmarshal([]byte(action.Config), ac)
	return ac
}

func (action *FlowAction) ExportDisplayName() string {
	return action.Name
}

func (action *FlowAction) ExportIcon() string {
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

func (action *FlowAction) ExportTypeInString() string {
	return resourcelist.GetResourceIDMappedType(action.Type)
}

func (action *FlowAction) SetContextByMap(context map[string]interface{}) {
	template := action.ExportTemplateInMap()
	template[ACTION_RUNTIME_INFO_FIELD_CONTEXT] = context
	templateJsonByte, _ := json.Marshal(template)
	action.Template = string(templateJsonByte)
}

func (action *FlowAction) UpdateAppConfig(actionConfig *FlowActionConfig, userID int) {
	action.Config = actionConfig.ExportToJSONString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *FlowAction) MergeRunFlowActionContextToRawTemplate(context map[string]interface{}) {
	template := action.ExportTemplateInMap()
	template[ACTION_RUNTIME_INFO_FIELD_CONTEXT] = context
	templateJsonByte, _ := json.Marshal(template)
	action.RawTemplate = string(templateJsonByte)
}

func (action *FlowAction) UpdateWithRunFlowActionRequest(req *request.RunFlowActionRequest, userID int) {
	action.MergeRunFlowActionContextToRawTemplate(req.ExportContext())
	action.Template = req.ExportTemplateInString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()

	// check if is onboarding action (which have no action storaged in database)
	if len(action.Template) == 0 {
		action.Template = action.RawTemplate
	}
}

func (action *FlowAction) UpdateFlowAcitonByUpdateFlowActionRequest(teamID int, workflowID int, userID int, req *request.UpdateFlowActionRequest) {
	action.TeamID = teamID
	action.WorkflowID = workflowID
	action.Version = APP_EDIT_VERSION // new action always created in builder edit mode, and it is edit version.
	action.ResourceID = idconvertor.ConvertStringToInt(req.ResourceID)
	action.Name = req.DisplayName
	action.Type = resourcelist.GetResourceNameMappedID(req.FlowActionType)
	action.TriggerMode = req.TriggerMode
	action.Transformer = req.ExportTransformerInString()
	action.Template = req.ExportTemplateInString()
	action.Config = req.ExportConfigInString()
	action.UpdatedBy = userID
	action.InitUpdatedAt()
}

func (action *FlowAction) IsVirtualFlowAction() bool {
	return resourcelist.IsVirtualResourceByIntType(action.Type)
}

func (action *FlowAction) IsLocalVirtualFlowAction() bool {
	return resourcelist.IsLocalVirtualResourceByIntType(action.Type)
}

func (action *FlowAction) IsRemoteVirtualFlowAction() bool {
	return resourcelist.IsRemoteVirtualResourceByIntType(action.Type)
}

func (action *FlowAction) ExportTransformerInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.Transformer), &payload)
	return payload
}

func (action *FlowAction) ExportTemplateInMap() map[string]interface{} {
	payload := make(map[string]interface{}, 0)
	json.Unmarshal([]byte(action.Template), &payload)
	// add resourceID, runByAnonymous, teamID field for extend action runtime info
	payload["resourceID"] = action.ResourceID
	payload["runByAnonymous"] = true
	payload["teamID"] = action.TeamID
	return payload
}

func (action *FlowAction) ExportRawTemplateInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.RawTemplate), &payload)
	return payload
}

func (action *FlowAction) ExportConfigInMap() map[string]interface{} {
	var payload map[string]interface{}
	json.Unmarshal([]byte(action.Config), &payload)
	return payload
}

// the action runtime does not pass the env info for virtual resource, so add them.
func (action *FlowAction) AppendRuntimeInfoForVirtualResource(authorization string, teamID int) {
	template := action.ExportTemplateInMap()
	template[ACTION_RUNTIME_INFO_FIELD_TEAM_ID] = teamID // the action.TeamID will invalied when onboarding
	template[ACTION_RUNTIME_INFO_FIELD_APP_ID] = action.WorkflowID
	template[ACTION_RUNTIME_INFO_FIELD_RESOURCE_ID] = action.ResourceID
	template[ACTION_RUNTIME_INFO_FIELD_ACTION_ID] = action.ID
	template[ACTION_RUNTIME_INFO_FIELD_AUTHORIZATION] = authorization
	template[ACTION_RUNTIME_INFO_FIELD_RUN_BY_ANONYMOUS] = (authorization == "")
	templateInByte, _ := json.Marshal(template)
	action.Template = string(templateInByte)
}

func (action *FlowAction) SetResourceIDByAiAgent(aiAgent *illaresourcemanagersdk.AIAgentForExport) {
	action.ResourceID = aiAgent.ExportIDInInt()
}

func DoesFlowActionHasBeenCreated(actionID int) bool {
	return actionID > INVALIED_ACTION_ID
}
