package request

import (
	"encoding/json"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

// The create action HTTP request body like:
// ```json
//
//	{
//	    "actionType": "postgresql",
//	    "displayName": "postgresql1",
//	    "resourceID": "ILAfx4p1C7cd",
//	    "content": {
//	        "mode": "sql",
//	        "query": ""
//	    },
//	    "isVirtualResource": true,
//	    "transformer": {
//	        "rawData": "",
//	        "enable": false
//	    },
//	    "triggerMode": "manually",
//	    "config": {
//	        "public": false,
//	        "advancedConfig": {
//	            "runtime": "none",
//	            "pages": [],
//	            "delayWhenLoaded": "",
//	            "displayLoadingPage": false,
//	            "isPeriodically": false,
//	            "periodInterval": ""
//	        }
//	    }
//	}
//
// ```
type CreateFlowActionRequest struct {
	ActionType        string                 `json:"actionType"         validate:"required"`
	DisplayName       string                 `json:"displayName"        validate:"required"`
	ResourceID        string                 `json:"resourceID,omitempty"`
	IsVirtualResource bool                   `json:"isVirtualResource"`
	Content           map[string]interface{} `json:"content"            validate:"required"`
	Transformer       map[string]interface{} `json:"transformer"        validate:"required"`
	TriggerMode       string                 `json:"triggerMode"        validate:"oneof=manually automate"`
	Config            map[string]interface{} `json:"config"`
}

func NewCreateFlowActionRequest() *CreateFlowActionRequest {
	return &CreateFlowActionRequest{}
}

func (req *CreateFlowActionRequest) ExportTransformerInString() string {
	jsonByte, _ := json.Marshal(req.Transformer)
	return string(jsonByte)
}

func (req *CreateFlowActionRequest) ExportActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.ActionType)
}

func (req *CreateFlowActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *CreateFlowActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *CreateFlowActionRequest) ExportConfigInString() string {
	jsonByte, _ := json.Marshal(req.Config)
	return string(jsonByte)
}

func (req *CreateFlowActionRequest) AppendVirtualResourceToTemplate(value interface{}) {
	req.Content[ACTION_REQUEST_CONTENT_FIELD_VIRTUAL_RESOURCE] = value
}

func (req *CreateFlowActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.ActionType)
}

func (req *CreateFlowActionRequest) IsLocalVirtualAction() bool {
	return resourcelist.IsLocalVirtualResource(req.ActionType)
}

func (req *CreateFlowActionRequest) IsRemoteVirtualAction() bool {
	return resourcelist.IsRemoteVirtualResource(req.ActionType)
}

func (req *CreateFlowActionRequest) NeedFetchResourceInfoFromSourceManager() bool {
	return resourcelist.NeedFetchResourceInfoFromSourceManager(req.ActionType)
}
