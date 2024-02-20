package illadrivesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	DRIVE_API_LIST_FILES              = "/api/v1/teams/%s/illaAction/files?%s"
	DRIVE_API_GET_UPLOAD_ADDRESS      = "/api/v1/teams/%s/illaAction/files"
	DRIVE_API_UPDATE_FILE_STATUS      = "/api/v1/teams/%s/illaAction/files/%s/status"
	DRIVE_API_GET_DOWNLOAD_SIGNED_URL = "/api/v1/teams/%s/illaAction/files/%s/url"
	DRIVE_API_DELETE_FILES            = "/api/v1/teams/%s/illaAction/files"
	DRIVE_API_RENAME_FILE             = "/api/v1/teams/%s/illaAction/files/%s/name"
)

// duplication strategies
const (
	DUPLICATION_STRATEGY_COVER  = "cover"
	DUPLICATION_STRATEGY_RENAME = "rename"
	DUPLICATION_STRATEGY_MANUAL = "manual"
)

type IllaDriveRestAPI struct {
	Config       *config.Config
	TeamID       int  `json:"teamID"`
	UserID       int  `json:"userID"`
	InstanceType int  `json:"instanceType"`
	InstanceID   int  `json:"instanceID"`
	Debug        bool `json:"-"`
}

func NewIllaDriveRestAPI(teamID int, userID int, instanceType int, instanceID int) *IllaDriveRestAPI {
	return &IllaDriveRestAPI{
		Config:       config.GetInstance(),
		TeamID:       teamID,
		UserID:       userID,
		InstanceType: instanceType,
		InstanceID:   instanceID,
	}
}

func (r *IllaDriveRestAPI) CloseDebug() {
	r.Debug = false
}

func (r *IllaDriveRestAPI) OpenDebug() {
	r.Debug = true
}

func (r *IllaDriveRestAPI) GenerateAccessJWTToken(usage string) (map[string]interface{}, error) {
	token, errInGenerateToken := GenerateDriveAPIActionToken(r, usage)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}
	return map[string]interface{}{
		"token": token,
	}, nil
}

func (r *IllaDriveRestAPI) ListFiles(path string, page int, limit int, fileID string, search string, sortedBy string, sortedType string, expirationType string, expiry string, hotlinkProtection bool) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_LIST_FILES)
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
	if sortedBy != "" {
		params += "sortedBy=" + sortedBy + "&"
	}
	if sortedType != "" {
		params += "sortedType=" + sortedType + "&"
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
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_LIST_FILES, idconvertor.ConvertIntToString(r.TeamID), params)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.ListFiles()  uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.ListFiles()  response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[DUMP] IllaDriveSDK.ListFiles()  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return nil, errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, errors.New(resp.String())
	}

	fmt.Printf("[DUMP] ListFiles raw response: %+v\n", resp.String())

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

	// no files hit
	if len(fileIDs) == 0 {
		fileList, errInNewFileList := NewFileListByListResponseAndExtendedFiles(listResponse, nil)
		if errInNewFileList != nil {
			return nil, errInNewFileList
		}
		return structs.Map(fileList), nil
	}

	// get file tinyurls
	tinyURLsMap, errInGenerateTinyURLs := r.generateDriveTinyURLs(fileIDs, expirationType, expiry, hotlinkProtection)
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

	return structs.Map(fileList), nil
}

func (r *IllaDriveRestAPI) GetUploadAddres(overwriteDuplicate bool, path string, fileName string, fileSize int64, contentType string) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_GET_UPLOAD_ADDRESS)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	// get folder ID for update
	folderID, errInGetFolderID := r.getFolderIDByPath(path)
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
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_UPLOAD_ADDRESS, idconvertor.ConvertIntToString(r.TeamID))
	resp, errInPost := client.R().
		SetHeader("Action-Token", actionToken).
		SetBody(req).
		Post(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.GetUploadAddres()  uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.GetUploadAddres()  response: %+v, err: %+v \n", resp, errInPost)
		log.Printf("[DUMP] IllaDriveSDK.GetUploadAddres()  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInPost != nil {
		return nil, errInPost
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}
	fmt.Printf("[DUMP] GetUploadAddres raw response: %+v\n", resp.String())

	// unmarshal
	var uploadResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &uploadResponse)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}

	return uploadResponse, nil
}

// The request like:

func (r *IllaDriveRestAPI) UpdateFileStatus(fileID string, status string) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}

	// calculate token
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_UPDATE_FILE_STATUS)
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
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_UPDATE_FILE_STATUS, idconvertor.ConvertIntToString(r.TeamID), fileID)
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
	fmt.Printf("[DUMP] UpdateFileStatus raw response: %+v\n", resp.String())

	// unmarshal
	var updateStatusResponse map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &updateStatusResponse)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	// get file tiny url prefix
	return updateStatusResponse, nil
}

func (r *IllaDriveRestAPI) GetMultipleUploadAddress(overwriteDuplicate bool, path string, fileNames []string, fileSizes []int64, contentTypes []string) ([]map[string]interface{}, error) {
	ret := make([]map[string]interface{}, 0)
	fmt.Printf("[DUMP] GetMultipleUploadAddress() fileName: %+v, fileSizes: %+v, contentTypes: %+v\n ", fileNames, fileSizes, contentTypes)
	for serial, fileName := range fileNames {
		uploadAddressInfo, errInGetUploadAddress := r.GetUploadAddres(overwriteDuplicate, path, fileName, fileSizes[serial], contentTypes[serial])
		fmt.Printf("[DUMP] uploadAddressInfo[%d]: %+v\n", serial, uploadAddressInfo)
		if errInGetUploadAddress != nil {
			return nil, errInGetUploadAddress
		}
		ret = append(ret, uploadAddressInfo)
	}
	return ret, nil
}

func (r *IllaDriveRestAPI) GetDownloadAddress(fileID string) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_GET_DOWNLOAD_ADDRESS)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}
	// get file download link, the request like:
	// ```
	// [GET] https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files/ILAfx4p1C7cp/url
	// ```
	// and response like:
	// ```json
	// {
	//     "id": "ILAfx4p1C7cp",
	//     "name": "007Y7SRMly1gvn2ydkciqj611a0m77a802.jpg",
	//     "contentType": "image/jpeg",
	//     "size": 103564,
	//     "downloadURL": "https://storage.googleapis.com/drive_34_49653013-6cdd-11ee-91ee-e25d6f70d4d8/175_b9fa8e7a-2586-4974-a195-1175ff86a68d.jpg?X-Goog-Algorithm=GOOG4-RSA-SHA256\u0026X-Goog-Credential=illa-drive-dev%40illa-drive.iam.gserviceaccount.com%2F20231108%2Fauto%2Fstorage%2Fgoog4_request\u0026X-Goog-Date=20231108T184657Z\u0026X-Goog-Expires=179\u0026X-Goog-Signature=1aaaddc7f389e2d2c179661835276ee1f71ed7d65119216978050f999034c13c251f3480e71651ea77ac34ef851298c15ef5118224deb8d74f58ec61bb15cf8df769b66f61dbdb446a480122dd257e0333ea16939648a9fb9ce39d4646dd549df1776613763aa6dadc01bb7bc5c695cac6588b7a3fdbfafac50a85a5bf47f5176de420269812f2fa04d8eab712a2e5434913deb2bdf3cd9b22061eea759a4a57a864e0be46e316cb0908d885889c2d7ebad8a8f5b2506b54a0d40a6322293de3391a66aaa004ff883894d7bfccc6b6982ed561e31e7a0445a2960a4672260d966d0acbf304ce4f0f7458271ae60b36b34bbda66d0f379fb52b87ecd26a97445c\u0026X-Goog-SignedHeaders=host",
	//     "createdAt": "2023-11-02T13:15:49.053259Z",
	//     "lastModifiedAt": "2023-11-02T13:15:50.571428Z"
	// }
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_GET_DOWNLOAD_SIGNED_URL, idconvertor.ConvertIntToString(r.TeamID), fileID)
	resp, errInGet := client.R().
		SetHeader("Action-Token", actionToken).
		Get(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.GetDownloadAddress()  uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.GetDownloadAddress()  response: %+v, err: %+v \n", resp, errInGet)
		log.Printf("[DUMP] IllaDriveSDK.GetDownloadAddress()  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInGet != nil {
		return nil, errInGet
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, errors.New(resp.String())
	}
	fmt.Printf("[DUMP] GetDownloadAddress raw response: %+v\n", resp.String())

	var downloadAddress map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(resp.String()), &downloadAddress)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	return downloadAddress, nil
}

func (r *IllaDriveRestAPI) GetMultipleDownloadAddres(fileIDs []string) ([]map[string]interface{}, error) {
	ret := make([]map[string]interface{}, 0)
	for _, fileID := range fileIDs {
		fileDownloadAddressInfo, errInGetDownloadAddress := r.GetDownloadAddress(fileID)
		if errInGetDownloadAddress != nil {
			return nil, errInGetDownloadAddress
		}
		ret = append(ret, fileDownloadAddressInfo)
	}
	return ret, nil
}

func (r *IllaDriveRestAPI) DeleteFile(fileID string) (map[string]interface{}, error) {
	return r.DeleteMultipleFile([]string{fileID})
}

func (r *IllaDriveRestAPI) DeleteMultipleFile(fileIDs []string) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_DELETE_FILES)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}
	// init request body
	req := NewDeleteFileRequestByParam(fileIDs)

	// delete file, the request like:
	// ```
	// [DELETE] https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files
	// {"ids":["ILAfx4p1C7cp"]}
	// ```
	// and response like:
	// ```
	// HTTP 200
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_DELETE_FILES, idconvertor.ConvertIntToString(r.TeamID))
	resp, errInDelete := client.R().
		SetHeader("Action-Token", actionToken).
		SetBody(req).
		Delete(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.DeleteMultipleFile()  uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.DeleteMultipleFile()  response: %+v, err: %+v \n", resp, errInDelete)
		log.Printf("[DUMP] IllaDriveSDK.DeleteMultipleFile()  resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInDelete != nil {
		return nil, errInDelete
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}
	fmt.Printf("[DUMP] DeleteMultipleFile raw response: %+v\n", resp.String())

	return map[string]interface{}{"deleted": true}, nil
}

func (r *IllaDriveRestAPI) RenameFile(fileID string, fileName string) (map[string]interface{}, error) {
	// self-host need skip this method.
	if !r.Config.IsCloudMode() {
		return nil, nil
	}
	actionToken, errInGenerateToken := GenerateDriveAPIActionToken(r, DRIVE_API_ACTION_RENAME_FILE)
	if errInGenerateToken != nil {
		return nil, errInGenerateToken
	}

	// init request body
	req := NewRenameFileRequestByParam(fileName)

	// rename file, the request like:
	// ```
	// [PUT] https://cloud-api-test.illacloud.com/drive/api/v1/teams/ILAfx4p1C7ey/files/ILAex4p1C74v/name
	// {"name":"222.exe"}
	// ```
	// and response like:
	// ```
	// HTTP 200
	// ```
	client := resty.New()
	uri := r.Config.GetIllaDriveAPIForSDK() + fmt.Sprintf(DRIVE_API_RENAME_FILE, idconvertor.ConvertIntToString(r.TeamID), fileID)
	resp, errInPut := client.R().
		SetHeader("Action-Token", actionToken).
		SetBody(req).
		Put(uri)
	if r.Debug {
		log.Printf("[DUMP] IllaDriveSDK.RenameFile() uri: %+v \n", uri)
		log.Printf("[DUMP] IllaDriveSDK.RenameFile() response: %+v, err: %+v \n", resp, errInPut)
		log.Printf("[DUMP] IllaDriveSDK.RenameFile() resp.StatusCode(): %+v \n", resp.StatusCode())
	}
	if errInPut != nil {
		return nil, errInPut
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}
	fmt.Printf("[DUMP] RenameFile raw response: %+v\n", resp.String())

	return map[string]interface{}{"renamed": true}, nil
}
