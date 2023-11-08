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
	DRIVE_API_LIST                    = "/api/v1/teams/%s/illaAction/files?%s"
	DRIVE_API_GET_DOWNLOAD_SIGNED_URL = "/api/v1/teams/%s/illaAction/files/%s/url"
	DRIVE_API_GET_UPLOAD_ADDRESS      = "/api/v1/teams/%s/illaAction/files"
	DRIVE_API_UPDATE_FILE_STATUS      = "/api/v1/teams/%s/illaAction/files/%s/status"
)

// duplication strategies
const (
	DUPLICATION_STRATEGY_COVER  = "cover"
	DUPLICATION_STRATEGY_RENAME = "rename"
	DUPLICATION_STRATEGY_MANUAL = "manual"
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

func (r *IllaDriveRestAPI) List(teamID int, path string, page int, limit int, fileID string, search string, expirationType string, expiry string, hotlinkProtection bool) (interface{}, error) {
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
	if fileID != "" {
		params += "fileID=" + fileID + "&"
	}
	if search != "" {
		params += "search=" + search + "&"
	}

	// get file list
	// the HTTP request like:
	// ```
	// https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files?path=%2Froot&page=1&limit=20&type=1&search=&sortedBy=lastModifiedAt&sortedType=desc
	// ```
	// and the response like:
	// ```json
	//{
	//    "path": "/root",
	//    "currentFolderID": "ILAfx4p1C7cX",
	//    "files": [
	//        {
	//            "id": "ILAfx4p1C7cW",
	//            "name": "PublicAccess",
	//            "owner": "Anonymous",
	//            "type": "anonymousFolder",
	//            "size": 0,
	//            "contentType": "",
	//            "createdAt": "2023-10-17T11:06:57.001514Z",
	//            "lastModifiedAt": "2023-10-17T11:06:57.001514Z",
	//            "lastModifiedBy": "Anonymous"
	//        },
	//        {
	//            "id": "ILAfx4p1C7al",
	//            "name": "lemmy.hjson",
	//            "owner": "karminski",
	//            "type": "file",
	//            "size": 590,
	//            "contentType": "",
	//            "createdAt": "2023-11-08T10:03:14.82347Z",
	//            "lastModifiedAt": "2023-11-08T10:03:16.395958Z",
	//            "lastModifiedBy": "karminski"
	//        },
	//		...
	// 	   ]
	//	}
	// ```
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

	var listResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &listResponse)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}

	rawFiles, errInExtractRawFiles := ExtractRawFilesFromListResponse(listResponse)
	if errInExtractRawFiles != nil {
		return nil, errInExtractRawFiles
	}
	fileIDs, errInExtractFileIDs := ExtractFileIDFromRawFiles(rawFiles)
	if errInExtractFileIDs != nil {
		return nil, errInExtractFileIDs
	}

	// get file tinyurls
	tinyURLsMap, errInGenerateTinyURLs := r.generateDriveTinyURLs(teamID, fileIDs, expirationType, expiry, hotlinkProtection)
	if errInGenerateTinyURLs != nil {
		return nil, errInGenerateTinyURLs
	}

	rawFilesExtened, errInExtendRawFilesTinyURL := ExtendRawFilesTinyURL(rawFiles, tinyURLsMap)
	if errInExtendRawFilesTinyURL != nil {
		return nil, errInExtendRawFilesTinyURL
	}
	fileList, errInNewFileList := NewFileListByListResponseAndExtendedFiles(listResponse, rawFilesExtened)
	if errInNewFileList != nil {
		return nil, errInNewFileList
	}

	return fileList, nil
}

func (r *IllaDriveRestAPI) GetUploadAddres(teamID int, overwriteDuplicate bool, path string, fileName string, fileSize int64, contentType string) (map[string]interface{}, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_GET_UPLOAD_ADDRES)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	// get folder ID for update
	folderID, errInGetFolderID := r.getFolderIDByPath(teamID, path)
	if errInGetFolderID != nil {
		return nil, errInGetFolderID
	}

	duplicationHandler := DUPLICATION_STRATEGY_RENAME
	if overwriteDuplicate {
		duplicationHandler = DUPLICATION_STRATEGY_COVER
	}

	// init request
	req := NewUploadFileRequestByParam(true, fileName, folderID, "file", fileSize, duplicationHandler, contentType)

	// get file upload address
	// the http request like:
	// ```
	// [POST] https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files
	// {"resumable":true,"name":"lemmy.hjson","folderID":"ILAfx4p1C7cX","type":"file","size":590,"duplicationHandler":"manual","contentType":""}
	// ````
	// and the response like:
	// ```json
	// {
	//     "id": "ILAex4p1C74v",
	//     "name": "putty.exe",
	//     "folderID": "ILAfx4p1C7cX",
	//     "type": "file",
	//     "resumable": true,
	//     "url": "https://storage.googleapis.com/drive_34_49653013-6cdd-11ee-91ee-e25d6f70d4d8/321_0e94bdab-e489-4ae7-91b2-b2a4bf134ee7.exe?X-Goog-Algorithm=GOOG4-RSA-SHA256\u0026X-Goog-Credential=illa-drive-dev%40illa-drive.iam.gserviceaccount.com%2F20231108%2Fauto%2Fstorage%2Fgoog4_request\u0026X-Goog-Date=20231108T144850Z\u0026X-Goog-Expires=1799\u0026X-Goog-Signature=0e63a4350d99e6e99af65820e0072a0dc43a7208c2fb0d65e6389d9e094d596fd6623a53d1d3a9e3323c4acc8433124a425d52867304b373b5ec2de411a036823231a28586dd40615a14334c030a8b9a50a346969cad9cd26e4fb84a0f447b04d9fda12f7becd958993a4ea754248e3ea57047ac20663883c3fea896d19779ac6b028361191c9c60ad88f09cf91f1fb0d1d99363600e476b00805a086b1998e3d9962712241fc3f22f4f8110c0a3bf88a6479a2cabc0494175c5699f0907770864e1976119e72aa13a6d29a63dddd37397161cb9e087e3c2c075cac7d1d755fd46283b7b8fd7d1fa861b95efe2f576559b05989a3d312f2e754b17959dc5cea3\u0026X-Goog-SignedHeaders=content-type%3Bhost%3Bx-goog-resumable"
	// }
	//```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_UPLOAD_ADDRESS, idconvertor.ConvertIntToString(teamID))
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
	var uploadResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &uploadResponse)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}

	return uploadResponse, nil
}

// The request like:

func (r *IllaDriveRestAPI) UpdateFileStatus(teamID int, fileID string, status string) (map[string]interface{}, error) {
	// self-hist need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(teamID, DRIVE_API_ACTION_UPDATE_FILE_STATUS)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	req := NewUpdateFileStatusRequestByParam(status)

	// get tiny url request,
	// the request like:
	// ```
	// [PUT] https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files/ILAfx4p1C7al/status
	// {"status":"complete"}
	// ````
	// and the response like:
	// ```json
	// {
	//     "id": "ILAex4p1C74v",
	//     "name": "putty.exe",
	//     "owner": "karminski",
	//     "type": "file",
	//     "size": 1647912,
	//     "contentType": "application/x-msdownload",
	//     "createdAt": "2023-11-08T14:48:50.941978Z",
	//     "lastModifiedAt": "2023-11-08T14:48:53.725975Z",
	//     "lastModifiedBy": "karminski"
	// }
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_UPDATE_FILE_STATUS, idconvertor.ConvertIntToString(teamID), fileID)
	resp, errInPost := client.R().
		SetHeader("Action-Token", actionToken).
		SetBody(req).
		Put(uri)
	if errInPost != nil {
		return nil, errInPost
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}
	// unmarshal
	var updateStatusResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &updateStatusResponse)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	// get file tiny url prefix
	return updateStatusResponse, nil
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
