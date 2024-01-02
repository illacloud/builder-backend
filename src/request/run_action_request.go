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
	RUN_ACTION_REQUEST_FIELD_CONTEXT = "context"
)

type RunActionRequest struct {
	ResourceID  string                 `json:"resourceID,omitempty"`
	ActionType  string                 `json:"actionType"         validate:"required"`
	DisplayName string                 `json:"displayName"        validate:"required"`
	Content     map[string]interface{} `json:"content"            validate:"required"`
	Context     map[string]interface{} `json:"context"            validate:"required"` // for action content raw param
}

func NewRunActionRequest() *RunActionRequest {
	return &RunActionRequest{}
}

func (req *RunActionRequest) ExportFlowActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.ActionType)
}

func (req *RunActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *RunActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *RunActionRequest) ExportTemplateWithContextInString() string {
	content := req.Content
	content[RUN_ACTION_REQUEST_FIELD_CONTEXT] = req.Content
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *RunActionRequest) ExportContextInString() string {
	jsonByte, _ := json.Marshal(req.Context)
	return string(jsonByte)
}

func (req *RunActionRequest) ExportContext() map[string]interface{} {
	return req.Context
}

func (req *RunActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.ActionType)
}
