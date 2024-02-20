// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package illadrive

import (
	"errors"
	"fmt"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
	illadrivesdk "github.com/illacloud/builder-backend/src/utils/illadrivesdk"
)

const (
	DRIVE_ACTION_OPTIONS_FIELD_OPERATION     = "operation"
	DRIVE_ACTION_OPTIONS_FIELD_TEAM_ID       = "teamID"
	DRIVE_ACTION_OPTIONS_FIELD_USER_ID       = "userID"
	DRIVE_ACTION_OPTIONS_FIELD_INSTANCE_TYPE = "instanceType"
	DRIVE_ACTION_OPTIONS_FIELD_INSTANCE_ID   = "instanceID"
)

type IllaDriveConnector struct {
	Action IllaDriveTemplate
}

// AI Agent have no validate resource options method
func (r *IllaDriveConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	return common.ValidateResult{Valid: true}, nil
}

func (r *IllaDriveConnector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	fmt.Printf("[DUMP] actionOptions: %+v \n", actionOptions)
	// check action options common field
	_, hitOperation := actionOptions[DRIVE_ACTION_OPTIONS_FIELD_OPERATION]
	if !hitOperation {
		return common.ValidateResult{Valid: false}, errors.New("missing operation field")

	}
	return common.ValidateResult{Valid: true}, nil
}

// AI Agent have no test connection method
func (r *IllaDriveConnector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: false}, errors.New("unsupported type: AI Agent")
}

// AI Agent have no meta info
func (r *IllaDriveConnector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: false}, errors.New("unsupported type: AI Agent")
}

func (r *IllaDriveConnector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	fmt.Printf("[DUMP] illadrive.Run() actionOptions: %+v\n", actionOptions)

	// resolve actionOptions
	teamID, _ := resolveIntFieldsFromActionOptions(actionOptions, DRIVE_ACTION_OPTIONS_FIELD_TEAM_ID)
	userID, _ := resolveIntFieldsFromActionOptions(actionOptions, DRIVE_ACTION_OPTIONS_FIELD_USER_ID)
	instanceType, _ := resolveIntFieldsFromActionOptions(actionOptions, DRIVE_ACTION_OPTIONS_FIELD_INSTANCE_TYPE)
	instanceID, _ := resolveIntFieldsFromActionOptions(actionOptions, DRIVE_ACTION_OPTIONS_FIELD_INSTANCE_ID)
	operation, hitOperation := actionOptions[DRIVE_ACTION_OPTIONS_FIELD_OPERATION]
	if !hitOperation {
		return res, errors.New("missing operation field")

	}

	// init
	driveAPI := illadrivesdk.NewIllaDriveRestAPI(teamID, userID, instanceType, instanceID)
	driveAPI.OpenDebug()

	// operation filter for illa drive sdk function call
	switch operation {
	case illadrivesdk.DRIVE_API_ACTION_LIST_FILES:
		// list files require field "path", "page", "limit", "fileID", "search", "sortedBy", "expirationType", "expiry", "hotlinkProtection"
		path, page, limit, fileID, search, sortedBy, sortedType, expirationType, expiry, hotlinkProtection, errInExtractParam := extractListFileOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.ListFiles(path, page, limit, fileID, search, sortedBy, sortedType, expirationType, expiry, hotlinkProtection)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_GET_UPLOAD_ADDRESS:
		// get upload address require field teamID int, overwriteDuplicate bool, path string, fileName string, fileSize int64, contentType string
		overwriteDuplicate, path, fileName, fileSize, contentType, errInExtractParam := extractGetUploadAddressOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.GetUploadAddres(overwriteDuplicate, path, fileName, fileSize, contentType)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_UPDATE_FILE_STATUS:
		// get upload address require field fileID string, status string
		fileID, status, errInExtractParam := extractUpdateFileStatusOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.UpdateFileStatus(fileID, status)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_GET_MULTIPLE_UPLOAD_ADDRESS:
		// get mutiple upload address require field: teamID int, overwriteDuplicate bool, path string, fileNames []string, fileSizes []int64, contentTypes []string
		overwriteDuplicate, path, fileNames, fileSizes, contentTypes, errInExtractParam := extractGetMutipleUploadAddressOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.GetMultipleUploadAddress(overwriteDuplicate, path, fileNames, fileSizes, contentTypes)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = ret
	case illadrivesdk.DRIVE_API_ACTION_GET_DOWNLOAD_ADDRESS:
		// get mutiple upload address require field: teamID int, fileID string
		fileID, errInExtractParam := extractFileIDFromOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.GetDownloadAddress(fileID)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_GET_MULTIPLE_DOWNLOAD_ADDRESS:
		// get mutiple upload address require field: teamID int, fileIDs []string
		fileIDs, errInExtractParam := extractFileIDsFromOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.GetMultipleDownloadAddres(fileIDs)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = ret

	case illadrivesdk.DRIVE_API_ACTION_DELETE_FILE:
		// delete field require field: teamID int, fileID string
		fileID, errInExtractParam := extractFileIDFromOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.DeleteFile(fileID)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_DELETE_MULTIPLE_FILE:
		// delete mutiple file require field: teamID int, fileIDs []string
		fileIDs, errInExtractParam := extractFileIDsFromOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.DeleteMultipleFile(fileIDs)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)
	case illadrivesdk.DRIVE_API_ACTION_RENAME_FILE:
		// / get mutiple upload address require field: teamID int, fileID string
		fileID, fileName, errInExtractParam := extractRenameFileOperationParams(actionOptions)
		if errInExtractParam != nil {
			return res, errInExtractParam
		}
		ret, errInCallAPI := driveAPI.RenameFile(fileID, fileName)
		if errInCallAPI != nil {
			return res, errInCallAPI
		}
		res.Rows = append(res.Rows, ret)

	}

	// feedback
	res.SetSuccess()
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func extractListFileOperationParams(actionOptions map[string]interface{}) (string, int, int, string, string, string, string, string, string, bool, error) {
	path := "/root"
	page := 1
	limit := 20
	fileID := ""     // [OPTIONAL], if fileID given will not use "search" field.
	search := ""     // [OPTIONAL], search for file name contains
	sortedBy := ""   // [OPTIONAL], sort result by field
	sortedType := "" // [OPTIONAL], sort order
	expirationType := "persistent"
	expiry := "300s"
	hotlinkProtection := true
	for key, value := range actionOptions {
		switch key {
		case "path":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field path assert failed")
			}
			path = valueAsserted
		case "page":
			valueAsserted, ValueAssertPass := value.(float64)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field page assert failed")
			}
			page = int(valueAsserted)
		case "limit":
			valueAsserted, ValueAssertPass := value.(float64)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field limit assert failed")
			}
			limit = int(valueAsserted)
		case "fileID":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field fileID assert failed")
			}
			fileID = valueAsserted
		case "search":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field search assert failed")
			}
			search = valueAsserted
		case "sortedBy":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field sortedBy assert failed")
			}
			sortedBy = valueAsserted
		case "sortedType":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field sortedType assert failed")
			}
			sortedType = valueAsserted
		case "expirationType":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field expirationType assert failed")
			}
			expirationType = valueAsserted
		case "expiry":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field expiry assert failed")
			}
			expiry = valueAsserted
		case "hotlinkProtection":
			valueAsserted, ValueAssertPass := value.(bool)
			if !ValueAssertPass {
				return "", 0, 0, "", "", "", "", "", "", false, errors.New("field hotlinkProtection assert failed")
			}
			hotlinkProtection = valueAsserted
		}
	}
	return path, page, limit, fileID, search, sortedBy, sortedType, expirationType, expiry, hotlinkProtection, nil
}

func extractGetUploadAddressOperationParams(actionOptions map[string]interface{}) (bool, string, string, int64, string, error) {
	overwriteDuplicate := false
	path := "/root"
	fileName := ""
	var fileSize int64
	contentType := ""
	for key, value := range actionOptions {
		switch key {
		case "overwriteDuplicate":
			valueAsserted, ValueAssertPass := value.(bool)
			if !ValueAssertPass {
				return false, "", "", 0, "", errors.New("field overwriteDuplicate assert failed")
			}
			overwriteDuplicate = valueAsserted
		case "path":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return false, "", "", 0, "", errors.New("field path assert failed")
			}
			path = valueAsserted
		case "fileName":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return false, "", "", 0, "", errors.New("field fileName assert failed")
			}
			fileName = valueAsserted
		case "fileSize":
			valueAsserted, ValueAssertPass := value.(float64)
			if !ValueAssertPass {
				return false, "", "", 0, "", errors.New("field fileSize assert failed")
			}
			fileSize = int64(valueAsserted)
		case "contentType":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return false, "", "", 0, "", errors.New("field contentType assert failed")
			}
			contentType = valueAsserted
		}
	}
	return overwriteDuplicate, path, fileName, fileSize, contentType, nil

}

func extractUpdateFileStatusOperationParams(actionOptions map[string]interface{}) (string, string, error) {
	fileID := ""
	status := ""
	for key, value := range actionOptions {
		switch key {
		case "fileID":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", "", errors.New("field fileID assert failed")
			}
			fileID = valueAsserted
		case "status":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", "", errors.New("field status assert failed")
			}
			status = valueAsserted
		}
	}
	return fileID, status, nil

}

func extractGetMutipleUploadAddressOperationParams(actionOptions map[string]interface{}) (bool, string, []string, []int64, []string, error) {
	overwriteDuplicate := false
	path := "/root"
	fileNames := make([]string, 0)
	fileSizes := make([]int64, 0)
	contentTypes := make([]string, 0)
	for key, value := range actionOptions {
		switch key {
		case "overwriteDuplicate":
			valueAsserted, ValueAssertPass := value.(bool)
			if !ValueAssertPass {
				return false, "", []string{}, []int64{}, []string{}, errors.New("field overwriteDuplicate assert failed")
			}
			overwriteDuplicate = valueAsserted
		case "path":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return false, "", []string{}, []int64{}, []string{}, errors.New("field path assert failed")
			}
			path = valueAsserted
		case "files":
			filesAsserted, filesAssertPass := value.([]interface{})
			if !filesAssertPass {
				return false, "", []string{}, []int64{}, []string{}, errors.New("field files assert failed")
			}
			for _, file := range filesAsserted {
				fileAsserted, fileAssertPass := file.(map[string]interface{})
				if !fileAssertPass {
					return false, "", []string{}, []int64{}, []string{}, errors.New("field file in files assert failed")
				}
				for subKey, subValue := range fileAsserted {
					switch subKey {
					case "fileName":
						subValueAsserted, subValueAssertPass := subValue.(string)
						if !subValueAssertPass {
							return false, "", []string{}, []int64{}, []string{}, errors.New("field fileName assert failed")
						}
						fileNames = append(fileNames, subValueAsserted)
					case "fileSize":
						subValueAsserted, subValueAssertPass := subValue.(float64)
						if !subValueAssertPass {
							return false, "", []string{}, []int64{}, []string{}, errors.New("field fileSize assert failed")
						}
						fileSizes = append(fileSizes, int64(subValueAsserted))
					case "contentType":
						subValueAsserted, subValueAssertPass := subValue.(string)
						if !subValueAssertPass {
							return false, "", []string{}, []int64{}, []string{}, errors.New("field contentType assert failed")
						}
						contentTypes = append(contentTypes, subValueAsserted)
					}
				}
			}
		}
	}
	return overwriteDuplicate, path, fileNames, fileSizes, contentTypes, nil
}

func extractFileIDFromOperationParams(actionOptions map[string]interface{}) (string, error) {
	fileID := ""
	for key, value := range actionOptions {
		switch key {
		case "fileID":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", errors.New("field fileID assert failed")
			}
			fileID = valueAsserted
		}
	}
	return fileID, nil

}

func extractFileIDsFromOperationParams(actionOptions map[string]interface{}) ([]string, error) {
	fileIDs := make([]string, 0)
	for key, value := range actionOptions {
		switch key {
		case "fileIDs":
			fileIDsAsserted, fileIDsAssertPass := value.([]interface{})
			if !fileIDsAssertPass {
				return []string{}, errors.New("field fileIDs assert failed")
			}
			for _, subValue := range fileIDsAsserted {
				subValueAsserted, subValueAssertPass := subValue.(string)
				if !subValueAssertPass {
					return []string{}, errors.New("field fileIDs values assert failed")
				}
				fileIDs = append(fileIDs, subValueAsserted)
			}
		}
	}
	return fileIDs, nil
}

func extractRenameFileOperationParams(actionOptions map[string]interface{}) (string, string, error) {
	fileID := ""
	fileName := ""
	for key, value := range actionOptions {
		switch key {
		case "fileID":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", "", errors.New("field fileID assert failed")
			}
			fileID = valueAsserted
		case "fileName":
			valueAsserted, ValueAssertPass := value.(string)
			if !ValueAssertPass {
				return "", "", errors.New("field fileName assert failed")
			}
			fileName = valueAsserted
		}
	}
	return fileID, fileName, nil

}
