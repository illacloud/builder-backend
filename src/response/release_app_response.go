package response

import "fmt"

type ReleaseAppResponse struct {
	Version int `json:"version"`
}

func NewReleaseAppResponse(version int) *ReleaseAppResponse {
	fmt.Printf("[DUMP] NewReleaseAppResponse: %+v\n", version)
	resp := &ReleaseAppResponse{
		Version: version,
	}
	return resp
}

func (resp *ReleaseAppResponse) ExportForFeedback() interface{} {
	fmt.Printf("[DUMP] ExportForFeedback: %+v\n", resp)

	return resp
}
