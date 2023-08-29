package request

import "encoding/json"

type GetAppByIDsInternalRequest struct {
	IDs []int `json:"ids"        validate:"required"`
}

func (req *GetAppByIDsInternalRequest) ExportInJSONString() string {
	jsonByte, _ := json.Marshal(req)
	return string(jsonByte)
}

func NewGetAppByIDsInternalRequest() *GetAppByIDsInternalRequest {
	return &GetAppByIDsInternalRequest{}
}
