package illadrivesdk

type DeleteFileRequest struct {
	IDs []string `json:"ids"`
}

func NewDeleteFileRequest() *DeleteFileRequest {
	return &DeleteFileRequest{}
}

func NewDeleteFileRequestByParam(ids []string) *DeleteFileRequest {
	return &DeleteFileRequest{
		IDs: ids,
	}
}
