package illadrivesdk

import "errors"

type FileList struct {
	Path            string        `json:"path"`
	CurrentFolderID string        `json:"currentFolderID"`
	Files           []interface{} `json:"files"`
}

func NewFileList() *FileList {
	return &FileList{}
}

func NewFileListByListResponse(listResponse map[string]interface{}) (*FileList, error) {
	pathRaw, hitPath := listResponse["path"]
	currentFolderIDRaw, hitCurrentFolderID := listResponse["currentFolderID"]
	if !hitPath || !hitCurrentFolderID {
		return nil, errors.New("invalied list reponse, missing path or currentFolderID field")
	}

	pathAsserted, pathRawAssertPass := pathRaw.(string)
	currentFolderIDAsserted, currentFolderIDRawAssertPass := currentFolderIDRaw.(string)
	if !pathRawAssertPass || !currentFolderIDRawAssertPass {
		return nil, errors.New("invalied list reponse structure, path or currentFolderID field assert failed")
	}

	return &FileList{
		Path:            pathAsserted,
		CurrentFolderID: currentFolderIDAsserted,
		Files:           make([]interface{}, 0),
	}, nil
}
