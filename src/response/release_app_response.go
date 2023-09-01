package response

import (
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
)

type ReleaseAppResponse struct {
	Version         int `json:"version"`
	ReleaseVersion  int `json:"releaseVersion"`
	MainlineVersion int `json:"mainlineVersion"`
}

func NewReleaseAppResponse(app *model.App) *ReleaseAppResponse {
	fmt.Printf("[DUMP] NewReleaseAppResponse: %+v\n", app.ExportReleaseVersion())
	resp := &ReleaseAppResponse{
		Version:         app.ExportReleaseVersion(),
		ReleaseVersion:  app.ExportReleaseVersion(),
		MainlineVersion: app.ExportMainlineVersion(),
	}
	return resp
}

func (resp *ReleaseAppResponse) ExportForFeedback() interface{} {
	fmt.Printf("[DUMP] ExportForFeedback: %+v\n", resp)

	return resp
}
