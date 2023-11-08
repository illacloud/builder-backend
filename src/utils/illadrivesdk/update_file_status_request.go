package illadrivesdk

type UpdateFileStatusRequest struct {
	Status string `json:"status"`
}

func NewUpdateFileStatusRequest() *UpdateFileStatusRequest {
	return &UpdateFileStatusRequest{}
}

func NewUpdateFileStatusRequestByParam(status string) *UpdateFileStatusRequest {
	return &UpdateFileStatusRequest{
		Status: status,
	}
}
