package request

type DuplicateAppRequest struct {
	Name string `json:"appName" validate:"required"`
}

func NewDuplicateAppRequest() *DuplicateAppRequest {
	return &DuplicateAppRequest{}
}

func (req *DuplicateAppRequest) ExportAppName() string {
	return req.Name
}
