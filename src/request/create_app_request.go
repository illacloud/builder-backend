package request

type CreateAppRequest struct {
	Name       string      `json:"appName" validate:"required"`
	InitScheme interface{} `json:"initScheme"`
	AppType    string      `json:"appType"`
}

func NewCreateAppRequest() *CreateAppRequest {
	return &CreateAppRequest{}
}

func (req *CreateAppRequest) ExportAppName() string {
	return req.Name
}

func (req *CreateAppRequest) ExportInitScheme() interface{} {
	return req.InitScheme
}

func (req *CreateAppRequest) ExportAppType() string {
	return req.AppType
}
