package model

type ReleaseAppResponse struct {
	Version int `json:"version"`
}

func NewReleaseAppResponse(version int) *ReleaseAppResponse {
	resp := &ReleaseAppResponse{
		Version: version,
	}
	return resp
}

func (resp *ReleaseAppResponse) ExportForFeedback() interface{} {
	return resp
}
