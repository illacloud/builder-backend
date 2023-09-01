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
//	    "resourceID": "ILAfx4p1C7cc",
//	    "actionType": "postgresql",
//	    "displayName": "postgresql1",
//	    "content": {
//	        "mode": "sql",
//	        "query": "select * from data;"
//	    }
//	}
//
// ```
type RunActionRequest struct {
	ResourceID  string                 `json:"resourceID,omitempty"`
	ActionType  string                 `json:"actionType"         validate:"required"`
	DisplayName string                 `json:"displayName"        validate:"required"`
	Content     map[string]interface{} `json:"content"            validate:"required"`
}

func NewRunActionRequest() *RunActionRequest {
	return &RunActionRequest{}
}

func (req *RunActionRequest) ExportActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.ActionType)
}

func (req *RunActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *RunActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *RunActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.ActionType)
}
