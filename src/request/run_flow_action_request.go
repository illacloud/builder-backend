package request

import (
	"encoding/json"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

// The run action HTTP request body like:
// ```json
//
//	{
//	    "resourceID": "ILAfx4p1C7dD",
//	    "actionType": "postgresql",
//	    "displayName": "postgresql1",
//	    "content": {
//	        "mode": "sql",
//	        "query": "select * from users where name like '%jame%';"
//	    },
//	    "context": {
//	        "input1.value": "jame"
//	    }
//	}
//
// ```

const (
	RUN_FLOW_ACTION_REQUEST_FIELD_CONTEXT = "context"
)

type RunFlowActionRequest struct {
	ResourceID     string                 `json:"resourceID,omitempty"`
	FlowActionType string                 `json:"flowActionType" validate:"required"`
	DisplayName    string                 `json:"displayName" validate:"required"`
	Content        map[string]interface{} `json:"content" validate:"required"`
	Context        map[string]interface{} `json:"context" validate:"required"` // for action content raw param
}

func NewRunFlowActionRequest() *RunFlowActionRequest {
	return &RunFlowActionRequest{}
}

func (req *RunFlowActionRequest) ExportFlowActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.FlowActionType)
}

func (req *RunFlowActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *RunFlowActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *RunFlowActionRequest) ExportTemplateWithContextInString() string {
	content := req.Content
	content[RUN_FLOW_ACTION_REQUEST_FIELD_CONTEXT] = req.Content
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *RunFlowActionRequest) ExportContextInString() string {
	jsonByte, _ := json.Marshal(req.Context)
	return string(jsonByte)
}

func (req *RunFlowActionRequest) ExportContext() map[string]interface{} {
	return req.Context
}

func (req *RunFlowActionRequest) DoesContextAvaliable() bool {
	return len(req.Context) > 0
}

func (req *RunFlowActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.FlowActionType)
}
