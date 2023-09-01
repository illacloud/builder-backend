package request

import "encoding/json"

// the update resource request like:
//
//	{
//	    "resourceName": "sample",
//	    "resourceType": "postgresql",
//	    "content": {
//	        "host": "111.111.111.111",
//	        "port": "5432",
//	        "databaseName": "dbName",
//	        "databaseUsername": "username",
//	        "databasePassword": "password",
//	        "ssl": {
//	            "ssl": false,
//	            "clientKey": "",
//	            "clientCert": "",
//	            "serverCert": ""
//	        }
//	    }
//	}
type UpdateResourceRequest struct {
	ResourceName string                 `json:"resourceName" validate:"required,min=1,max=128"`
	ResourceType string                 `json:"resourceType" validate:"required"`
	Content      map[string]interface{} `json:"content" 	    validate:"required"`
}

func NewUpdateResourceRequest() *UpdateResourceRequest {
	return &UpdateResourceRequest{}
}

func (req *UpdateResourceRequest) ExportType() string {
	return req.ResourceType
}

func (req *UpdateResourceRequest) ExportOptionsInString() string {
	content, _ := json.Marshal(req.Content)
	return string(content)
}
