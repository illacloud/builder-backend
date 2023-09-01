package response

type IsPublicAppResponse struct {
	IsPublic bool `json:"isPublic"`
}

func NewIsPublicAppResponse(isPublic bool) *IsPublicAppResponse {
	resp := &IsPublicAppResponse{
		IsPublic: isPublic,
	}
	return resp
}

func (resp *IsPublicAppResponse) ExportForFeedback() interface{} {
	return resp
}
