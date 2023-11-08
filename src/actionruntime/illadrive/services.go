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

type IllaDriveConnector struct {
	Action IllaDriveTemplate
}

// AI Agent have no validate resource options method
func (r *IllaDriveConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	return common.ValidateResult{Valid: true}, nil
}

func (r *IllaDriveConnector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	fmt.Printf("[DUMP] actionOptions: %+v \n", actionOptions)
	_, errorInNewRequest := resourcemanager.NewRunIllaDriveRequest(actionOptions)
	if errorInNewRequest != nil {
		return common.ValidateResult{Valid: false}, errorInNewRequest
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

func (r *IllaDriveConnector) ListFiles(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.ListFiles(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) GetUploadAddres(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.GetUploadAddres(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) GetMutipleUploadAddres(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.GetMutipleUploadAddres(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) UpdateFileStatus(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.UpdateFileStatus(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) GetDownloadAddress(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.GetDownloadAddress(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) GetMutipleDownloadAddres(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.GetMutipleDownloadAddres(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) DeleteFile(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.DeleteFile(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) DeleteMultipleFile(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.DeleteMultipleFile(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}

func (r *IllaDriveConnector) ModifyFileName(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// call api
	api, errInNewAPI := illadrivesdk.NewIllaDriveRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	runIllaDriveResult, errInRunIllaDrive := api.ModifyFileName(actionOptions)
	fmt.Printf("[DUMP] runIllaDriveResult: %+v\n", runIllaDriveResult)
	fmt.Printf("[DUMP] errInRunIllaDrive: %+v\n", errInRunIllaDrive)

	if errInRunIllaDrive != nil {
		return res, errInRunIllaDrive
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runIllaDriveResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}
