package illadrivesdk

import (
	"errors"
	"fmt"
)

func ExtractRawFilesFromListResponse(listResponse map[string]interface{}) ([]map[string]interface{}, error) {
	// extract file ids
	listedFiles, hitListedFiles := listResponse["files"]
	if !hitListedFiles {
		return nil, errors.New("mossing files field in list response")
	}

	fmt.Printf("[DUMP] ExtractRawFilesFromListResponse() listedFiles: %+v\n", listedFiles)

	// assert sub structure
	listedFilesAsserted, assertListedFilesPass := listedFiles.([]interface{})
	if !assertListedFilesPass {
		return []map[string]interface{}{}, nil
	}
	files := make([]map[string]interface{}, 0)
	for _, listedFile := range listedFilesAsserted {
		listedFileAsserted, listedFileAssertePass := listedFile.(map[string]interface{})
		if !listedFileAssertePass {
			return nil, errors.New("invalied file list element returned")
		}
		files = append(files, listedFileAsserted)
	}
	return files, nil
}

func ExtractFileIDFromRawFiles(files []map[string]interface{}) ([]string, error) {
	fileIDs := make([]string, 0)
	for _, file := range files {
		fileType, hitFileType := file["type"]
		if !hitFileType {
			return nil, errors.New("missing field type from files data")
		}
		fileTypeString, fileTypeAssertPass := fileType.(string)
		if !fileTypeAssertPass {
			return nil, errors.New("invalied file type data type")
		}
		// we should only download file, not folder
		if fileTypeString != "file" {
			continue
		}
		fileID, hitFileID := file["id"]
		if !hitFileID {
			return nil, errors.New("missing field id from files data")
		}
		fileIDString, fileIDAssertPass := fileID.(string)
		if !fileIDAssertPass {
			return nil, errors.New("invalied file ID data type")
		}
		fileIDs = append(fileIDs, fileIDString)
	}
	return fileIDs, nil
}

func ExtendRawFilesTinyURL(files []map[string]interface{}, tinyURLsMap map[string]string) ([]map[string]interface{}, error) {
	for serial, file := range files {
		fileID, hitFileID := file["id"]
		if !hitFileID {
			return nil, errors.New("missing field id from files data for extend")
		}
		fileIDString, fileIDAssertPass := fileID.(string)
		if !fileIDAssertPass {
			return nil, errors.New("invalied file ID data type")
		}
		tinyURL, hitTinyURL := tinyURLsMap[fileIDString]
		if !hitTinyURL {
			files[serial]["tinyURL"] = ""
		} else {
			files[serial]["tinyURL"] = tinyURL
		}
	}
	return files, nil
}
