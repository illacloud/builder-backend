package illadrivesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	DRIVE_API_GENERATE_TINY_URL_BATCH = "/api/v1/teams/%s/illaAction/links/batch"
)

const (
	FILE_TINY_URL_PREFIX_TEST       = "https://cloud-api-test.illacloud.com/drive/f/%s"
	FILE_TINY_URL_PREFIX_BETA       = "https://cloud-api-beta.illacloud.com/drive/f/%s"
	FILE_TINY_URL_PREFIX_PRODUCTION = "https://cloud-api.illacloud.com/drive/f/%s"
)

// the function response are fileID => fileTinyURL lookup map.
func (r *IllaDriveRestAPI) generateDriveTinyURLs(fileIDs []string, expirationType string, expiry string, hotlinkProtection bool) (map[string]string, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_GENERATE_TINY_URLS)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	req := NewGenerateTinyURLRequestByParam(fileIDs, expirationType, expiry, hotlinkProtection)

	// get tiny url request, the request like:
	// ```json
	// {"ids":["ILAfx4p1C7cp"],"expirationType":"persistent","hotlinkProtection":true}
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GENERATE_TINY_URL_BATCH, idconvertor.ConvertIntToString(r.TeamID))
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
	// get file tiny url prefix
	tinyURLPrefix := FILE_TINY_URL_PREFIX_TEST
	if r.Config.IsCloudProductionMode() {
		tinyURLPrefix = FILE_TINY_URL_PREFIX_PRODUCTION
	} else if r.Config.IsCloudBetaMode() {
		tinyURLPrefix = FILE_TINY_URL_PREFIX_BETA
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
		tinyURLMaps[fileIDAsserted] = fmt.Sprintf(tinyURLPrefix, tinyURLStringAsserted)
	}
	return tinyURLMaps, nil
}
