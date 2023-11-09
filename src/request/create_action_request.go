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
type CreateActionRequest struct {
	ActionType        string                 `json:"actionType"         validate:"required"`
	DisplayName       string                 `json:"displayName"        validate:"required"`
	ResourceID        string                 `json:"resourceID,omitempty"`
	IsVirtualResource bool                   `json:"isVirtualResource"`
	Content           map[string]interface{} `json:"content"            validate:"required"`
	Transformer       map[string]interface{} `json:"transformer"        validate:"required"`
	TriggerMode       string                 `json:"triggerMode"        validate:"oneof=manually automate"`
	Config            map[string]interface{} `json:"config"`
}

func NewCreateActionRequest() *CreateActionRequest {
	return &CreateActionRequest{}
}

func (req *CreateActionRequest) ExportTransformerInString() string {
	jsonByte, _ := json.Marshal(req.Transformer)
	return string(jsonByte)
}

func (req *CreateActionRequest) ExportActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.ActionType)
}

func (req *CreateActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *CreateActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *CreateActionRequest) ExportConfigInString() string {
	jsonByte, _ := json.Marshal(req.Config)
	return string(jsonByte)
}

func (req *CreateActionRequest) AppendVirtualResourceToTemplate(value interface{}) {
	req.Content[ACTION_REQUEST_CONTENT_FIELD_VIRTUAL_RESOURCE] = value
}

func (req *CreateActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.ActionType)
}

func (req *CreateActionRequest) IsLocalVirtualAction() bool {
	return resourcelist.IsLocalVirtualResource(req.ActionType)
}

func (req *CreateActionRequest) IsRemoteVirtualAction() bool {
	return resourcelist.IsRemoteVirtualResource(req.ActionType)
}

func (req *CreateActionRequest) NeedFetchResourceInfoFromSourceManager() bool {
	return resourcelist.NeedFetchResourceInfoFromSourceManager(req.ActionType)
}
