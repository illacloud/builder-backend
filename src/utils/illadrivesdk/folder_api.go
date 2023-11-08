package illadrivesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	DRIVE_API_GET_DOLDER_ID_BY_PATH = "/api/v1/teams/%s/illaAction/folder?%s"
)

func (r *IllaDriveRestAPI) GetFolderIDByPath(teamID int, path string) (int, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return 0, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_GET_DOLDER_ID_BY_PATH)
	if errInGenerateToken != nil {
		return 0, errInGenerateToken
	}

	// concat request param
	params := ""
	if path != "" {
		params += "path=" + path + "&"
	} else {
		path = "/root"
	}

	// get file list
	// the http get response should be like:
	// ```json
	// {"id":"ILAfx4p1C7dB"}
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_DOLDER_ID_BY_PATH, idconvertor.ConvertIntToString(teamID), params)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return 0, errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return 0, errors.New(resp.String())
	}

	var getFolderIDResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &getFolderIDResponse)
	if errInUnMarshal != nil {
		return 0, errInUnMarshal
	}

	folderIDRaw, hitFolderID := getFolderIDResponse["id"]
	if !hitFolderID {
		return 0, errors.New("invalied response, missing id field")
	}

	folderID, folderIDAssertPass := folderIDRaw.(int)
	if !folderIDAssertPass {
		return 0, errors.New("invalied id type, assert failed")
	}
	return folderID, nil
}
