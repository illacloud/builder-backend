package request

// The create action HTTP request body like:
// ```json
//
//		{
//			"actions":[
//				{
//		    	"actionType": "postgresql",
//		    	"displayName": "postgresql1",
//		    	"resourceID": "ILAfx4p1C7cd",
//		    	"content": {
//		    	    "mode": "sql",
//		    	    "query": ""
//		    	},
//		    	"isVirtualResource": true,
//		    	"transformer": {
//		    	    "rawData": "",
//		    	    "enable": false
//		    	},
//		    	"triggerMode": "manually",
//		    	"config": {
//		    	    "public": false,
//		    	    "advancedConfig": {
//		    	        "runtime": "none",
//		    	        "pages": [],
//		    	        "delayWhenLoaded": "",
//		    	        "displayLoadingPage": false,
//		    	        "isPeriodically": false,
//		    	        "periodInterval": ""
//		    	    }
//		    	}
//			},
//	 ...
//	 ]
//	}
//
// ```
type CreateActionByBatchRequest struct {
	Actions []*CreateActionRequest `json:"actions"`
}

func NewCreateActionByBatchRequest() *CreateActionByBatchRequest {
	return &CreateActionByBatchRequest{}
}

func (req *CreateActionByBatchRequest) ExportActions() []*CreateActionRequest {
	return req.Actions
}
