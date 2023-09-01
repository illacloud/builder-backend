package request

import "encoding/json"

// the test resource connection request like:
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
type TestResourceConnectionRequest struct {
	ResourceName string                 `json:"resourceName" validate:"required,min=1,max=128"`
	ResourceType string                 `json:"resourceType" validate:"required"`
	Content      map[string]interface{} `json:"content" 	    validate:"required"`
}

func NewTestResourceConnectionRequest() *TestResourceConnectionRequest {
	return &TestResourceConnectionRequest{}
}

func (req *TestResourceConnectionRequest) ExportType() string {
	return req.ResourceType
}

func (req *TestResourceConnectionRequest) ExportOptionsInString() string {
	content, _ := json.Marshal(req.Content)
	return string(content)
}
