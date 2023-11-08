package illadrivesdk

type RenameFileRequest struct {
	name string `json:"name"`
}

func NewRenameFileRequest() *RenameFileRequest {
	return &RenameFileRequest{}
}

func NewRenameFileRequestByParam(name string) *RenameFileRequest {
	return &RenameFileRequest{
		Name: name,
	}
}
