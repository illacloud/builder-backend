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
	DRIVE_API_GET_FOLDER_ID_BY_PATH = "/api/v1/teams/%s/illaAction/folder?%s"
)

func (r *IllaDriveRestAPI) getFolderIDByPath(path string) (string, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return "", nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_GET_FOLDER_ID_BY_PATH)
	if errInGenerateToken != nil {
		return "", errInGenerateToken
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
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_FOLDER_ID_BY_PATH, idconvertor.ConvertIntToString(r.TeamID), params)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[DUMP] IllaDriveSDK.GetFolderIDByPath() resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return "", errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return "", errors.New(resp.String())
	}

	var getFolderIDResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &getFolderIDResponse)
	if errInUnMarshal != nil {
		return "", errInUnMarshal
	}

	folderIDRaw, hitFolderID := getFolderIDResponse["id"]
	if !hitFolderID {
		return "", errors.New("invalied response, missing id field")
	}

	folderID, folderIDAssertPass := folderIDRaw.(string)
	if !folderIDAssertPass {
		return "", errors.New("invalied id type, assert failed")
	}
	return folderID, nil
}
