package model

type CreateAppRequest struct {
	Name       string        `json:"appName" validate:"required"`
	InitScheme []interface{} `json:"initScheme"`
}

func NewCreateAppRequest() *CreateAppRequest {
	return &CreateAppRequest{}
}

func (req *CreateAppRequest) ExportName() string {
	return req.Name
}

func (req *CreateAppRequest) ExportInitScheme() []interface{} {
	return req.InitScheme
}

func (req *CreateAppRequest) IsRequestWithScheme() bool {
	if len(req.InitScheme) > 0 {
		return true
	}
	return false
}
