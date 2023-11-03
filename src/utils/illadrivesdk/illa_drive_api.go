package illadrivesdk

import (
	"errors"

	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

type IllaDriveRestAPI struct {
	Config *config.Config
	Debug  bool `json:"-"`
}

func NewIllaDriveRestAPI() (*IllaDriveRestAPI, error) {
	return &IllaDriveRestAPI{
		Config: config.GetInstance(),
	}, nil
}

func (r *IllaDriveRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaDriveRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaDriveRestAPI) GetResource(resourceType int, resourceID int) (map[string]interface{}, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}
	switch resourceType {
	case resourcelist.TYPE_AI_AGENT_ID:
		return r.GetAIAgent(resourceID)
	default:
		return nil, errors.New("Invalied resource type: " + resourcelist.GetResourceIDMappedType(resourceType))
	}
}

func (r *IllaDriveRestAPI) List(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) Read(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetUploadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetMutipleUploadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetDownloadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetMutipleDownloadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) DeleteFile(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) DeleteMultipleFile(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) UpdateFileProperty(driveID int) (map[string]interface{}, error) {

}
