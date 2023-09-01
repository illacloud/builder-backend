package request

import "encoding/json"

// the create resource request like:
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
type CreateResourceRequest struct {
	ResourceName string                 `json:"resourceName" validate:"required,min=1,max=128"`
	ResourceType string                 `json:"resourceType" validate:"required"`
	Content      map[string]interface{} `json:"content" 	    validate:"required"`
}

func NewCreateResourceRequest() *CreateResourceRequest {
	return &CreateResourceRequest{}
}

func (req *CreateResourceRequest) ExportType() string {
	return req.ResourceType
}

func (req *CreateResourceRequest) ExportOptionsInString() string {
	content, _ := json.Marshal(req.Content)
	return string(content)
}
