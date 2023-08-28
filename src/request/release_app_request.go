package request

type ReleaseAppRequest struct {
	Public bool `json:"public" validate:"required"`
}

func NewReleaseAppRequest() *ReleaseAppRequest {
	return &ReleaseAppRequest{}
}

func (req *ReleaseAppRequest) ExportPublic() bool {
	return req.Public
}
