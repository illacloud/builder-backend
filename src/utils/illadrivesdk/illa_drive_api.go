package illadrivesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	DRIVE_API_LIST                    = "/api/v1/teams/%s/files?%s"
	DRIVE_API_GET_DOWNLOAD_SIGNED_URL = "/api/v1/teams/%s/files/%s/url"
	DRIVE_API_GENERATE_TINY_URL_BATCH = "/api/v1/teams/%s/links/batch"
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

func (r *IllaDriveRestAPI) GenerateAccessJWTToken(teamID int, usage string) (map[string]interface{}, error) {
	token, errInGenerateToken := GenerateDriveAPIActionToken(teamID, usage)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}
	return map[string]interface{}{
		"token": token,
	}, nil
}

// the request like:
// ```json
// {"ids":["ILAfx4p1C7cp"],"expirationType":"persistent","hotlinkProtection":true}
// ```
func (r *IllaDriveRestAPI) generateDriveTinyURLs(teamID int, fileIDs []string, expirationType string, expiry string, hotlinkProtection bool) (map[string]string, error) {
	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_GENERATE_TINY_URLS)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	req := NewGenerateTinyURLRequestByParam(fileIDs, expirationType, expiry, hotlinkProtection)

	// get file list
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GENERATE_TINY_URL_BATCH, idconvertor.ConvertIntToString(teamID))
	resp, errInPost := client.R().
		SetHeader("Action-Token", actionToken).
		SetBody(req).
		Post(uri)
	if errInPost != nil {
		return nil, errInPost
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}
	// unmarshal
	var tinyURLs []interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &tinyURLs)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	// fill into lookup table
	tinyURLMaps := make(map[string]string)
	for _, tinyURL := range tinyURLs {
		tinyURLAsserted, assertPass := tinyURL.(map[string]interface{})
		if !assertPass {
			return nil, errors.New("assert failed in tiny url struct")
		}
		fileID, hitFileID := tinyURLAsserted["fileID"]
		tinyURL, hitTinyURL := tinyURLAsserted["tinyURL"]
		if !hitFileID || !hitTinyURL {
			return nil, errors.New("tiny url response struct missing field")

		}
		fileIDAsserted, fileIDAssertPass := fileID.(string)
		tinyURLStringAsserted, tinyURLAssertPass := tinyURL.(string)
		if !fileIDAssertPass || !tinyURLAssertPass {
			return nil, errors.New("assert failed in tiny url struct field")
		}
		tinyURLMaps[fileIDAsserted] = tinyURLStringAsserted
	}
	return tinyURLMaps, nil
}

func (r *IllaDriveRestAPI) List(teamID int, path string, page int, limit int, search string) (map[string]interface{}, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_LIST)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	// concat request param
	params := ""
	if path != "" {
		params += "path=" + path + "&"
	}
	if page != 0 {
		params += "page=" + strconv.Itoa(page) + "&"
	}
	if limit != 0 {
		params += "limit=" + strconv.Itoa(limit) + "&"
	}
	if search != "" {
		params += "search=" + search + "&"
	}

	// get file list
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_LIST, idconvertor.ConvertIntToString(teamID), params)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  uri: %+v \n", uri)
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return nil, errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, errors.New(resp.String())
	}

	var driveFiles map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &driveFiles)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}

	// get file

	return driveFiles, nil
}

func (r *IllaDriveRestAPI) ReadFileProperty(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetUploadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetMutipleUploadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) GetDownloadAddres(teamID int, fileID string) (map[string]interface{}, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_GET_DOWNLOAD_ADDRES)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}
	// run
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_DOWNLOAD_SIGNED_URL, idconvertor.ConvertIntToString(teamID), fileID)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  uri: %+v \n", uri)
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[IllaResourceManagerRestAPI.GetAiAgent()]  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return nil, errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, errors.New(resp.String())
	}

	var downloadAddress map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &downloadAddress)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	return downloadAddress, nil
}

func (r *IllaDriveRestAPI) GetMutipleDownloadAddres(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) DeleteFile(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) DeleteMultipleFile(driveID int) (map[string]interface{}, error) {

}

func (r *IllaDriveRestAPI) UpdateFileProperty(driveID int) (map[string]interface{}, error) {

}
