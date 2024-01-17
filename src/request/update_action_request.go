package request

import (
	"encoding/json"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

// The update action HTTP request body like:
// ```json
//
//	{
//	    "actionID": "ILAex4p1C7rD",
//	    "uid": "781f0ed4-62eb-4615-bd41-80bf2af8ceb4",
//	    "teamID": "ILAfx4p1C7bN",
//	    "resourceID": "ILAfx4p1C7cc",
//	    "displayName": "postgresql1",
//	    "actionType": "postgresql",
//	    "isVirtualResource": false,
//	    "content": {
//	        "mode": "sql",
//	        "query": "select * from data;"
//	    },
//	    "transformer": {
//	        "enable": false,
//	        "rawData": ""
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
//	    },
//	    "createdAt": "2023-08-25T10:18:21.914894943Z",
//	    "createdBy": "ILAfx4p1C7dX",
//	    "updatedAt": "2023-08-25T10:18:21.91489513Z",
//	    "updatedBy": "ILAfx4p1C7dX"
//	}
//
// ```
type UpdateActionRequest struct {
	ActionID          string                 `json:"actionID"         `
	ActionType        string                 `json:"actionType"         validate:"required"`
	DisplayName       string                 `json:"displayName"        validate:"required"`
	ResourceID        string                 `json:"resourceID,omitempty"`
	IsVirtualResource bool                   `json:"isVirtualResource"`
	Content           map[string]interface{} `json:"content"            validate:"required"`
	Transformer       map[string]interface{} `json:"transformer"        validate:"required"`
	TriggerMode       string                 `json:"triggerMode"        validate:"oneof=manually automate"`
	Config            map[string]interface{} `json:"config"`
}

func NewUpdateActionRequest() *UpdateActionRequest {
	return &UpdateActionRequest{}
}

func (req *UpdateActionRequest) ExportTransformerInString() string {
	jsonByte, _ := json.Marshal(req.Transformer)
	return string(jsonByte)
}

func (req *UpdateActionRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}

func (req *UpdateActionRequest) ExportActionIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ActionID)
}

func (req *UpdateActionRequest) ExportActionTypeInInt() int {
	return resourcelist.GetResourceNameMappedID(req.ActionType)
}

func (req *UpdateActionRequest) ExportTemplateInString() string {
	jsonByte, _ := json.Marshal(req.Content)
	return string(jsonByte)
}

func (req *UpdateActionRequest) ExportConfigInString() string {
	jsonByte, _ := json.Marshal(req.Config)
	return string(jsonByte)
}

func (req *UpdateActionRequest) AppendVirtualResourceToTemplate(value interface{}) {
	req.Content[ACTION_REQUEST_CONTENT_FIELD_VIRTUAL_RESOURCE] = value
}

func (req *UpdateActionRequest) IsVirtualAction() bool {
	return resourcelist.IsVirtualResource(req.ActionType)
}

func (req *UpdateActionRequest) IsLocalVirtualAction() bool {
	return resourcelist.IsLocalVirtualResource(req.ActionType)
}

func (req *UpdateActionRequest) IsRemoteVirtualAction() bool {
	return resourcelist.IsRemoteVirtualResource(req.ActionType)
}

func (req *UpdateActionRequest) NeedFetchResourceInfoFromSourceManager() bool {
	return resourcelist.NeedFetchResourceInfoFromSourceManager(req.ActionType)
}
