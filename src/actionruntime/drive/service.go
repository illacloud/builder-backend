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

type DriveConnector struct {
	Action DriveTemplate
}

// Drive have no validate resource options method
func (r *DriveConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	return common.ValidateResult{Valid: true}, nil
}

func (r *DriveConnector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	fmt.Printf("[DUMP] actionOptions: %+v \n", actionOptions)
	_, errorInNewIllaDriveSDK := illadrivesdk.NewIllaDriveSDK(actionOptions)
	if errorInNewIllaDriveSDK != nil {
		return common.ValidateResult{Valid: false}, errorInNewIllaDriveSDK
	}

	return common.ValidateResult{Valid: true}, nil
}

// Drive have no test connection method
func (r *DriveConnector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: false}, errors.New("unsupported type: Drive")
}

// Drive have no meta info
func (r *DriveConnector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: false}, errors.New("unsupported type: Drive")
}
