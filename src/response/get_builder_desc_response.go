package response

import (
	"time"
)

type GetBuilderDescResponse struct {
	AppNum         int       `json:"appNum"`
	ResourceNum    int       `json:"resourceNum"`
	ActionNum      int       `json:"actionNum"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}

type EmptyBuilderDescResponse struct {
	AppNum         int    `json:"appNum"`
	ResourceNum    int    `json:"resourceNum"`
	ActionNum      int    `json:"actionNum"`
	LastModifiedAt string `json:"lastModifiedAt"` // is "" by first time enter builder.
}

func NewGetBuilderDescResponse(appNum int, resourceNum int, actionNum int, lastModifiedAt time.Time) *GetBuilderDescResponse {
	return &GetBuilderDescResponse{
		AppNum:         appNum,
		ResourceNum:    resourceNum,
		ActionNum:      actionNum,
		LastModifiedAt: lastModifiedAt,
	}
}

func NewEmptyBuilderDescResponse(appNum int, resourceNum int, actionNum int) *EmptyBuilderDescResponse {
	return &EmptyBuilderDescResponse{
		AppNum:      appNum,
		ResourceNum: resourceNum,
		ActionNum:   actionNum,
	}
}

func (resp *GetBuilderDescResponse) ExportForFeedback() interface{} {
	return resp
}

func (resp *EmptyBuilderDescResponse) ExportForFeedback() interface{} {
	return resp
}
